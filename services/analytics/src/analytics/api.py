"""
API endpoints for Analytics Service
"""
import logging
from typing import List, Dict, Any, Optional
from fastapi import APIRouter, HTTPException, Query, Path
from pydantic import BaseModel, Field

from .services import AnalyticsService
from .database import mongodb

logger = logging.getLogger(__name__)

# Initialize router
router = APIRouter(prefix="/api/v1", tags=["analytics"])

# Initialize services
analytics_service = AnalyticsService()


# Pydantic models for API
class RecommendationResponse(BaseModel):
    video_id: str
    score: float
    title: str = ""
    tags: List[str] = []
    quality_score: float = 0.0
    thumbnails: List[str] = []
    duration: int = 0
    user_id: str = ""
    created_at: Optional[str] = None


class SimilarVideoResponse(BaseModel):
    video_id: str
    similarity_score: float
    title: str = ""
    tags: List[str] = []
    quality_score: float = 0.0
    thumbnails: List[str] = []
    duration: int = 0
    user_id: str = ""
    created_at: Optional[str] = None


class UserInteractionRequest(BaseModel):
    video_id: str = Field(..., description="Video ID that user interacted with")
    interaction_type: str = Field(..., description="Type of interaction: view, download, like, share")


class UserPreferencesResponse(BaseModel):
    user_id: str
    preferred_tags: Dict[str, int] = {}
    quality_preference: float = 0.5
    interaction_count: int = 0
    updated_at: Optional[str] = None


class VideoAnalysisResponse(BaseModel):
    video_id: str
    user_id: str
    metadata: Dict[str, Any] = {}
    quality: Dict[str, Any] = {}
    safety: Dict[str, Any] = {}
    tags: List[str] = []
    thumbnails: List[str] = []
    status: str = "pending"
    created_at: Optional[str] = None


# Recommendation endpoints
@router.get(
    "/recommendations/user/{user_id}",
    response_model=List[RecommendationResponse],
    summary="Get personalized recommendations for a user"
)
async def get_user_recommendations(
    user_id: str = Path(..., description="User ID to get recommendations for"),
    limit: int = Query(10, ge=1, le=50, description="Number of recommendations to return"),
    use_cache: bool = Query(True, description="Whether to use cached recommendations")
):
    """
    Get personalized video recommendations for a user using hybrid filtering.
    
    Combines content-based filtering (based on video features and user preferences)
    with collaborative filtering (based on similar users' behavior).
    """
    try:
        from .recommendation_engine import recommendation_engine
        
        if recommendation_engine is None:
            raise HTTPException(status_code=503, detail="Recommendation engine not initialized")
        
        recommendations = await recommendation_engine.get_recommendations(
            user_id=user_id,
            n_recommendations=limit,
            use_cache=use_cache
        )
        
        return [RecommendationResponse(**rec) for rec in recommendations]
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to get user recommendations: {e}")
        raise HTTPException(status_code=500, detail="Failed to get recommendations")


@router.get(
    "/recommendations/video/{video_id}/similar",
    response_model=List[SimilarVideoResponse],
    summary="Get videos similar to a specific video"
)
async def get_similar_videos(
    video_id: str = Path(..., description="Video ID to find similar videos for"),
    limit: int = Query(10, ge=1, le=50, description="Number of similar videos to return"),
    use_cache: bool = Query(True, description="Whether to use cached recommendations")
):
    """
    Get videos similar to a specific video using content-based filtering.
    
    Uses video features like tags, quality metrics, duration, and metadata
    to find similar content.
    """
    try:
        from .recommendation_engine import recommendation_engine
        
        if recommendation_engine is None:
            raise HTTPException(status_code=503, detail="Recommendation engine not initialized")
        
        similar_videos = await recommendation_engine.get_similar_videos(
            video_id=video_id,
            n_recommendations=limit,
            use_cache=use_cache
        )
        
        return [SimilarVideoResponse(**video) for video in similar_videos]
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to get similar videos: {e}")
        raise HTTPException(status_code=500, detail="Failed to get similar videos")


@router.post(
    "/recommendations/user/{user_id}/interaction",
    summary="Track user interaction with a video"
)
async def track_user_interaction(
    interaction: UserInteractionRequest,
    user_id: str = Path(..., description="User ID")
):
    """
    Track user interaction with a video to improve future recommendations.
    
    Supported interaction types:
    - view: User viewed the video
    - download: User downloaded the converted MP3
    - like: User liked the video
    - share: User shared the video
    """
    try:
        from .recommendation_engine import recommendation_engine
        
        if recommendation_engine is None:
            raise HTTPException(status_code=503, detail="Recommendation engine not initialized")
        
        await recommendation_engine.update_user_interaction(
            user_id=user_id,
            video_id=interaction.video_id,
            interaction_type=interaction.interaction_type
        )
        
        return {"message": "Interaction tracked successfully"}
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to track user interaction: {e}")
        raise HTTPException(status_code=500, detail="Failed to track interaction")


@router.get(
    "/recommendations/retrain",
    summary="Retrain recommendation models"
)
async def retrain_models():
    """
    Manually trigger retraining of recommendation models.
    
    This endpoint forces retraining of both content-based and collaborative
    filtering models with the latest data.
    """
    try:
        from .recommendation_engine import recommendation_engine
        
        if recommendation_engine is None:
            raise HTTPException(status_code=503, detail="Recommendation engine not initialized")
        
        await recommendation_engine.train_models(force_retrain=True)
        return {"message": "Models retrained successfully"}
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to retrain models: {e}")
        raise HTTPException(status_code=500, detail="Failed to retrain models")


# User preferences endpoints
@router.get(
    "/users/{user_id}/preferences",
    response_model=UserPreferencesResponse,
    summary="Get user preferences"
)
async def get_user_preferences(
    user_id: str = Path(..., description="User ID")
):
    """Get user preferences and interaction history."""
    try:
        preferences = await analytics_service.user_preferences.get_preferences(user_id)
        
        if not preferences:
            raise HTTPException(status_code=404, detail="User preferences not found")
        
        return UserPreferencesResponse(**preferences)
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to get user preferences: {e}")
        raise HTTPException(status_code=500, detail="Failed to get user preferences")


# Video analysis endpoints
@router.get(
    "/videos/{video_id}/analysis",
    response_model=VideoAnalysisResponse,
    summary="Get video analysis results"
)
async def get_video_analysis(
    video_id: str = Path(..., description="Video ID")
):
    """Get complete analysis results for a video."""
    try:
        analysis = await analytics_service.video_analytics.get_analysis(video_id)
        
        if not analysis:
            raise HTTPException(status_code=404, detail="Video analysis not found")
        
        return VideoAnalysisResponse(**analysis)
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to get video analysis: {e}")
        raise HTTPException(status_code=500, detail="Failed to get video analysis")


@router.get(
    "/users/{user_id}/videos",
    response_model=List[VideoAnalysisResponse],
    summary="Get user's video analyses"
)
async def get_user_videos(
    user_id: str = Path(..., description="User ID"),
    limit: int = Query(50, ge=1, le=100, description="Number of videos to return")
):
    """Get all video analyses for a specific user."""
    try:
        videos = await analytics_service.video_analytics.get_user_videos(user_id, limit)
        
        return [VideoAnalysisResponse(**video) for video in videos]
        
    except Exception as e:
        logger.error(f"Failed to get user videos: {e}")
        raise HTTPException(status_code=500, detail="Failed to get user videos")


# Popular content endpoints
@router.get(
    "/videos/popular",
    response_model=List[RecommendationResponse],
    summary="Get popular videos"
)
async def get_popular_videos(
    limit: int = Query(20, ge=1, le=50, description="Number of popular videos to return"),
    min_quality: float = Query(0.6, ge=0.0, le=1.0, description="Minimum quality score")
):
    """Get popular videos based on quality scores."""
    try:
        collection = mongodb.get_collection("video_analytics")
        
        cursor = collection.find({
            "quality.overall_score": {"$gte": min_quality},
            "status": {"$in": ["completed", "approved"]}
        }).sort("quality.overall_score", -1).limit(limit)
        
        videos = await cursor.to_list(length=limit)
        
        popular_videos = []
        for video in videos:
            popular_videos.append({
                "video_id": video["video_id"],
                "score": video.get("quality", {}).get("overall_score", 0),
                "title": video.get("metadata", {}).get("title", ""),
                "tags": video.get("tags", []),
                "quality_score": video.get("quality", {}).get("overall_score", 0),
                "thumbnails": video.get("thumbnails", []),
                "duration": video.get("metadata", {}).get("duration", 0),
                "user_id": video.get("user_id", ""),
                "created_at": str(video.get("created_at", ""))
            })
        
        return [RecommendationResponse(**video) for video in popular_videos]
        
    except Exception as e:
        logger.error(f"Failed to get popular videos: {e}")
        raise HTTPException(status_code=500, detail="Failed to get popular videos")


# Search endpoints
@router.get(
    "/videos/search",
    response_model=List[VideoAnalysisResponse],
    summary="Search videos by tags and metadata"
)
async def search_videos(
    query: str = Query(..., description="Search query"),
    tags: Optional[List[str]] = Query(None, description="Filter by tags"),
    min_quality: float = Query(0.0, ge=0.0, le=1.0, description="Minimum quality score"),
    limit: int = Query(20, ge=1, le=100, description="Number of results to return")
):
    """Search videos by tags and metadata."""
    try:
        collection = mongodb.get_collection("video_analytics")
        
        # Build search filter
        search_filter = {
            "status": {"$in": ["completed", "approved"]},
            "quality.overall_score": {"$gte": min_quality}
        }
        
        # Add text search
        if query:
            search_filter["$or"] = [
                {"tags": {"$regex": query, "$options": "i"}},
                {"metadata.title": {"$regex": query, "$options": "i"}},
                {"metadata.description": {"$regex": query, "$options": "i"}}
            ]
        
        # Add tag filter
        if tags:
            search_filter["tags"] = {"$in": tags}
        
        cursor = collection.find(search_filter).sort("quality.overall_score", -1).limit(limit)
        videos = await cursor.to_list(length=limit)
        
        return [VideoAnalysisResponse(**video) for video in videos]
        
    except Exception as e:
        logger.error(f"Failed to search videos: {e}")
        raise HTTPException(status_code=500, detail="Failed to search videos")


# Health check for recommendation system
@router.get(
    "/recommendations/health",
    summary="Check recommendation system health"
)
async def recommendation_health():
    """Check the health of the recommendation system."""
    try:
        from .recommendation_engine import recommendation_engine
        
        if recommendation_engine is None:
            return {
                "status": "not_initialized",
                "message": "Recommendation engine not initialized"
            }
        
        health_status = {
            "status": "healthy",
            "content_recommender_trained": recommendation_engine.content_recommender.is_trained,
            "collaborative_recommender_trained": recommendation_engine.collaborative_recommender.is_trained,
            "cache_connected": recommendation_engine.cache.redis_client is not None,
            "last_training_time": str(recommendation_engine.last_training_time) if recommendation_engine.last_training_time else None,
            "models_status": {
                "content_based": {
                    "feature_matrix_size": recommendation_engine.content_recommender.feature_matrix.shape if recommendation_engine.content_recommender.feature_matrix is not None else None,
                    "video_count": len(recommendation_engine.content_recommender.video_ids)
                },
                "collaborative": {
                    "user_count": len(recommendation_engine.collaborative_recommender.user_ids),
                    "video_count": len(recommendation_engine.collaborative_recommender.video_ids),
                    "matrix_shape": recommendation_engine.collaborative_recommender.user_item_matrix.shape if recommendation_engine.collaborative_recommender.user_item_matrix is not None else None
                }
            }
        }
        
        return health_status
        
    except Exception as e:
        logger.error(f"Failed to get recommendation health: {e}")
        raise HTTPException(status_code=500, detail="Failed to get recommendation health")