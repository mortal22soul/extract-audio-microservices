"""Analytics service business logic"""
import asyncio
import logging
import os
import tempfile
from typing import Dict, Any, List, Optional
from datetime import datetime

from .database import mongodb, VideoAnalyticsModel
from .messaging import publisher
from .models import video_quality_model
from .quality_analyzer import quality_analyzer, quality_scorer
from .video_processor import metadata_extractor, thumbnail_generator, video_indexer
from .config import config

logger = logging.getLogger(__name__)


class AnalyticsService:
    """Main analytics service class"""
    
    def __init__(self):
        self.video_analytics = VideoAnalyticsModel(mongodb)
        self.processing_semaphore = asyncio.Semaphore(config.MAX_CONCURRENT_JOBS)
    
    async def handle_video_analysis(self, message_data: Dict[str, Any]) -> None:
        """Handle complete video analysis request"""
        video_id = message_data.get("video_id")
        user_id = message_data.get("user_id")
        video_path = message_data.get("video_path")
        
        if not all([video_id, user_id, video_path]):
            logger.error("Missing required fields in video analysis message")
            return
        
        async with self.processing_semaphore:
            try:
                logger.info(f"Starting video analysis for video_id: {video_id}")
                
                # Perform all analysis tasks
                metadata = await metadata_extractor.extract_metadata(video_path)
                thumbnails = await thumbnail_generator.generate_thumbnails(video_path, video_id)
                quality_metrics = await self._analyze_quality(video_path)
                
                # Default safety score since moderation is removed
                safety_score = {"overall_score": 1.0, "is_safe": True}
                
                # Calculate weighted quality score
                weighted_quality = quality_scorer.calculate_weighted_score(
                    quality_metrics, safety_score, metadata
                )
                
                tags = await self._extract_tags(metadata, quality_metrics)
                
                # Create search index
                search_index = video_indexer.create_search_index(metadata)
                
                # Store analysis results
                analysis_data = {
                    "video_id": video_id,
                    "user_id": user_id,
                    "metadata": metadata,
                    "thumbnails": thumbnails,
                    "quality": quality_metrics,
                    "weighted_quality": weighted_quality,
                    "safety": safety_score,
                    "tags": tags,
                    "search_index": search_index,
                    "status": "completed"
                }
                
                await self.video_analytics.create_analysis(analysis_data)
                # Publish completion message
                await publisher.publish_analysis_complete(video_id, user_id, analysis_data)
                
                logger.info(f"Completed video analysis for video_id: {video_id}")
                
            except Exception as e:
                logger.error(f"Video analysis failed for video_id {video_id}: {e}")
                await publisher.publish_analysis_error(video_id, user_id, str(e))
    
    async def handle_quality_analysis(self, message_data: Dict[str, Any]) -> None:
        """Handle quality analysis request"""
        video_id = message_data.get("video_id")
        video_path = message_data.get("video_path")
        
        if not all([video_id, video_path]):
            logger.error("Missing required fields in quality analysis message")
            return
        
        try:
            quality_metrics = await self._analyze_quality(video_path)
            
            # Update existing analysis or create new one
            existing = await self.video_analytics.get_analysis(video_id)
            if existing:
                await self.video_analytics.update_analysis(video_id, {"quality": quality_metrics})
            
            logger.info(f"Completed quality analysis for video_id: {video_id}")
            
        except Exception as e:
            logger.error(f"Quality analysis failed for video_id {video_id}: {e}")
    
    async def handle_thumbnail_generation(self, message_data: Dict[str, Any]) -> None:
        """Handle thumbnail generation request"""
        video_id = message_data.get("video_id")
        video_path = message_data.get("video_path")
        
        if not all([video_id, video_path]):
            logger.error("Missing required fields in thumbnail generation message")
            return
        
        try:
            thumbnails = await thumbnail_generator.generate_thumbnails(video_path, video_id)
            
            # Update existing analysis
            existing = await self.video_analytics.get_analysis(video_id)
            if existing:
                await self.video_analytics.update_analysis(video_id, {"thumbnails": thumbnails})
            
            logger.info(f"Generated thumbnails for video_id: {video_id}")
            
        except Exception as e:
            logger.error(f"Thumbnail generation failed for video_id {video_id}: {e}")
    
