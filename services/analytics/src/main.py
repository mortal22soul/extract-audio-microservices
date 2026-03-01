"""
Analytics Service - ML-powered video analysis service
"""
import asyncio
import logging
import uvicorn
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from analytics.config import config
from analytics.database import mongodb
from analytics.messaging import consumer, publisher
from analytics.models import video_quality_model
from analytics.services import AnalyticsService
from analytics.api import router

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan manager"""
    # Startup
    logger.info("Starting Analytics Service...")
    
    try:
        # Connect to databases and messaging
        await mongodb.connect()
        await consumer.connect()
        await publisher.connect()
        
        # Load ML models
        await video_quality_model.load_model()
        
        # Initialize analytics service
        analytics_service = AnalyticsService()
        
        # Register message handlers
        consumer.register_handler("video_analysis", analytics_service.handle_video_analysis)
        consumer.register_handler("quality_analysis", analytics_service.handle_quality_analysis)
        consumer.register_handler("thumbnail_generation", analytics_service.handle_thumbnail_generation)
        consumer.register_handler("safety_check", analytics_service.handle_safety_check)
        consumer.register_handler("user_interaction", analytics_service.handle_user_interaction)
        
        # Start consuming messages in background
        asyncio.create_task(consumer.start_consuming())
        
        logger.info("Analytics Service started successfully")
        
        yield
        
    except Exception as e:
        logger.error(f"Failed to start Analytics Service: {e}")
        raise
    
    # Shutdown
    logger.info("Shutting down Analytics Service...")
    
    try:
        await consumer.stop_consuming()
        await consumer.disconnect()
        await publisher.disconnect()
        await mongodb.disconnect()
        
        # Cleanup ML models
        video_quality_model.cleanup()
        
        logger.info("Analytics Service shut down successfully")
        
    except Exception as e:
        logger.error(f"Error during shutdown: {e}")


app = FastAPI(
    title="Analytics Service",
    description="ML-powered video analysis service",
    version="0.1.0",
    lifespan=lifespan
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include API router
app.include_router(router)

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "analytics",
        "models_loaded": {
            "video_quality": video_quality_model.is_loaded
        },
        "connections": {
            "mongodb": mongodb.client is not None,
            "rabbitmq_consumer": consumer.connection is not None,
            "rabbitmq_publisher": publisher.connection is not None
        }
    }

@app.get("/")
async def root():
    """Root endpoint"""
    return {"message": "Analytics Service is running"}

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host=config.HOST,
        port=config.PORT,
        reload=config.DEBUG
    )