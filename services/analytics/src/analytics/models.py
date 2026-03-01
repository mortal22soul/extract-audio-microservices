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


# Global model instances
video_quality_model = VideoQualityModel()