"""Configuration management for analytics service"""
import os
from typing import Optional


class Config:
    """Application configuration"""
    
    # Server configuration
    HOST: str = os.getenv("HOST", "0.0.0.0")
    PORT: int = int(os.getenv("PORT", "8000"))
    DEBUG: bool = os.getenv("DEBUG", "false").lower() == "true"
    
    # MongoDB configuration
    MONGODB_URL: str = os.getenv("MONGODB_URI", "mongodb://localhost:27017")
    MONGODB_DATABASE: str = os.getenv("MONGODB_DATABASE", "video_converter")
    
    # Redis configuration
    REDIS_URL: str = os.getenv("REDIS_URL", "redis://localhost:6379")
    
    # RabbitMQ configuration
    RABBITMQ_URL: str = os.getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
    RABBITMQ_QUEUE: str = os.getenv("RABBITMQ_QUEUE", "video_analysis")
    RABBITMQ_EXCHANGE: str = os.getenv("RABBITMQ_EXCHANGE", "video_processing")
    
    # ML Models configuration
    MODELS_PATH: str = os.getenv("MODELS_PATH", "/app/models")
    CONTENT_SAFETY_MODEL: str = os.getenv("CONTENT_SAFETY_MODEL", "unitary/toxic-bert")
    
    # File storage configuration
    TEMP_DIR: str = os.getenv("TEMP_DIR", "/tmp/analytics")
    THUMBNAILS_DIR: str = os.getenv("THUMBNAILS_DIR", "/app/thumbnails")
    
    # Processing configuration
    MAX_CONCURRENT_JOBS: int = int(os.getenv("MAX_CONCURRENT_JOBS", "4"))
    THUMBNAIL_COUNT: int = int(os.getenv("THUMBNAIL_COUNT", "5"))
    THUMBNAIL_WIDTH: int = int(os.getenv("THUMBNAIL_WIDTH", "320"))
    THUMBNAIL_HEIGHT: int = int(os.getenv("THUMBNAIL_HEIGHT", "240"))

    # CORS configuration
    CORS_ORIGINS: str = os.getenv("CORS_ORIGINS", "http://localhost:3000,http://localhost:3001")


config = Config()