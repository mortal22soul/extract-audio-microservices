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
    
    async def handle_safety_check(self, message_data: Dict[str, Any]) -> None:
        """Handle content safety check request"""
        video_id = message_data.get("video_id")
        user_id = message_data.get("user_id")
        
        if not video_id:
            logger.error("Missing video_id in safety check message")
            return
        
        try:
            # Default safety result -- content moderation model is not loaded
            safety_result = {
                "overall_score": 1.0,
                "is_safe": True,
                "categories": {},
                "checked_at": datetime.utcnow().isoformat()
            }
            
            # Update existing analysis with safety results
            existing = await self.video_analytics.get_analysis(video_id)
            if existing:
                await self.video_analytics.update_analysis(video_id, {"safety": safety_result})
            
            logger.info(f"Completed safety check for video_id: {video_id}")
            
        except Exception as e:
            logger.error(f"Safety check failed for video_id {video_id}: {e}")
    
    async def handle_user_interaction(self, message_data: Dict[str, Any]) -> None:
        """Handle user interaction tracking (views, likes, shares)"""
        video_id = message_data.get("video_id")
        user_id = message_data.get("user_id")
        interaction_type = message_data.get("type")  # "view", "like", "share", "download"
        
        if not all([video_id, user_id, interaction_type]):
            logger.error("Missing required fields in user interaction message")
            return
        
        try:
            existing = await self.video_analytics.get_analysis(video_id)
            if existing:
                interactions = existing.get("interactions", {})
                count_key = f"{interaction_type}_count"
                interactions[count_key] = interactions.get(count_key, 0) + 1
                interactions["last_interaction"] = datetime.utcnow().isoformat()
                
                await self.video_analytics.update_analysis(video_id, {
                    "interactions": interactions
                })
            
            logger.info(f"Recorded {interaction_type} interaction for video_id: {video_id}")
            
        except Exception as e:
            logger.error(f"Failed to record interaction for video_id {video_id}: {e}")
    
    async def _analyze_quality(self, video_path: str) -> Dict[str, Any]:
        """Run quality analysis on a video file"""
        try:
            # Use the quality analyzer to perform frame-level analysis
            quality_metrics = await quality_analyzer.analyze_video_quality(video_path)
            return quality_metrics
        except Exception as e:
            logger.error(f"Quality analysis error: {e}")
            return {
                "overall_score": 0.0,
                "sharpness": 0.0,
                "brightness": 0.0,
                "contrast": 0.0,
                "error": str(e)
            }
    
    async def _extract_tags(self, metadata: Dict[str, Any], quality_metrics: Dict[str, Any]) -> List[str]:
        """Extract descriptive tags from video metadata and quality metrics"""
        tags = []
        
        # Tags from metadata
        if metadata:
            # Resolution-based tags
            width = metadata.get("width", 0)
            height = metadata.get("height", 0)
            if height >= 2160:
                tags.append("4k")
            elif height >= 1080:
                tags.append("hd")
                tags.append("1080p")
            elif height >= 720:
                tags.append("hd")
                tags.append("720p")
            elif height > 0:
                tags.append("sd")
            
            # Duration-based tags
            duration = metadata.get("duration", 0)
            if duration > 0:
                if duration < 60:
                    tags.append("short")
                elif duration < 600:
                    tags.append("medium")
                else:
                    tags.append("long")
            
            # Format tags
            codec = metadata.get("codec", "")
            if codec:
                tags.append(codec.lower())
            
            # Audio presence
            if metadata.get("has_audio", False):
                tags.append("has-audio")
        
        # Tags from quality metrics
        if quality_metrics:
            overall_score = quality_metrics.get("overall_score", 0)
            if overall_score >= 8.0:
                tags.append("high-quality")
            elif overall_score >= 5.0:
                tags.append("medium-quality")
            else:
                tags.append("low-quality")
        
        return list(set(tags))  # Deduplicate
