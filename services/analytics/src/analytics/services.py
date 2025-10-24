"""Analytics service business logic"""
import asyncio
import logging
import os
import tempfile
from typing import Dict, Any, List, Optional
from datetime import datetime

from .database import mongodb, VideoAnalyticsModel, UserPreferencesModel, VideoFeaturesModel
from .messaging import publisher
from .models import content_safety_model, video_quality_model, recommendation_model
from .quality_analyzer import quality_analyzer, content_moderator, quality_scorer
from .video_processor import metadata_extractor, thumbnail_generator, video_indexer
from .config import config

logger = logging.getLogger(__name__)


class AnalyticsService:
    """Main analytics service class"""
    
    def __init__(self):
        self.video_analytics = VideoAnalyticsModel(mongodb)
        self.user_preferences = UserPreferencesModel(mongodb)
        self.video_features = VideoFeaturesModel(mongodb)
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
                safety_score = await self._check_content_safety(video_path, metadata)
                
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
                
                # Store features for recommendations
                features = self._extract_features(analysis_data)
                await self.video_features.store_features(video_id, user_id, features)
                
                # Update user preferences
                await self._update_user_preferences(user_id, tags, quality_metrics)
                
                # Handle content moderation workflow
                await self._handle_moderation_workflow(video_id, user_id, safety_score, weighted_quality)
                
                # Trigger recommendation model update (async, don't wait)
                asyncio.create_task(self._update_recommendation_models())
                
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
        video_path = message_data.get("video_path")
        
        if not all([video_id, video_path]):
            logger.error("Missing required fields in safety check message")
            return
        
        try:
            # Extract metadata for safety analysis
            metadata = await metadata_extractor.extract_metadata(video_path)
            safety_score = await self._check_content_safety(video_path, metadata)
            
            # Update existing analysis
            existing = await self.video_analytics.get_analysis(video_id)
            if existing:
                await self.video_analytics.update_analysis(video_id, {"safety": safety_score})
            
            logger.info(f"Completed safety check for video_id: {video_id}")
            
        except Exception as e:
            logger.error(f"Safety check failed for video_id {video_id}: {e}")
    

    
    async def _analyze_quality(self, video_path: str) -> Dict[str, Any]:
        """Analyze video quality metrics using comprehensive quality analyzer"""
        try:
            # Use the new comprehensive quality analyzer
            quality_metrics = await quality_analyzer.analyze_quality(video_path)
            
            logger.info(f"Quality analysis completed for {video_path}")
            return quality_metrics
            
        except Exception as e:
            logger.error(f"Quality analysis failed: {e}")
            return {
                "sharpness_score": 0.0,
                "brightness_score": 0.0,
                "contrast_score": 0.0,
                "noise_score": 0.0,
                "overall_score": 0.0,
                "quality_category": "unknown",
                "resolution_category": "unknown",
                "error": str(e)
            }
    
    async def _check_content_safety(self, video_path: str, metadata: Dict[str, Any]) -> Dict[str, Any]:
        """Check content safety using comprehensive content moderation"""
        try:
            # Use the new comprehensive content moderation analyzer
            safety_result = await content_moderator.analyze_content_safety(video_path, metadata)
            
            logger.info(f"Content safety analysis completed for {video_path}")
            return safety_result
            
        except Exception as e:
            logger.error(f"Content safety check failed: {e}")
            return {
                "overall_score": 0.5,  # Neutral score on error
                "is_safe": True,
                "confidence": 0.0,
                "flags": [],
                "analysis_methods": [],
                "moderation_action": "review",
                "error": str(e)
            }
    
    async def _extract_tags(self, metadata: Dict[str, Any], quality_metrics: Dict[str, Any]) -> List[str]:
        """Extract tags from video metadata and analysis"""
        tags = []
        
        try:
            # Resolution-based tags
            resolution = metadata.get("resolution", "")
            if "1920x1080" in resolution or "1080" in resolution:
                tags.append("hd")
                tags.append("1080p")
            elif "1280x720" in resolution or "720" in resolution:
                tags.append("hd")
                tags.append("720p")
            elif "3840x2160" in resolution or "2160" in resolution:
                tags.append("4k")
                tags.append("ultra-hd")
            
            # Duration-based tags
            duration = metadata.get("duration", 0)
            if duration < 60:
                tags.append("short")
            elif duration < 600:
                tags.append("medium")
            else:
                tags.append("long")
            
            # Quality-based tags
            overall_quality = quality_metrics.get("overall_score", 0)
            if overall_quality > 0.8:
                tags.append("high-quality")
            elif overall_quality > 0.6:
                tags.append("good-quality")
            elif overall_quality > 0.4:
                tags.append("medium-quality")
            else:
                tags.append("low-quality")
            
            # Aspect ratio tags
            aspect_ratio = metadata.get("aspect_ratio", 1.0)
            if abs(aspect_ratio - 16/9) < 0.1:
                tags.append("widescreen")
            elif abs(aspect_ratio - 4/3) < 0.1:
                tags.append("standard")
            elif aspect_ratio > 2:
                tags.append("ultra-wide")
            
            # File size tags
            file_size = metadata.get("file_size", 0)
            if file_size > 1024 * 1024 * 1024:  # > 1GB
                tags.append("large-file")
            elif file_size < 10 * 1024 * 1024:  # < 10MB
                tags.append("small-file")
            
        except Exception as e:
            logger.error(f"Tag extraction failed: {e}")
        
        return list(set(tags))  # Remove duplicates
    
    def _extract_features(self, analysis_data: Dict[str, Any]) -> Dict[str, Any]:
        """Extract features for recommendation system"""
        features = {}
        
        try:
            # Quality features
            quality = analysis_data.get("quality", {})
            features["quality_score"] = quality.get("overall_score", 0)
            features["sharpness"] = quality.get("sharpness_score", 0)
            features["brightness"] = quality.get("brightness_score", 0)
            features["contrast"] = quality.get("contrast_score", 0)
            
            # Metadata features
            metadata = analysis_data.get("metadata", {})
            features["duration"] = metadata.get("duration", 0)
            features["resolution_width"] = metadata.get("width", 0)
            features["resolution_height"] = metadata.get("height", 0)
            features["aspect_ratio"] = metadata.get("aspect_ratio", 1.0)
            features["fps"] = metadata.get("fps", 0)
            
            # Safety features
            safety = analysis_data.get("safety", {})
            features["safety_score"] = safety.get("overall_score", 0.5)
            
            # Tag features (one-hot encoding)
            tags = analysis_data.get("tags", [])
            for tag in tags:
                features[f"tag_{tag}"] = 1
            
        except Exception as e:
            logger.error(f"Feature extraction failed: {e}")
        
        return features
    
    async def _update_user_preferences(self, user_id: str, tags: List[str], quality_metrics: Dict[str, Any]) -> None:
        """Update user preferences based on video analysis"""
        try:
            # Get existing preferences
            preferences = await self.user_preferences.get_preferences(user_id)
            
            if not preferences:
                preferences = {
                    "user_id": user_id,
                    "preferred_tags": {},
                    "quality_preference": 0.5,
                    "interaction_count": 0
                }
            
            # Update tag preferences
            tag_prefs = preferences.get("preferred_tags", {})
            for tag in tags:
                tag_prefs[tag] = tag_prefs.get(tag, 0) + 1
            
            # Update quality preference (moving average)
            current_quality = preferences.get("quality_preference", 0.5)
            new_quality = quality_metrics.get("overall_score", 0.5)
            interaction_count = preferences.get("interaction_count", 0)
            
            # Weighted average with more weight on recent interactions
            alpha = 0.1  # Learning rate
            updated_quality = current_quality * (1 - alpha) + new_quality * alpha
            
            # Update preferences
            preferences.update({
                "preferred_tags": tag_prefs,
                "quality_preference": updated_quality,
                "interaction_count": interaction_count + 1
            })
            
            await self.user_preferences.update_preferences(user_id, preferences)
            
        except Exception as e:
            logger.error(f"Failed to update user preferences: {e}")
    
    async def get_recommendations(self, user_id: str, limit: int = 10) -> List[Dict[str, Any]]:
        """Get video recommendations for a user"""
        try:
            # Get user preferences
            preferences = await self.user_preferences.get_preferences(user_id)
            
            if not preferences:
                # Return popular videos for new users
                return await self._get_popular_videos(limit)
            
            # Get user's preferred tags
            preferred_tags = preferences.get("preferred_tags", {})
            top_tags = sorted(preferred_tags.items(), key=lambda x: x[1], reverse=True)[:5]
            tag_list = [tag for tag, _ in top_tags]
            
            if not tag_list:
                return await self._get_popular_videos(limit)
            
            # Get user's videos to exclude
            user_videos = await self.video_analytics.get_user_videos(user_id, 100)
            exclude_video_ids = [video["video_id"] for video in user_videos]
            
            # Find similar videos based on tags
            recommendations = []
            for tag in tag_list:
                similar_videos = await self.video_analytics.get_similar_videos(
                    [tag], "", limit * 2
                )
                
                for video in similar_videos:
                    if video["video_id"] not in exclude_video_ids:
                        recommendations.append({
                            "video_id": video["video_id"],
                            "similarity_score": self._calculate_similarity_score(video, preferences),
                            "tags": video.get("tags", []),
                            "quality_score": video.get("quality", {}).get("overall_score", 0),
                            "thumbnails": video.get("thumbnails", [])
                        })
            
            # Sort by similarity score and return top results
            recommendations.sort(key=lambda x: x["similarity_score"], reverse=True)
            return recommendations[:limit]
            
        except Exception as e:
            logger.error(f"Failed to get recommendations: {e}")
            return []
    
    async def _get_popular_videos(self, limit: int) -> List[Dict[str, Any]]:
        """Get popular videos based on quality scores"""
        try:
            # This is a simplified implementation
            # In a real system, you might track view counts, likes, etc.
            collection = mongodb.get_collection("video_analytics")
            
            cursor = collection.find({
                "quality.overall_score": {"$gte": 0.6}
            }).sort("quality.overall_score", -1).limit(limit)
            
            videos = await cursor.to_list(length=limit)
            
            recommendations = []
            for video in videos:
                recommendations.append({
                    "video_id": video["video_id"],
                    "similarity_score": video.get("quality", {}).get("overall_score", 0),
                    "tags": video.get("tags", []),
                    "quality_score": video.get("quality", {}).get("overall_score", 0),
                    "thumbnails": video.get("thumbnails", [])
                })
            
            return recommendations
            
        except Exception as e:
            logger.error(f"Failed to get popular videos: {e}")
            return []
    
    def _calculate_similarity_score(self, video: Dict[str, Any], preferences: Dict[str, Any]) -> float:
        """Calculate similarity score between video and user preferences"""
        try:
            score = 0.0
            
            # Tag similarity
            video_tags = set(video.get("tags", []))
            preferred_tags = preferences.get("preferred_tags", {})
            
            tag_score = 0.0
            for tag in video_tags:
                if tag in preferred_tags:
                    tag_score += preferred_tags[tag]
            
            if preferred_tags:
                tag_score /= sum(preferred_tags.values())
            
            # Quality similarity
            video_quality = video.get("quality", {}).get("overall_score", 0)
            preferred_quality = preferences.get("quality_preference", 0.5)
            quality_score = 1.0 - abs(video_quality - preferred_quality)
            
            # Weighted combination
            score = 0.7 * tag_score + 0.3 * quality_score
            
            return min(1.0, max(0.0, score))
            
        except Exception as e:
            logger.error(f"Failed to calculate similarity score: {e}")
            return 0.0
    
    async def _handle_moderation_workflow(
        self,
        video_id: str,
        user_id: str,
        safety_score: Dict[str, Any],
        weighted_quality: Dict[str, Any]
    ) -> None:
        """Handle content moderation workflow based on analysis results"""
        try:
            moderation_action = safety_score.get("moderation_action", "review")
            overall_safety = safety_score.get("overall_score", 0.5)
            quality_score = weighted_quality.get("overall_score", 0.0)
            
            # Create moderation record
            moderation_data = {
                "video_id": video_id,
                "user_id": user_id,
                "safety_score": overall_safety,
                "quality_score": quality_score,
                "moderation_action": moderation_action,
                "flags": safety_score.get("flags", []),
                "confidence": safety_score.get("confidence", 0.0),
                "analysis_methods": safety_score.get("analysis_methods", []),
                "timestamp": datetime.utcnow(),
                "status": "pending"
            }
            
            # Store moderation record
            await self._store_moderation_record(moderation_data)
            
            # Take appropriate action based on moderation result
            if moderation_action == "block":
                await self._block_content(video_id, user_id, "Content blocked due to safety concerns")
                logger.warning(f"Content blocked for video_id: {video_id}")
                
            elif moderation_action == "review":
                await self._flag_for_review(video_id, user_id, "Content flagged for manual review")
                logger.info(f"Content flagged for review: {video_id}")
                
            elif moderation_action == "flag":
                await self._flag_content(video_id, user_id, "Content flagged with warnings")
                logger.info(f"Content flagged with warnings: {video_id}")
                
            else:  # approve
                await self._approve_content(video_id, user_id)
                logger.info(f"Content approved: {video_id}")
                
        except Exception as e:
            logger.error(f"Moderation workflow failed for video_id {video_id}: {e}")
    
    async def _store_moderation_record(self, moderation_data: Dict[str, Any]) -> None:
        """Store moderation record in database"""
        try:
            collection = mongodb.get_collection("content_moderation")
            await collection.insert_one(moderation_data)
            
        except Exception as e:
            logger.error(f"Failed to store moderation record: {e}")
    
    async def _block_content(self, video_id: str, user_id: str, reason: str) -> None:
        """Block content and notify relevant services"""
        try:
            # Update video status to blocked
            await self.video_analytics.update_analysis(video_id, {
                "status": "blocked",
                "block_reason": reason,
                "blocked_at": datetime.utcnow()
            })
            
            # Publish blocking notification
            await publisher.publish_content_blocked(video_id, user_id, reason)
            
        except Exception as e:
            logger.error(f"Failed to block content {video_id}: {e}")
    
    async def _flag_for_review(self, video_id: str, user_id: str, reason: str) -> None:
        """Flag content for manual review"""
        try:
            # Update video status to under review
            await self.video_analytics.update_analysis(video_id, {
                "status": "under_review",
                "review_reason": reason,
                "flagged_at": datetime.utcnow()
            })
            
            # Publish review notification
            await publisher.publish_content_flagged_for_review(video_id, user_id, reason)
            
        except Exception as e:
            logger.error(f"Failed to flag content for review {video_id}: {e}")
    
    async def _flag_content(self, video_id: str, user_id: str, reason: str) -> None:
        """Flag content with warnings but allow it to proceed"""
        try:
            # Update video with warning flags
            await self.video_analytics.update_analysis(video_id, {
                "status": "flagged",
                "warning_reason": reason,
                "flagged_at": datetime.utcnow()
            })
            
            # Publish warning notification
            await publisher.publish_content_warning(video_id, user_id, reason)
            
        except Exception as e:
            logger.error(f"Failed to flag content {video_id}: {e}")
    
    async def _approve_content(self, video_id: str, user_id: str) -> None:
        """Approve content for normal processing"""
        try:
            # Update video status to approved
            await self.video_analytics.update_analysis(video_id, {
                "status": "approved",
                "approved_at": datetime.utcnow()
            })
            
            # Publish approval notification
            await publisher.publish_content_approved(video_id, user_id)
            
        except Exception as e:
            logger.error(f"Failed to approve content {video_id}: {e}")
    
    async def _update_recommendation_models(self) -> None:
        """Update recommendation models with new data (non-blocking)"""
        try:
            # Import here to avoid circular imports
            from .recommendation_engine import recommendation_engine
            
            # Check if recommendation engine is initialized
            if recommendation_engine is None:
                return
            
            # Check if enough time has passed since last training
            current_time = datetime.utcnow()
            if (recommendation_engine.last_training_time and 
                current_time - recommendation_engine.last_training_time < recommendation_engine.training_interval):
                return
            
            # Train models in background
            await recommendation_engine.train_models()
            logger.info("Recommendation models updated successfully")
            
        except Exception as e:
            logger.error(f"Failed to update recommendation models: {e}")
    
    async def handle_user_interaction(self, message_data: Dict[str, Any]) -> None:
        """Handle user interaction tracking for recommendations"""
        user_id = message_data.get("user_id")
        video_id = message_data.get("video_id")
        interaction_type = message_data.get("interaction_type", "view")
        
        if not all([user_id, video_id]):
            logger.error("Missing required fields in user interaction message")
            return
        
        try:
            # Import here to avoid circular imports
            from .recommendation_engine import recommendation_engine
            
            # Check if recommendation engine is initialized
            if recommendation_engine is None:
                logger.warning("Recommendation engine not initialized, skipping interaction tracking")
                return
            
            # Update user interaction
            await recommendation_engine.update_user_interaction(
                user_id, video_id, interaction_type
            )
            
            logger.info(f"Tracked user interaction: {user_id} -> {video_id} ({interaction_type})")
            
        except Exception as e:
            logger.error(f"Failed to handle user interaction: {e}")