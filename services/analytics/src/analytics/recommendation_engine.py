"""
Recommendation Engine for Video Content
Implements content-based and collaborative filtering
"""
import asyncio
import logging
import numpy as np
from typing import Dict, Any, List, Optional, Tuple
from datetime import datetime, timedelta
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity
from sklearn.decomposition import TruncatedSVD
from sklearn.preprocessing import StandardScaler
import redis.asyncio as redis

from .database import mongodb, VideoAnalyticsModel, UserPreferencesModel, VideoFeaturesModel
from .config import config

logger = logging.getLogger(__name__)


class ContentBasedRecommender:
    """Content-based recommendation system using video features"""
    
    def __init__(self):
        self.vectorizer = TfidfVectorizer(
            max_features=1000,
            stop_words='english',
            ngram_range=(1, 2),
            min_df=2
        )
        self.scaler = StandardScaler()
        self.feature_matrix = None
        self.video_ids = []
        self.text_features = None
        self.numerical_features = None
        self.is_trained = False
    
    async def build_feature_matrix(self, video_data: List[Dict[str, Any]]) -> None:
        """Build feature matrix from video data"""
        if not video_data:
            logger.warning("No video data provided for feature matrix")
            return
        
        try:
            # Extract features
            text_features = []
            numerical_features = []
            self.video_ids = []
            
            for video in video_data:
                # Text features (tags, metadata)
                tags = video.get('tags', [])
                title = video.get('metadata', {}).get('title', '')
                description = video.get('metadata', {}).get('description', '')
                
                combined_text = ' '.join(tags) + ' ' + title + ' ' + description
                text_features.append(combined_text)
                
                # Numerical features
                quality = video.get('quality', {})
                metadata = video.get('metadata', {})
                safety = video.get('safety', {})
                
                numerical_feature = [
                    quality.get('overall_score', 0),
                    quality.get('sharpness_score', 0),
                    quality.get('brightness_score', 0),
                    quality.get('contrast_score', 0),
                    metadata.get('duration', 0) / 3600,  # Normalize to hours
                    metadata.get('width', 0) / 1920,  # Normalize to 1080p
                    metadata.get('height', 0) / 1080,
                    metadata.get('fps', 0) / 60,  # Normalize to 60fps
                    safety.get('overall_score', 0.5),
                ]
                
                numerical_features.append(numerical_feature)
                self.video_ids.append(video['video_id'])
            
            # Build text feature matrix
            if text_features:
                self.text_features = self.vectorizer.fit_transform(text_features)
            
            # Build numerical feature matrix
            if numerical_features:
                numerical_array = np.array(numerical_features)
                self.numerical_features = self.scaler.fit_transform(numerical_array)
            
            # Combine features
            if self.text_features is not None and self.numerical_features is not None:
                # Convert sparse matrix to dense for concatenation
                text_dense = self.text_features.toarray()
                self.feature_matrix = np.hstack([text_dense, self.numerical_features])
            elif self.text_features is not None:
                self.feature_matrix = self.text_features.toarray()
            elif self.numerical_features is not None:
                self.feature_matrix = self.numerical_features
            
            self.is_trained = True
            logger.info(f"Built feature matrix with {len(self.video_ids)} videos")
            
        except Exception as e:
            logger.error(f"Failed to build feature matrix: {e}")
            raise
    
    def get_content_recommendations(
        self, 
        video_id: str, 
        n_recommendations: int = 10
    ) -> List[Tuple[str, float]]:
        """Get content-based recommendations for a video"""
        if not self.is_trained or self.feature_matrix is None:
            return []
        
        try:
            if video_id not in self.video_ids:
                return []
            
            # Find video index
            video_idx = self.video_ids.index(video_id)
            
            # Calculate similarity scores
            video_features = self.feature_matrix[video_idx].reshape(1, -1)
            similarity_scores = cosine_similarity(video_features, self.feature_matrix).flatten()
            
            # Get top similar videos (excluding the input video)
            similar_indices = similarity_scores.argsort()[::-1][1:n_recommendations+1]
            
            recommendations = []
            for idx in similar_indices:
                if similarity_scores[idx] > 0.1:  # Minimum similarity threshold
                    recommendations.append((
                        self.video_ids[idx],
                        float(similarity_scores[idx])
                    ))
            
            return recommendations
            
        except Exception as e:
            logger.error(f"Content recommendation failed: {e}")
            return []
    
    def get_user_content_recommendations(
        self,
        user_preferences: Dict[str, Any],
        exclude_video_ids: List[str],
        n_recommendations: int = 10
    ) -> List[Tuple[str, float]]:
        """Get content-based recommendations based on user preferences"""
        if not self.is_trained or self.feature_matrix is None:
            return []
        
        try:
            # Create user preference vector
            preferred_tags = user_preferences.get('preferred_tags', {})
            quality_preference = user_preferences.get('quality_preference', 0.5)
            
            # Build user text profile
            user_text = ' '.join([tag for tag, count in preferred_tags.items() if count > 0])
            
            if user_text:
                user_text_features = self.vectorizer.transform([user_text]).toarray()
            else:
                user_text_features = np.zeros((1, self.text_features.shape[1]))
            
            # Build user numerical profile
            user_numerical = np.array([[
                quality_preference,  # overall_score preference
                quality_preference,  # sharpness preference
                quality_preference,  # brightness preference
                quality_preference,  # contrast preference
                0.5,  # duration (neutral)
                0.5,  # width (neutral)
                0.5,  # height (neutral)
                0.5,  # fps (neutral)
                0.8,  # safety preference (high)
            ]])
            
            if self.numerical_features is not None:
                user_numerical_scaled = self.scaler.transform(user_numerical)
            else:
                user_numerical_scaled = user_numerical
            
            # Combine user features
            if self.text_features is not None and self.numerical_features is not None:
                user_features = np.hstack([user_text_features, user_numerical_scaled])
            elif self.text_features is not None:
                user_features = user_text_features
            elif self.numerical_features is not None:
                user_features = user_numerical_scaled
            else:
                return []
            
            # Calculate similarity scores
            similarity_scores = cosine_similarity(user_features, self.feature_matrix).flatten()
            
            # Get top similar videos (excluding user's videos)
            recommendations = []
            for idx in similarity_scores.argsort()[::-1]:
                video_id = self.video_ids[idx]
                if video_id not in exclude_video_ids and similarity_scores[idx] > 0.1:
                    recommendations.append((video_id, float(similarity_scores[idx])))
                    if len(recommendations) >= n_recommendations:
                        break
            
            return recommendations
            
        except Exception as e:
            logger.error(f"User content recommendation failed: {e}")
            return []


class CollaborativeFilteringRecommender:
    """Collaborative filtering recommendation system"""
    
    def __init__(self, n_components: int = 50):
        self.n_components = n_components
        self.svd = TruncatedSVD(n_components=n_components, random_state=42)
        self.user_item_matrix = None
        self.user_ids = []
        self.video_ids = []
        self.is_trained = False
    
    async def build_user_item_matrix(self, interaction_data: List[Dict[str, Any]]) -> None:
        """Build user-item interaction matrix"""
        if not interaction_data:
            logger.warning("No interaction data provided for collaborative filtering")
            return
        
        try:
            # Create user-item mapping
            user_set = set()
            video_set = set()
            
            for interaction in interaction_data:
                user_set.add(interaction['user_id'])
                video_set.add(interaction['video_id'])
            
            self.user_ids = sorted(list(user_set))
            self.video_ids = sorted(list(video_set))
            
            # Create user-item matrix
            n_users = len(self.user_ids)
            n_videos = len(self.video_ids)
            
            self.user_item_matrix = np.zeros((n_users, n_videos))
            
            user_id_to_idx = {user_id: idx for idx, user_id in enumerate(self.user_ids)}
            video_id_to_idx = {video_id: idx for idx, video_id in enumerate(self.video_ids)}
            
            # Fill matrix with interaction scores
            for interaction in interaction_data:
                user_idx = user_id_to_idx[interaction['user_id']]
                video_idx = video_id_to_idx[interaction['video_id']]
                
                # Weight different interaction types
                interaction_type = interaction.get('interaction_type', 'view')
                weight = {
                    'view': 1.0,
                    'download': 2.0,
                    'like': 3.0,
                    'share': 2.5
                }.get(interaction_type, 1.0)
                
                self.user_item_matrix[user_idx, video_idx] += weight
            
            # Apply SVD for dimensionality reduction
            if np.sum(self.user_item_matrix) > 0:
                # Adjust n_components based on data size
                max_components = min(n_users, n_videos) - 1
                if max_components > 0:
                    self.svd.n_components = min(self.n_components, max_components)
                    self.svd.fit(self.user_item_matrix)
                    self.is_trained = True
                    logger.info(f"Built user-item matrix: {n_users} users, {n_videos} videos, {self.svd.n_components} components")
                else:
                    logger.warning("Not enough data for SVD decomposition")
            
        except Exception as e:
            logger.error(f"Failed to build user-item matrix: {e}")
            raise
    
    def get_collaborative_recommendations(
        self,
        user_id: str,
        exclude_video_ids: List[str],
        n_recommendations: int = 10
    ) -> List[Tuple[str, float]]:
        """Get collaborative filtering recommendations for a user"""
        if not self.is_trained or user_id not in self.user_ids:
            return []
        
        try:
            user_idx = self.user_ids.index(user_id)
            
            # Get user's latent factors
            user_factors = self.svd.transform(self.user_item_matrix[user_idx:user_idx+1])
            
            # Get all video factors
            video_factors = self.svd.components_.T
            
            # Calculate predicted ratings
            predicted_ratings = np.dot(user_factors, self.svd.components_).flatten()
            
            # Get recommendations
            recommendations = []
            for idx in predicted_ratings.argsort()[::-1]:
                video_id = self.video_ids[idx]
                if (video_id not in exclude_video_ids and 
                    self.user_item_matrix[user_idx, idx] == 0):  # Not already interacted
                    
                    score = predicted_ratings[idx]
                    if score > 0:
                        recommendations.append((video_id, float(score)))
                        if len(recommendations) >= n_recommendations:
                            break
            
            return recommendations
            
        except Exception as e:
            logger.error(f"Collaborative filtering failed: {e}")
            return []
    
    def get_similar_users(self, user_id: str, n_users: int = 10) -> List[Tuple[str, float]]:
        """Find users with similar preferences"""
        if not self.is_trained or user_id not in self.user_ids:
            return []
        
        try:
            user_idx = self.user_ids.index(user_id)
            user_vector = self.user_item_matrix[user_idx].reshape(1, -1)
            
            # Calculate user similarities
            similarities = cosine_similarity(user_vector, self.user_item_matrix).flatten()
            
            # Get most similar users (excluding self)
            similar_users = []
            for idx in similarities.argsort()[::-1][1:n_users+1]:
                if similarities[idx] > 0.1:
                    similar_users.append((self.user_ids[idx], float(similarities[idx])))
            
            return similar_users
            
        except Exception as e:
            logger.error(f"Similar users calculation failed: {e}")
            return []


class RecommendationCache:
    """Redis-based caching for recommendations"""
    
    def __init__(self):
        self.redis_client = None
        self.cache_ttl = 3600  # 1 hour
    
    async def connect(self):
        """Connect to Redis"""
        try:
            self.redis_client = redis.from_url(config.REDIS_URL)
            await self.redis_client.ping()
            logger.info("Connected to Redis for recommendation caching")
        except Exception as e:
            logger.error(f"Failed to connect to Redis: {e}")
            self.redis_client = None
    
    async def get_cached_recommendations(
        self, 
        cache_key: str
    ) -> Optional[List[Dict[str, Any]]]:
        """Get cached recommendations"""
        if not self.redis_client:
            return None
        
        try:
            cached_data = await self.redis_client.get(cache_key)
            if cached_data:
                import json
                return json.loads(cached_data)
        except Exception as e:
            logger.error(f"Failed to get cached recommendations: {e}")
        
        return None
    
    async def cache_recommendations(
        self,
        cache_key: str,
        recommendations: List[Dict[str, Any]]
    ) -> None:
        """Cache recommendations"""
        if not self.redis_client:
            return
        
        try:
            import json
            cached_data = json.dumps(recommendations, default=str)
            await self.redis_client.setex(cache_key, self.cache_ttl, cached_data)
        except Exception as e:
            logger.error(f"Failed to cache recommendations: {e}")
    
    async def invalidate_user_cache(self, user_id: str) -> None:
        """Invalidate all cached recommendations for a user"""
        if not self.redis_client:
            return
        
        try:
            pattern = f"recommendations:user:{user_id}:*"
            keys = await self.redis_client.keys(pattern)
            if keys:
                await self.redis_client.delete(*keys)
        except Exception as e:
            logger.error(f"Failed to invalidate user cache: {e}")


class HybridRecommendationEngine:
    """Hybrid recommendation engine combining content-based and collaborative filtering"""
    
    def __init__(self):
        self.content_recommender = ContentBasedRecommender()
        self.collaborative_recommender = CollaborativeFilteringRecommender()
        self.cache = RecommendationCache()
        self.video_analytics = VideoAnalyticsModel(mongodb)
        self.user_preferences = UserPreferencesModel(mongodb)
        self.video_features = VideoFeaturesModel(mongodb)
        
        # Weights for hybrid approach
        self.content_weight = 0.6
        self.collaborative_weight = 0.4
        
        self.last_training_time = None
        self.training_interval = timedelta(hours=6)  # Retrain every 6 hours
    
    async def initialize(self):
        """Initialize the recommendation engine"""
        await self.cache.connect()
        await self.train_models()
    
    async def train_models(self, force_retrain: bool = False) -> None:
        """Train both recommendation models"""
        current_time = datetime.utcnow()
        
        if (not force_retrain and 
            self.last_training_time and 
            current_time - self.last_training_time < self.training_interval):
            return
        
        try:
            logger.info("Training recommendation models...")
            
            # Get video data for content-based filtering
            video_collection = mongodb.get_collection("video_analytics")
            video_cursor = video_collection.find({
                "status": {"$in": ["completed", "approved"]}
            }).limit(10000)
            
            video_data = await video_cursor.to_list(length=10000)
            
            if video_data:
                await self.content_recommender.build_feature_matrix(video_data)
            
            # Get interaction data for collaborative filtering
            interaction_data = await self.user_preferences.get_all_interactions()
            
            if interaction_data:
                await self.collaborative_recommender.build_user_item_matrix(interaction_data)
            
            self.last_training_time = current_time
            logger.info("Recommendation models trained successfully")
            
        except Exception as e:
            logger.error(f"Failed to train recommendation models: {e}")
            raise
    
    async def get_recommendations(
        self,
        user_id: str,
        n_recommendations: int = 10,
        use_cache: bool = True
    ) -> List[Dict[str, Any]]:
        """Get hybrid recommendations for a user"""
        cache_key = f"recommendations:user:{user_id}:hybrid:{n_recommendations}"
        
        # Check cache first
        if use_cache:
            cached_recommendations = await self.cache.get_cached_recommendations(cache_key)
            if cached_recommendations:
                return cached_recommendations
        
        try:
            # Get user preferences and videos
            user_preferences = await self.user_preferences.get_preferences(user_id)
            user_videos = await self.video_analytics.get_user_videos(user_id, 100)
            exclude_video_ids = [video["video_id"] for video in user_videos]
            
            # Get content-based recommendations
            content_recommendations = []
            if user_preferences:
                content_recommendations = self.content_recommender.get_user_content_recommendations(
                    user_preferences, exclude_video_ids, n_recommendations * 2
                )
            
            # Get collaborative filtering recommendations
            collaborative_recommendations = self.collaborative_recommender.get_collaborative_recommendations(
                user_id, exclude_video_ids, n_recommendations * 2
            )
            
            # Combine recommendations with hybrid scoring
            combined_scores = {}
            
            # Add content-based scores
            for video_id, score in content_recommendations:
                combined_scores[video_id] = combined_scores.get(video_id, 0) + (
                    score * self.content_weight
                )
            
            # Add collaborative filtering scores
            for video_id, score in collaborative_recommendations:
                combined_scores[video_id] = combined_scores.get(video_id, 0) + (
                    score * self.collaborative_weight
                )
            
            # Sort by combined score
            sorted_recommendations = sorted(
                combined_scores.items(), 
                key=lambda x: x[1], 
                reverse=True
            )[:n_recommendations]
            
            # Enrich with video metadata
            recommendations = []
            for video_id, score in sorted_recommendations:
                video_data = await self.video_analytics.get_analysis(video_id)
                if video_data:
                    recommendations.append({
                        "video_id": video_id,
                        "score": float(score),
                        "title": video_data.get("metadata", {}).get("title", ""),
                        "tags": video_data.get("tags", []),
                        "quality_score": video_data.get("quality", {}).get("overall_score", 0),
                        "thumbnails": video_data.get("thumbnails", []),
                        "duration": video_data.get("metadata", {}).get("duration", 0),
                        "user_id": video_data.get("user_id"),
                        "created_at": video_data.get("created_at")
                    })
            
            # Fallback to popular videos if no recommendations
            if not recommendations:
                recommendations = await self._get_popular_videos(n_recommendations)
            
            # Cache recommendations
            if use_cache:
                await self.cache.cache_recommendations(cache_key, recommendations)
            
            return recommendations
            
        except Exception as e:
            logger.error(f"Failed to get hybrid recommendations: {e}")
            return await self._get_popular_videos(n_recommendations)
    
    async def get_similar_videos(
        self,
        video_id: str,
        n_recommendations: int = 10,
        use_cache: bool = True
    ) -> List[Dict[str, Any]]:
        """Get videos similar to a specific video"""
        cache_key = f"recommendations:video:{video_id}:similar:{n_recommendations}"
        
        # Check cache first
        if use_cache:
            cached_recommendations = await self.cache.get_cached_recommendations(cache_key)
            if cached_recommendations:
                return cached_recommendations
        
        try:
            # Get content-based similar videos
            similar_videos = self.content_recommender.get_content_recommendations(
                video_id, n_recommendations
            )
            
            # Enrich with video metadata
            recommendations = []
            for similar_video_id, score in similar_videos:
                video_data = await self.video_analytics.get_analysis(similar_video_id)
                if video_data:
                    recommendations.append({
                        "video_id": similar_video_id,
                        "similarity_score": float(score),
                        "title": video_data.get("metadata", {}).get("title", ""),
                        "tags": video_data.get("tags", []),
                        "quality_score": video_data.get("quality", {}).get("overall_score", 0),
                        "thumbnails": video_data.get("thumbnails", []),
                        "duration": video_data.get("metadata", {}).get("duration", 0),
                        "user_id": video_data.get("user_id"),
                        "created_at": video_data.get("created_at")
                    })
            
            # Cache recommendations
            if use_cache:
                await self.cache.cache_recommendations(cache_key, recommendations)
            
            return recommendations
            
        except Exception as e:
            logger.error(f"Failed to get similar videos: {e}")
            return []
    
    async def _get_popular_videos(self, limit: int) -> List[Dict[str, Any]]:
        """Get popular videos as fallback recommendations"""
        try:
            collection = mongodb.get_collection("video_analytics")
            
            cursor = collection.find({
                "quality.overall_score": {"$gte": 0.6},
                "status": {"$in": ["completed", "approved"]}
            }).sort("quality.overall_score", -1).limit(limit)
            
            videos = await cursor.to_list(length=limit)
            
            recommendations = []
            for video in videos:
                recommendations.append({
                    "video_id": video["video_id"],
                    "score": video.get("quality", {}).get("overall_score", 0),
                    "title": video.get("metadata", {}).get("title", ""),
                    "tags": video.get("tags", []),
                    "quality_score": video.get("quality", {}).get("overall_score", 0),
                    "thumbnails": video.get("thumbnails", []),
                    "duration": video.get("metadata", {}).get("duration", 0),
                    "user_id": video.get("user_id"),
                    "created_at": video.get("created_at")
                })
            
            return recommendations
            
        except Exception as e:
            logger.error(f"Failed to get popular videos: {e}")
            return []
    
    async def update_user_interaction(
        self,
        user_id: str,
        video_id: str,
        interaction_type: str
    ) -> None:
        """Update user interaction and invalidate cache"""
        try:
            # Track interaction
            await self.user_preferences.track_interaction(user_id, video_id, interaction_type)
            
            # Invalidate user's cached recommendations
            await self.cache.invalidate_user_cache(user_id)
            
            logger.info(f"Updated interaction: {user_id} -> {video_id} ({interaction_type})")
            
        except Exception as e:
            logger.error(f"Failed to update user interaction: {e}")


# Global recommendation engine instance (initialized in main.py)
recommendation_engine = None