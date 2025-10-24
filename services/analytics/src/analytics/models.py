"""Base classes for ML model integration"""
import asyncio
import logging
from abc import ABC, abstractmethod
from typing import Dict, Any, List, Optional, Tuple
from concurrent.futures import ThreadPoolExecutor
import numpy as np

logger = logging.getLogger(__name__)


class BaseMLModel(ABC):
    """Base class for ML models with async support"""
    
    def __init__(self, model_name: str, max_workers: int = 2):
        self.model_name = model_name
        self.model = None
        self.is_loaded = False
        self.executor = ThreadPoolExecutor(max_workers=max_workers)
    
    @abstractmethod
    async def load_model(self) -> None:
        """Load the ML model (implemented by subclasses)"""
        pass
    
    @abstractmethod
    def _predict_sync(self, input_data: Any) -> Any:
        """Synchronous prediction method (implemented by subclasses)"""
        pass
    
    async def predict(self, input_data: Any) -> Any:
        """Async prediction wrapper"""
        if not self.is_loaded:
            await self.load_model()
        
        # Run prediction in thread pool to avoid blocking
        loop = asyncio.get_event_loop()
        result = await loop.run_in_executor(
            self.executor,
            self._predict_sync,
            input_data
        )
        return result
    
    async def batch_predict(self, input_batch: List[Any]) -> List[Any]:
        """Batch prediction for multiple inputs"""
        if not self.is_loaded:
            await self.load_model()
        
        # Process batch in thread pool
        loop = asyncio.get_event_loop()
        results = await loop.run_in_executor(
            self.executor,
            self._batch_predict_sync,
            input_batch
        )
        return results
    
    def _batch_predict_sync(self, input_batch: List[Any]) -> List[Any]:
        """Synchronous batch prediction"""
        return [self._predict_sync(input_data) for input_data in input_batch]
    
    def cleanup(self) -> None:
        """Cleanup resources"""
        if self.executor:
            self.executor.shutdown(wait=True)


class ContentSafetyModel(BaseMLModel):
    """Content safety model for video moderation"""
    
    def __init__(self, model_name: str = "unitary/toxic-bert"):
        super().__init__(model_name)
        self.pipeline = None
    
    async def load_model(self) -> None:
        """Load content safety model"""
        if self.is_loaded:
            return
        
        try:
            # Import transformers in thread to avoid blocking
            loop = asyncio.get_event_loop()
            self.pipeline = await loop.run_in_executor(
                self.executor,
                self._load_transformers_model
            )
            self.is_loaded = True
            logger.info(f"Loaded content safety model: {self.model_name}")
            
        except Exception as e:
            logger.error(f"Failed to load content safety model: {e}")
            raise
    
    def _load_transformers_model(self):
        """Load transformers model synchronously"""
        from transformers import pipeline
        return pipeline(
            "text-classification",
            model=self.model_name,
            device=-1  # Use CPU
        )
    
    def _predict_sync(self, text: str) -> Dict[str, Any]:
        """Predict content safety score"""
        if not self.pipeline:
            raise RuntimeError("Model not loaded")
        
        try:
            results = self.pipeline(text)
            
            # Convert to safety score (0.0 = unsafe, 1.0 = safe)
            toxic_score = 0.0
            for result in results:
                if result['label'].lower() in ['toxic', 'hate', 'threat']:
                    toxic_score = max(toxic_score, result['score'])
            
            safety_score = 1.0 - toxic_score
            is_safe = safety_score > 0.7  # Threshold for safety
            
            return {
                "overall_score": safety_score,
                "is_safe": is_safe,
                "raw_results": results
            }
            
        except Exception as e:
            logger.error(f"Content safety prediction failed: {e}")
            return {
                "overall_score": 0.5,  # Neutral score on error
                "is_safe": True,
                "raw_results": [],
                "error": str(e)
            }


class VideoQualityModel(BaseMLModel):
    """Video quality analysis model"""
    
    def __init__(self):
        super().__init__("video_quality_analyzer")
    
    async def load_model(self) -> None:
        """Load video quality model (OpenCV-based)"""
        if self.is_loaded:
            return
        
        try:
            # Import OpenCV in thread
            loop = asyncio.get_event_loop()
            await loop.run_in_executor(
                self.executor,
                self._load_opencv
            )
            self.is_loaded = True
            logger.info("Loaded video quality model")
            
        except Exception as e:
            logger.error(f"Failed to load video quality model: {e}")
            raise
    
    def _load_opencv(self):
        """Load OpenCV synchronously"""
        import cv2
        # Test OpenCV is working
        cv2.getVersionString()
    
    def _predict_sync(self, frames: List[np.ndarray]) -> Dict[str, Any]:
        """Analyze video quality from frames"""
        import cv2
        
        if not frames:
            return {
                "sharpness_score": 0.0,
                "brightness_score": 0.0,
                "contrast_score": 0.0,
                "overall_score": 0.0,
                "resolution_category": "unknown"
            }
        
        try:
            sharpness_scores = []
            brightness_scores = []
            contrast_scores = []
            
            for frame in frames:
                # Convert to grayscale for analysis
                gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
                
                # Calculate sharpness using Laplacian variance
                sharpness = cv2.Laplacian(gray, cv2.CV_64F).var()
                sharpness_scores.append(sharpness)
                
                # Calculate brightness (mean pixel value)
                brightness = np.mean(gray)
                brightness_scores.append(brightness)
                
                # Calculate contrast (standard deviation)
                contrast = np.std(gray)
                contrast_scores.append(contrast)
            
            # Average scores across frames
            avg_sharpness = np.mean(sharpness_scores)
            avg_brightness = np.mean(brightness_scores)
            avg_contrast = np.mean(contrast_scores)
            
            # Normalize scores to 0-1 range
            sharpness_norm = min(avg_sharpness / 1000.0, 1.0)  # Normalize sharpness
            brightness_norm = avg_brightness / 255.0  # Brightness already 0-255
            contrast_norm = min(avg_contrast / 128.0, 1.0)  # Normalize contrast
            
            # Calculate overall score (weighted average)
            overall_score = (
                0.4 * sharpness_norm +
                0.3 * brightness_norm +
                0.3 * contrast_norm
            )
            
            # Determine resolution category based on frame size
            height, width = frames[0].shape[:2]
            if height >= 2160:
                resolution_category = "ultra"
            elif height >= 1080:
                resolution_category = "high"
            elif height >= 720:
                resolution_category = "medium"
            else:
                resolution_category = "low"
            
            return {
                "sharpness_score": float(sharpness_norm),
                "brightness_score": float(brightness_norm),
                "contrast_score": float(contrast_norm),
                "overall_score": float(overall_score),
                "resolution_category": resolution_category
            }
            
        except Exception as e:
            logger.error(f"Video quality analysis failed: {e}")
            return {
                "sharpness_score": 0.0,
                "brightness_score": 0.0,
                "contrast_score": 0.0,
                "overall_score": 0.0,
                "resolution_category": "unknown",
                "error": str(e)
            }


class RecommendationModel(BaseMLModel):
    """Content-based recommendation model"""
    
    def __init__(self):
        super().__init__("recommendation_engine")
        self.vectorizer = None
        self.feature_matrix = None
        self.video_ids = []
    
    async def load_model(self) -> None:
        """Load recommendation model"""
        if self.is_loaded:
            return
        
        try:
            # Import scikit-learn in thread
            loop = asyncio.get_event_loop()
            await loop.run_in_executor(
                self.executor,
                self._load_sklearn
            )
            self.is_loaded = True
            logger.info("Loaded recommendation model")
            
        except Exception as e:
            logger.error(f"Failed to load recommendation model: {e}")
            raise
    
    def _load_sklearn(self):
        """Load scikit-learn synchronously"""
        from sklearn.feature_extraction.text import TfidfVectorizer
        from sklearn.metrics.pairwise import cosine_similarity
        
        self.vectorizer = TfidfVectorizer(
            max_features=1000,
            stop_words='english',
            ngram_range=(1, 2)
        )
    
    async def build_feature_matrix(self, video_data: List[Dict[str, Any]]) -> None:
        """Build feature matrix from video data"""
        if not self.is_loaded:
            await self.load_model()
        
        loop = asyncio.get_event_loop()
        await loop.run_in_executor(
            self.executor,
            self._build_feature_matrix_sync,
            video_data
        )
    
    def _build_feature_matrix_sync(self, video_data: List[Dict[str, Any]]) -> None:
        """Build feature matrix synchronously"""
        from sklearn.metrics.pairwise import cosine_similarity
        
        # Extract text features (tags, titles, descriptions)
        text_features = []
        self.video_ids = []
        
        for video in video_data:
            tags = video.get('tags', [])
            title = video.get('title', '')
            description = video.get('description', '')
            
            # Combine text features
            combined_text = ' '.join(tags) + ' ' + title + ' ' + description
            text_features.append(combined_text)
            self.video_ids.append(video['video_id'])
        
        # Build TF-IDF matrix
        if text_features:
            self.feature_matrix = self.vectorizer.fit_transform(text_features)
            logger.info(f"Built feature matrix with {len(text_features)} videos")
    
    def _predict_sync(self, video_id: str) -> List[Tuple[str, float]]:
        """Get recommendations for a video"""
        from sklearn.metrics.pairwise import cosine_similarity
        
        if self.feature_matrix is None or video_id not in self.video_ids:
            return []
        
        try:
            # Find video index
            video_idx = self.video_ids.index(video_id)
            
            # Calculate similarity scores
            video_features = self.feature_matrix[video_idx]
            similarity_scores = cosine_similarity(video_features, self.feature_matrix).flatten()
            
            # Get top similar videos (excluding the input video)
            similar_indices = similarity_scores.argsort()[::-1][1:11]  # Top 10, excluding self
            
            recommendations = []
            for idx in similar_indices:
                if similarity_scores[idx] > 0.1:  # Minimum similarity threshold
                    recommendations.append((
                        self.video_ids[idx],
                        float(similarity_scores[idx])
                    ))
            
            return recommendations
            
        except Exception as e:
            logger.error(f"Recommendation generation failed: {e}")
            return []


# Global model instances
content_safety_model = ContentSafetyModel()
video_quality_model = VideoQualityModel()
recommendation_model = RecommendationModel()