package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/video-converter/notification/internal/config"
	"github.com/video-converter/notification/internal/models"
	"github.com/video-converter/notification/internal/service"
)

type Consumer struct {
	config             *config.Config
	conn               *amqp091.Connection
	channel            *amqp091.Channel
	notificationService *service.NotificationService
	stopCh             chan struct{}
	done               chan struct{}
}

func New(cfg *config.Config, notificationService *service.NotificationService) *Consumer {
	return &Consumer{
		config:              cfg,
		notificationService: notificationService,
		stopCh:              make(chan struct{}),
		done:                make(chan struct{}),
	}
}

func (c *Consumer) Start() error {
	log.Println("Starting notification consumer...")
	
	// Connect to RabbitMQ
	if err := c.connect(); err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	
	// Declare queue
	if err := c.declareQueue(); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	
	// Start consuming messages
	go c.consume()
	
	log.Println("Notification consumer started successfully")
	return nil
}

func (c *Consumer) Stop() {
	log.Println("Stopping notification consumer...")
	close(c.stopCh)
	<-c.done
	
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	
	log.Println("Notification consumer stopped")
}

func (c *Consumer) connect() error {
	var err error
	
	// Retry connection with exponential backoff
	for attempt := 1; attempt <= 5; attempt++ {
		c.conn, err = amqp091.Dial(c.config.RabbitMQ.URL)
		if err == nil {
			c.channel, err = c.conn.Channel()
			if err == nil {
				log.Printf("Connected to RabbitMQ on attempt %d", attempt)
				return nil
			}
		}
		
		log.Printf("Failed to connect to RabbitMQ (attempt %d): %v", attempt, err)
		if attempt < 5 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	
	return fmt.Errorf("failed to connect after 5 attempts: %w", err)
}

func (c *Consumer) declareQueue() error {
	// Declare the notification queue
	_, err := c.channel.QueueDeclare(
		c.config.RabbitMQ.QueueName, // name
		true,                        // durable
		false,                       // delete when unused
		false,                       // exclusive
		false,                       // no-wait
		amqp091.Table{
			"x-dead-letter-exchange": "notifications.dlx",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	
	// Declare dead letter exchange
	err = c.channel.ExchangeDeclare(
		"notifications.dlx", // name
		"direct",           // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter exchange: %w", err)
	}
	
	// Declare dead letter queue
	_, err = c.channel.QueueDeclare(
		"notifications.dlq", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter queue: %w", err)
	}
	
	// Bind dead letter queue to exchange
	err = c.channel.QueueBind(
		"notifications.dlq", // queue name
		"",                  // routing key
		"notifications.dlx", // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind dead letter queue: %w", err)
	}
	
	// Set QoS to process one message at a time
	err = c.channel.Qos(1, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}
	
	return nil
}

func (c *Consumer) consume() {
	defer close(c.done)
	
	// Start consuming messages
	msgs, err := c.channel.Consume(
		c.config.RabbitMQ.QueueName, // queue
		"notification-consumer",      // consumer
		false,                       // auto-ack
		false,                       // exclusive
		false,                       // no-local
		false,                       // no-wait
		nil,                         // args
	)
	if err != nil {
		log.Printf("Failed to register consumer: %v", err)
		return
	}
	
	for {
		select {
		case <-c.stopCh:
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed, attempting to reconnect...")
				if err := c.reconnect(); err != nil {
					log.Printf("Failed to reconnect: %v", err)
					return
				}
				continue
			}
			
			c.processMessage(msg)
		}
	}
}

func (c *Consumer) processMessage(msg amqp091.Delivery) {
	log.Printf("Received notification message: %s", msg.MessageId)
	
	// Parse the notification message
	var notification models.NotificationMessage
	if err := json.Unmarshal(msg.Body, &notification); err != nil {
		log.Printf("Failed to unmarshal notification message: %v", err)
		msg.Nack(false, false) // Send to dead letter queue
		return
	}
	
	// Process the notification with retry logic
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	err := c.processNotificationWithRetry(ctx, &notification, 3)
	if err != nil {
		log.Printf("Failed to process notification after retries: %v", err)
		msg.Nack(false, false) // Send to dead letter queue
		return
	}
	
	// Acknowledge the message
	if err := msg.Ack(false); err != nil {
		log.Printf("Failed to acknowledge message: %v", err)
	}
	
	log.Printf("Successfully processed notification: %s", notification.ID)
}

func (c *Consumer) processNotificationWithRetry(ctx context.Context, notification *models.NotificationMessage, maxRetries int) error {
	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := c.notificationService.ProcessNotification(ctx, notification)
		if err == nil {
			return nil
		}
		
		lastErr = err
		log.Printf("Notification processing attempt %d failed: %v", attempt, err)
		
		if attempt < maxRetries {
			// Wait before retry with exponential backoff
			backoff := time.Duration(attempt) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}
	}
	
	return fmt.Errorf("failed to process notification after %d attempts: %w", maxRetries, lastErr)
}

func (c *Consumer) reconnect() error {
	log.Println("Attempting to reconnect to RabbitMQ...")
	
	// Close existing connections
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	
	// Reconnect
	if err := c.connect(); err != nil {
		return err
	}
	
	// Redeclare queue
	if err := c.declareQueue(); err != nil {
		return err
	}
	
	log.Println("Successfully reconnected to RabbitMQ")
	return nil
}