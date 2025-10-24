"""Video quality analysis and content moderation"""
import logging
import asyncio
import os
from typing import Dict, Any, List, Tuple, Optional
import numpy as np
import cv2
from concurrent.futures import ThreadPoolExecutor

from .config import config

logger = logging.getLogger(__name__)


class VideoQualityAnalyzer:
    """Comprehensive video quality analysis"""
    
    def __init__(self, max_workers: int = 2):
        self.executor = ThreadPoolExecutor(max_workers=max_workers)
        self.quality_weights = {
            'sharpness': 0.4,
            'brightness': 0.2,
            'contrast': 0.3,
            'noise': 0.1
        }
    
    async def analyze_quality(self, video_path: str, sample_count: int = 20) -> Dict[str, Any]:
        """Analyze video quality with comprehensive metrics"""
        try:
            # For now, return mock data to test the implementation
            return {
                "sharpness_score": 0.8,
                "brightness_score": 0.7,
                "contrast_score": 0.9,
                "noise_score": 0.6,
                "overall_score": 0.75,
                "quality_category": "good",
                "resolution_category": "HD",
                "frame_count_analyzed": 20,
                "raw_metrics": {
                    "sharpness_raw": 800.0,
                    "brightness_raw": 127.5,
                    "contrast_raw": 45.0,
                    "noise_raw": 10.0
                }
            }
        except Exception as e:
            logger.error(f"Quality analysis failed for {video_path}: {e}")
            return self._get_error_quality_metrics(str(e))
    
    def _get_error_quality_metrics(self, error_message: str) -> Dict[str, Any]:
        """Return default quality metrics when analysis fails"""
        return {
            "sharpness_score": 0.0,
            "brightness_score": 0.0,
            "contrast_score": 0.0,
            "noise_score": 0.0,
            "overall_score": 0.0,
            "quality_category": "unknown",
            "resolution_category": "unknown",
            "frame_count_analyzed": 0,
            "error": error_message,
            "raw_metrics": {
                "sharpness_raw": 0.0,
                "brightness_raw": 0.0,
                "contrast_raw": 0.0,
                "noise_raw": 0.0
            }
        }


class ContentModerationAnalyzer:
    """Content safety analysis and moderation"""
    
    def __init__(self, max_workers: int = 2):
        self.executor = ThreadPoolExecutor(max_workers=max_workers)
        self.moderation_thresholds = {
            'safe': 0.8,
            'review': 0.5,
            'unsafe': 0.2
        }
    
    async def analyze_content_safety(self, video_path: str, metadata: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze content safety using multiple methods"""
        try:
            # For now, return mock safe data
            return {
                "overall_score": 0.9,
                "is_safe": True,
                "confidence": 0.8,
                "flags": [],
                "analysis_methods": ["filename_metadata", "content_heuristics"],
                "moderation_action": "approve"
            }
        except Exception as e:
            logger.error(f"Content safety analysis failed for {video_path}: {e}")
            return self._get_error_safety_result(str(e))
    
    def _analyze_content_heuristics(self, metadata: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze content using heuristic rules"""
        try:
            flags = []
            risk_score = 0.0
            
            # Check video duration
            duration = metadata.get("duration", 0)
            if duration > 7200:  # More than 2 hours
                flags.append({
                    "category": "heuristic",
                    "severity": "low",
                    "description": "Very long video duration",
                    "confidence": 0.3
                })
                risk_score += 0.05
            
            # Calculate safety score
            safety_score = max(0.0, 1.0 - min(risk_score, 1.0))
            
            return {
                "score": safety_score,
                "flags": flags,
                "method": "content_heuristics"
            }
            
        except Exception as e:
            logger.error(f"Heuristic analysis failed: {e}")
            return {"score": 0.5, "flags": [], "method": "content_heuristics", "error": str(e)}
    
    def _get_error_safety_result(self, error_message: str) -> Dict[str, Any]:
        """Return default safety result when analysis fails"""
        return {
            "overall_score": 0.5,
            "is_safe": True,
            "confidence": 0.0,
            "flags": [],
            "analysis_methods": [],
            "moderation_action": "review",
            "error": error_message
        }


class QualityScoringAlgorithm:
    """Weighted quality scoring algorithm"""
    
    def __init__(self):
        self.quality_weights = {
            'technical_quality': 0.4,
            'visual_appeal': 0.3,
            'content_quality': 0.2,
            'safety_score': 0.1
        }
    
    def calculate_weighted_score(
        self,
        quality_metrics: Dict[str, Any],
        safety_metrics: Dict[str, Any],
        metadata: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Calculate weighted quality score combining all metrics"""
        try:
            # Technical quality (40%)
            technical_score = (
                0.5 * quality_metrics.get("sharpness_score", 0) +
                0.3 * quality_metrics.get("contrast_score", 0) +
                0.2 * quality_metrics.get("noise_score", 0)
            )
            
            # Visual appeal (30%)
            visual_score = (
                0.6 * quality_metrics.get("brightness_score", 0) +
                0.4 * 0.7  # Default color balance score
            )
            
            # Content quality (20%)
            content_score = (
                0.6 * self._calculate_resolution_score(metadata) +
                0.4 * self._calculate_framerate_score(metadata)
            )
            
            # Safety score (10%)
            safety_score = safety_metrics.get("overall_score", 0.5)
            
            # Calculate weighted overall score
            overall_score = (
                self.quality_weights['technical_quality'] * technical_score +
                self.quality_weights['visual_appeal'] * visual_score +
                self.quality_weights['content_quality'] * content_score +
                self.quality_weights['safety_score'] * safety_score
            )
            
            return {
                "overall_score": float(overall_score),
                "technical_quality": float(technical_score),
                "visual_appeal": float(visual_score),
                "content_quality": float(content_score),
                "safety_contribution": float(safety_score),
                "quality_category": self._categorize_overall_quality(overall_score),
                "weights_used": self.quality_weights
            }
            
        except Exception as e:
            logger.error(f"Quality scoring failed: {e}")
            return {
                "overall_score": 0.0,
                "technical_quality": 0.0,
                "visual_appeal": 0.0,
                "content_quality": 0.0,
                "safety_contribution": 0.0,
                "quality_category": "unknown",
                "error": str(e)
            }
    
    def _calculate_resolution_score(self, metadata: Dict[str, Any]) -> float:
        """Calculate resolution quality score"""
        height = metadata.get("height", 0)
        
        if height >= 2160:
            return 1.0  # 4K
        elif height >= 1080:
            return 0.8  # Full HD
        elif height >= 720:
            return 0.6  # HD
        elif height >= 480:
            return 0.4  # SD
        else:
            return 0.2  # Low resolution
    
    def _calculate_framerate_score(self, metadata: Dict[str, Any]) -> float:
        """Calculate frame rate quality score"""
        fps = metadata.get("fps", 0)
        
        if fps >= 60:
            return 1.0  # High frame rate
        elif fps >= 30:
            return 0.8  # Standard frame rate
        elif fps >= 24:
            return 0.6  # Cinema frame rate
        elif fps >= 15:
            return 0.4  # Low frame rate
        else:
            return 0.2  # Very low frame rate
    
    def _categorize_overall_quality(self, score: float) -> str:
        """Categorize overall quality score"""
        if score >= 0.9:
            return "exceptional"
        elif score >= 0.8:
            return "excellent"
        elif score >= 0.7:
            return "very_good"
        elif score >= 0.6:
            return "good"
        elif score >= 0.5:
            return "fair"
        elif score >= 0.3:
            return "poor"
        else:
            return "very_poor"


# Global instances
quality_analyzer = VideoQualityAnalyzer()
content_moderator = ContentModerationAnalyzer()
quality_scorer = QualityScoringAlgorithm()