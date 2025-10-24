"""RabbitMQ messaging for analytics service"""
import asyncio
import json
import logging
from typing import Callable, Dict, Any, Optional
from aio_pika import connect_robust, Message, IncomingMessage, ExchangeType
from aio_pika.abc import AbstractConnection, AbstractChannel, AbstractQueue, AbstractExchange

from .config import config

logger = logging.getLogger(__name__)


class RabbitMQConsumer:
    """RabbitMQ consumer for video analysis jobs"""
    
    def __init__(self):
        self.connection: Optional[AbstractConnection] = None
        self.channel: Optional[AbstractChannel] = None
        self.queue: Optional[AbstractQueue] = None
        self.exchange: Optional[AbstractExchange] = None
        self.message_handlers: Dict[str, Callable] = {}
        self._consuming = False
    
    async def connect(self) -> None:
        """Connect to RabbitMQ"""
        try:
            self.connection = await connect_robust(config.RABBITMQ_URL)
            self.channel = await self.connection.channel()
            
            # Set QoS to process one message at a time
            await self.channel.set_qos(prefetch_count=1)
            
            # Declare exchange
            self.exchange = await self.channel.declare_exchange(
                config.RABBITMQ_EXCHANGE,
                ExchangeType.TOPIC,
                durable=True
            )
            
            # Declare queue
            self.queue = await self.channel.declare_queue(
                config.RABBITMQ_QUEUE,
                durable=True
            )
            
            # Bind queue to exchange with routing keys
            await self.queue.bind(self.exchange, "video.analyze")
            await self.queue.bind(self.exchange, "video.quality")
            await self.queue.bind(self.exchange, "video.thumbnails")
            await self.queue.bind(self.exchange, "video.safety")
            
            logger.info("Connected to RabbitMQ successfully")
            
        except Exception as e:
            logger.error(f"Failed to connect to RabbitMQ: {e}")
            raise
    
    async def disconnect(self) -> None:
        """Disconnect from RabbitMQ"""
        self._consuming = False
        
        if self.connection and not self.connection.is_closed:
            await self.connection.close()
            logger.info("Disconnected from RabbitMQ")
    
    def register_handler(self, message_type: str, handler: Callable) -> None:
        """Register message handler for specific message type"""
        self.message_handlers[message_type] = handler
        logger.info(f"Registered handler for message type: {message_type}")
    
    async def start_consuming(self) -> None:
        """Start consuming messages"""
        if not self.queue:
            raise RuntimeError("Not connected to RabbitMQ")
        
        self._consuming = True
        logger.info("Starting to consume messages...")
        
        async def process_message(message: IncomingMessage) -> None:
            """Process incoming message"""
            async with message.process():
                try:
                    # Parse message
                    body = json.loads(message.body.decode())
                    message_type = body.get("type")
                    
                    logger.info(f"Received message type: {message_type}")
                    
                    # Find and execute handler
                    if message_type in self.message_handlers:
                        handler = self.message_handlers[message_type]
                        await handler(body)
                        logger.info(f"Successfully processed message: {message_type}")
                    else:
                        logger.warning(f"No handler found for message type: {message_type}")
                
                except json.JSONDecodeError as e:
                    logger.error(f"Failed to parse message JSON: {e}")
                except Exception as e:
                    logger.error(f"Error processing message: {e}")
                    # Message will be rejected and potentially requeued
                    raise
        
        # Start consuming
        await self.queue.consume(process_message)
        
        # Keep consuming until stopped
        while self._consuming:
            await asyncio.sleep(1)
    
    async def stop_consuming(self) -> None:
        """Stop consuming messages"""
        self._consuming = False
        logger.info("Stopped consuming messages")


class RabbitMQPublisher:
    """RabbitMQ publisher for sending analysis results"""
    
    def __init__(self):
        self.connection: Optional[AbstractConnection] = None
        self.channel: Optional[AbstractChannel] = None
        self.exchange: Optional[AbstractExchange] = None
    
    async def connect(self) -> None:
        """Connect to RabbitMQ"""
        try:
            self.connection = await connect_robust(config.RABBITMQ_URL)
            self.channel = await self.connection.channel()
            
            # Declare exchange
            self.exchange = await self.channel.declare_exchange(
                config.RABBITMQ_EXCHANGE,
                ExchangeType.TOPIC,
                durable=True
            )
            
            logger.info("Publisher connected to RabbitMQ successfully")
            
        except Exception as e:
            logger.error(f"Failed to connect publisher to RabbitMQ: {e}")
            raise
    
    async def disconnect(self) -> None:
        """Disconnect from RabbitMQ"""
        if self.connection and not self.connection.is_closed:
            await self.connection.close()
            logger.info("Publisher disconnected from RabbitMQ")
    
    async def publish_analysis_complete(self, video_id: str, user_id: str, analysis_data: Dict[str, Any]) -> None:
        """Publish analysis completion message"""
        message_data = {
            "type": "analysis_complete",
            "video_id": video_id,
            "user_id": user_id,
            "analysis": analysis_data,
            "timestamp": asyncio.get_event_loop().time()
        }
        
        await self._publish_message("video.analysis.complete", message_data)
    
    async def publish_analysis_error(self, video_id: str, user_id: str, error: str) -> None:
        """Publish analysis error message"""
        message_data = {
            "type": "analysis_error",
            "video_id": video_id,
            "user_id": user_id,
            "error": error,
            "timestamp": asyncio.get_event_loop().time()
        }
        
        await self._publish_message("video.analysis.error", message_data)
    
    async def publish_content_blocked(self, video_id: str, user_id: str, reason: str) -> None:
        """Publish content blocked message"""
        message_data = {
            "type": "content_blocked",
            "video_id": video_id,
            "user_id": user_id,
            "reason": reason,
            "timestamp": asyncio.get_event_loop().time()
        }
        
        await self._publish_message("video.moderation.blocked", message_data)
    
    async def publish_content_flagged_for_review(self, video_id: str, user_id: str, reason: str) -> None:
        """Publish content flagged for review message"""
        message_data = {
            "type": "content_flagged_for_review",
            "video_id": video_id,
            "user_id": user_id,
            "reason": reason,
            "timestamp": asyncio.get_event_loop().time()
        }
        
        await self._publish_message("video.moderation.review", message_data)
    
    async def publish_content_warning(self, video_id: str, user_id: str, reason: str) -> None:
        """Publish content warning message"""
        message_data = {
            "type": "content_warning",
            "video_id": video_id,
            "user_id": user_id,
            "reason": reason,
            "timestamp": asyncio.get_event_loop().time()
        }
        
        await self._publish_message("video.moderation.warning", message_data)
    
    async def publish_content_approved(self, video_id: str, user_id: str) -> None:
        """Publish content approved message"""
        message_data = {
            "type": "content_approved",
            "video_id": video_id,
            "user_id": user_id,
            "timestamp": asyncio.get_event_loop().time()
        }
        
        await self._publish_message("video.moderation.approved", message_data)
    
    async def _publish_message(self, routing_key: str, message_data: Dict[str, Any]) -> None:
        """Publish message to exchange"""
        if not self.exchange:
            raise RuntimeError("Publisher not connected to RabbitMQ")
        
        message_body = json.dumps(message_data).encode()
        message = Message(
            message_body,
            delivery_mode=2,  # Make message persistent
            content_type="application/json"
        )
        
        await self.exchange.publish(message, routing_key=routing_key)
        logger.info(f"Published message with routing key: {routing_key}")


# Global messaging instances
consumer = RabbitMQConsumer()
publisher = RabbitMQPublisher()