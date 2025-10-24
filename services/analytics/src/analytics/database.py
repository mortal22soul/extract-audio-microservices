"""Database connection and models for analytics service"""
import asyncio
from typing import Optional, Dict, Any, List
from datetime import datetime
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorDatabase, AsyncIOMotorCollection
from pymongo import IndexModel, ASCENDING, DESCENDING
import logging

from .config import config

logger = logging.getLogger(__name__)


class MongoDB:
    """MongoDB connection manager"""
    
    def __init__(self):
        self.client: Optional[AsyncIOMotorClient] = None
        self.database: Optional[AsyncIOMotorDatabase] = None
        self.collections: Dict[str, AsyncIOMotorCollection] = {}
    
    async def connect(self) -> None:
        """Connect to MongoDB"""
        try:
            self.client = AsyncIOMotorClient(config.MONGODB_URL)
            self.database = self.client[config.MONGODB_DATABASE]
            
            # Test connection
            await self.client.admin.command('ping')
            logger.info("Connected to MongoDB successfully")
            
            # Initialize collections
            await self._initialize_collections()
            
        except Exception as e:
            logger.error(f"Failed to connect to MongoDB: {e}")
            raise
    
    async def disconnect(self) -> None:
        """Disconnect from MongoDB"""
        if self.client:
            self.client.close()
            logger.info("Disconnected from MongoDB")
    
    async def _initialize_collections(self) -> None:
        """Initialize collections and indexes"""
        # Video analytics collection
        self.collections["video_analytics"] = self.database["video_analytics"]
        await self._create_video_analytics_indexes()
        
        # User preferences collection
        self.collections["user_preferences"] = self.database["user_preferences"]
        await self._create_user_preferences_indexes()
        
        # Video features collection (for recommendations)
        self.collections["video_features"] = self.database["video_features"]
        await self._create_video_features_indexes()
        
        # User interactions collection (for collaborative filtering)
        self.collections["user_interactions"] = self.database["user_interactions"]
        await self._create_user_interactions_indexes()
        
        logger.info("Collections and indexes initialized")
    
    async def _create_video_analytics_indexes(self) -> None:
        """Create indexes for video_analytics collection"""
        indexes = [
            IndexModel([("video_id", ASCENDING)], unique=True),
            IndexModel([("user_id", ASCENDING)]),
            IndexModel([("created_at", DESCENDING)]),
            IndexModel([("quality.overall_score", DESCENDING)]),
            IndexModel([("safety.overall_score", DESCENDING)]),
            IndexModel([("tags", ASCENDING)]),
        ]
        await self.collections["video_analytics"].create_indexes(indexes)
    
    async def _create_user_preferences_indexes(self) -> None:
        """Create indexes for user_preferences collection"""
        indexes = [
            IndexModel([("user_id", ASCENDING)], unique=True),
            IndexModel([("updated_at", DESCENDING)]),
        ]
        await self.collections["user_preferences"].create_indexes(indexes)
    
    async def _create_video_features_indexes(self) -> None:
        """Create indexes for video_features collection"""
        indexes = [
            IndexModel([("video_id", ASCENDING)], unique=True),
            IndexModel([("user_id", ASCENDING)]),
            IndexModel([("created_at", DESCENDING)]),
        ]
        await self.collections["video_features"].create_indexes(indexes)
    
    async def _create_user_interactions_indexes(self) -> None:
        """Create indexes for user_interactions collection"""
        indexes = [
            IndexModel([("user_id", ASCENDING), ("video_id", ASCENDING)]),
            IndexModel([("user_id", ASCENDING), ("timestamp", DESCENDING)]),
            IndexModel([("video_id", ASCENDING)]),
            IndexModel([("timestamp", DESCENDING)]),
            IndexModel([("interaction_type", ASCENDING)]),
        ]
        await self.collections["user_interactions"].create_indexes(indexes)
    
    def get_collection(self, name: str) -> AsyncIOMotorCollection:
        """Get collection by name"""
        if name not in self.collections:
            raise ValueError(f"Collection {name} not found")
        return self.collections[name]


class VideoAnalyticsModel:
    """Model for video analytics data"""
    
    def __init__(self, mongodb: MongoDB):
        self.collection = mongodb.get_collection("video_analytics")
    
    async def create_analysis(self, analysis_data: Dict[str, Any]) -> str:
        """Create new video analysis record"""
        analysis_data["created_at"] = datetime.utcnow()
        analysis_data["updated_at"] = datetime.utcnow()
        
        result = await self.collection.insert_one(analysis_data)
        return str(result.inserted_id)
    
    async def get_analysis(self, video_id: str) -> Optional[Dict[str, Any]]:
        """Get video analysis by video_id"""
        return await self.collection.find_one({"video_id": video_id})
    
    async def update_analysis(self, video_id: str, update_data: Dict[str, Any]) -> bool:
        """Update video analysis"""
        update_data["updated_at"] = datetime.utcnow()
        
        result = await self.collection.update_one(
            {"video_id": video_id},
            {"$set": update_data}
        )
        return result.modified_count > 0
    
    async def get_user_videos(self, user_id: str, limit: int = 50) -> List[Dict[str, Any]]:
        """Get all video analyses for a user"""
        cursor = self.collection.find(
            {"user_id": user_id}
        ).sort("created_at", DESCENDING).limit(limit)
        
        return await cursor.to_list(length=limit)
    
    async def get_similar_videos(self, tags: List[str], exclude_video_id: str, limit: int = 10) -> List[Dict[str, Any]]:
        """Get videos with similar tags"""
        cursor = self.collection.find({
            "tags": {"$in": tags},
            "video_id": {"$ne": exclude_video_id}
        }).sort("quality.overall_score", DESCENDING).limit(limit)
        
        return await cursor.to_list(length=limit)


class UserPreferencesModel:
    """Model for user preferences data"""
    
    def __init__(self, mongodb: MongoDB):
        self.collection = mongodb.get_collection("user_preferences")
    
    async def get_preferences(self, user_id: str) -> Optional[Dict[str, Any]]:
        """Get user preferences"""
        return await self.collection.find_one({"user_id": user_id})
    
    async def update_preferences(self, user_id: str, preferences: Dict[str, Any]) -> bool:
        """Update user preferences"""
        preferences["updated_at"] = datetime.utcnow()
        
        result = await self.collection.update_one(
            {"user_id": user_id},
            {"$set": preferences},
            upsert=True
        )
        return result.modified_count > 0 or result.upserted_id is not None
    
    async def track_interaction(self, user_id: str, video_id: str, interaction_type: str) -> bool:
        """Track user interaction with video"""
        interaction = {
            "video_id": video_id,
            "type": interaction_type,  # "view", "download", "like", etc.
            "timestamp": datetime.utcnow()
        }
        
        # Store in user preferences
        result = await self.collection.update_one(
            {"user_id": user_id},
            {
                "$push": {"interactions": interaction},
                "$set": {"updated_at": datetime.utcnow()}
            },
            upsert=True
        )
        
        # Also store in separate interactions collection for collaborative filtering
        interactions_collection = self.collection.database["user_interactions"]
        await interactions_collection.insert_one({
            "user_id": user_id,
            "video_id": video_id,
            "interaction_type": interaction_type,
            "timestamp": datetime.utcnow()
        })
        
        return result.modified_count > 0 or result.upserted_id is not None
    
    async def get_all_interactions(self, limit: int = 10000) -> List[Dict[str, Any]]:
        """Get all user interactions for collaborative filtering"""
        interactions_collection = self.collection.database["user_interactions"]
        cursor = interactions_collection.find({}).sort("timestamp", DESCENDING).limit(limit)
        return await cursor.to_list(length=limit)


class VideoFeaturesModel:
    """Model for video features used in recommendations"""
    
    def __init__(self, mongodb: MongoDB):
        self.collection = mongodb.get_collection("video_features")
    
    async def store_features(self, video_id: str, user_id: str, features: Dict[str, Any]) -> bool:
        """Store video features for recommendation system"""
        feature_data = {
            "video_id": video_id,
            "user_id": user_id,
            "features": features,
            "created_at": datetime.utcnow()
        }
        
        result = await self.collection.update_one(
            {"video_id": video_id},
            {"$set": feature_data},
            upsert=True
        )
        return result.modified_count > 0 or result.upserted_id is not None
    
    async def get_features(self, video_id: str) -> Optional[Dict[str, Any]]:
        """Get video features"""
        return await self.collection.find_one({"video_id": video_id})
    
    async def get_all_features(self, limit: int = 1000) -> List[Dict[str, Any]]:
        """Get all video features for similarity computation"""
        cursor = self.collection.find({}).limit(limit)
        return await cursor.to_list(length=limit)


# Global database instance
mongodb = MongoDB()