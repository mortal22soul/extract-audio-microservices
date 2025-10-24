package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type MessageHandler struct {
	client    *RabbitMQClient
	processor ConversionProcessor
}

type ConversionProcessor interface {
	ProcessConversion(ctx context.Context, msg ConversionMessage) error
}

type MessageMetrics struct {
	ProcessedCount int64
	FailedCount    int64
	RetryCount     int64
	LastProcessed  time.Time
}

func NewMessageHandler(client *RabbitMQClient, processor ConversionProcessor) *MessageHandler {
	return &MessageHandler{
		client:    client,
		processor: processor,
	}
}

func (h *MessageHandler) StartConsuming(ctx context.Context) error {
	log.Println("Starting message consumption...")

	// Set QoS to limit unacknowledged messages per consumer
	err := h.client.channel.Qos(1, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming messages
	msgs, err := h.client.channel.Consume(
		h.client.config.ConversionQueue, // queue
		"converter-worker",              // consumer tag
		false,                           // auto-ack
		false,                           // exclusive
		false,                           // no-local
		false,                           // no-wait
		nil,                             // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("Started consuming from queue: %s", h.client.config.ConversionQueue)

	// Process messages
	for {
		select {
		case <-ctx.Done():
			log.Println("Message consumption stopped by context")
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed")
				return fmt.Errorf("message channel closed")
			}

			h.handleMessage(ctx, msg)
		}
	}
}

func (h *MessageHandler) handleMessage(ctx context.Context, delivery amqp091.Delivery) {
	startTime := time.Now()
	
	// Parse message
	var conversionMsg ConversionMessage
	if err := json.Unmarshal(delivery.Body, &conversionMsg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		h.rejectMessage(delivery, false) // Don't requeue malformed messages
		return
	}

	log.Printf("Processing conversion job: %s (attempt: %d)", 
		conversionMsg.JobID, h.getRetryCount(delivery))

	// Create processing context with timeout
	processCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// Process the conversion
	if err := h.processor.ProcessConversion(processCtx, conversionMsg); err != nil {
		log.Printf("Failed to process conversion job %s: %v", conversionMsg.JobID, err)
		h.handleProcessingError(delivery, conversionMsg, err)
		return
	}

	// Acknowledge successful processing
	if err := delivery.Ack(false); err != nil {
		log.Printf("Failed to acknowledge message: %v", err)
	} else {
		duration := time.Since(startTime)
		log.Printf("Successfully processed conversion job %s in %v", 
			conversionMsg.JobID, duration)
	}
}

func (h *MessageHandler) handleProcessingError(delivery amqp091.Delivery, msg ConversionMessage, err error) {
	retryCount := h.getRetryCount(delivery)
	maxRetries := 3

	if retryCount < maxRetries {
		// Requeue for retry with delay
		log.Printf("Requeuing job %s for retry (attempt %d/%d)", 
			msg.JobID, retryCount+1, maxRetries)
		
		// Add retry delay by publishing to a delay queue (if implemented)
		// For now, just requeue immediately
		h.rejectMessage(delivery, true)
	} else {
		// Max retries reached, send to dead letter queue
		log.Printf("Max retries reached for job %s, sending to dead letter queue", msg.JobID)
		h.rejectMessage(delivery, false)
		
		// Optionally send failure notification
		h.sendFailureNotification(msg, err)
	}
}

func (h *MessageHandler) rejectMessage(delivery amqp091.Delivery, requeue bool) {
	if err := delivery.Nack(false, requeue); err != nil {
		log.Printf("Failed to reject message: %v", err)
	}
}

func (h *MessageHandler) getRetryCount(delivery amqp091.Delivery) int {
	if delivery.Headers == nil {
		return 0
	}
	
	if count, ok := delivery.Headers["x-retry-count"]; ok {
		if retryCount, ok := count.(int32); ok {
			return int(retryCount)
		}
	}
	
	return 0
}

func (h *MessageHandler) sendFailureNotification(msg ConversionMessage, err error) {
	notification := NotificationMessage{
		Type:    "conversion_failed",
		UserID:  msg.UserID,
		JobID:   msg.JobID,
		Subject: "Video Conversion Failed",
		Body:    fmt.Sprintf("Failed to convert video after multiple attempts: %s", err.Error()),
		Data: map[string]interface{}{
			"video_id": msg.VideoID,
			"error":    err.Error(),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.client.PublishNotification(ctx, notification); err != nil {
		log.Printf("Failed to send failure notification: %v", err)
	}
}

// ProcessDeadLetterQueue processes messages from the dead letter queue for manual inspection
func (h *MessageHandler) ProcessDeadLetterQueue(ctx context.Context) error {
	dlqName := h.client.config.ConversionQueue + ".dlq"
	
	msgs, err := h.client.channel.Consume(
		dlqName,           // queue
		"dlq-processor",   // consumer tag
		false,             // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume from dead letter queue: %w", err)
	}

	log.Printf("Started processing dead letter queue: %s", dlqName)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("dead letter queue channel closed")
			}

			h.processDLQMessage(msg)
		}
	}
}

func (h *MessageHandler) processDLQMessage(delivery amqp091.Delivery) {
	var conversionMsg ConversionMessage
	if err := json.Unmarshal(delivery.Body, &conversionMsg); err != nil {
		log.Printf("Failed to unmarshal DLQ message: %v", err)
		delivery.Ack(false)
		return
	}

	log.Printf("Processing DLQ message for job: %s", conversionMsg.JobID)
	
	// Log the failed message for manual inspection
	log.Printf("DLQ Message Details - JobID: %s, UserID: %s, VideoID: %s, Headers: %+v",
		conversionMsg.JobID, conversionMsg.UserID, conversionMsg.VideoID, delivery.Headers)

	// For now, just acknowledge the message
	// In a production system, you might want to:
	// 1. Store in a database for manual review
	// 2. Send to an admin notification system
	// 3. Attempt manual recovery
	delivery.Ack(false)
}