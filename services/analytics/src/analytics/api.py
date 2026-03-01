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


class PopularVideoResponse(BaseModel):
    video_id: str
    score: float
    title: str = ""
    tags: List[str] = []
    quality_score: float = 0.0
    thumbnails: List[str] = []
    duration: int = 0
    user_id: str = ""
    created_at: Optional[str] = None


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
    response_model=List[PopularVideoResponse],
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
        
        return [PopularVideoResponse(**video) for video in popular_videos]
        
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


