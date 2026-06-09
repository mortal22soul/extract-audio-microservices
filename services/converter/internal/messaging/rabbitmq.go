package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  RabbitMQConfig
}

type RabbitMQConfig struct {
	URL                  string
	ConversionQueue      string
	NotificationExchange string
}

type ConversionMessage struct {
	JobID     string `json:"job_id"`
	UserID    string `json:"user_id"`
	VideoID   string `json:"video_id"`
	VideoPath string `json:"video_path"`
	Filename  string `json:"filename"`
}

type NotificationMessage struct {
	Type    string `json:"type"`
	UserID  string `json:"user_id"`
	JobID   string `json:"job_id"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

func NewRabbitMQClient(config RabbitMQConfig) (*RabbitMQClient, error) {
	conn, err := amqp091.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	client := &RabbitMQClient{
		conn:    conn,
		channel: channel,
		config:  config,
	}

	// Setup queues and exchanges
	if err := client.setupInfrastructure(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to setup RabbitMQ infrastructure: %w", err)
	}

	log.Printf("Connected to RabbitMQ: %s", config.URL)
	return client, nil
}

func (r *RabbitMQClient) setupInfrastructure() error {
	// Declare conversion queue
	_, err := r.channel.QueueDeclare(
		r.config.ConversionQueue, // name
		true,                     // durable
		false,                    // delete when unused
		false,                    // exclusive
		false,                    // no-wait
		amqp091.Table{
			"x-dead-letter-exchange": r.config.ConversionQueue + ".dlx",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare conversion queue: %w", err)
	}

	// Declare dead letter queue
	_, err = r.channel.QueueDeclare(
		r.config.ConversionQueue+".dlq", // name
		true,                            // durable
		false,                           // delete when unused
		false,                           // exclusive
		false,                           // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter queue: %w", err)
	}

	// Declare dead letter exchange
	err = r.channel.ExchangeDeclare(
		r.config.ConversionQueue+".dlx", // name
		"direct",                        // type
		true,                            // durable
		false,                           // auto-deleted
		false,                           // internal
		false,                           // no-wait
		nil,                             // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter exchange: %w", err)
	}

	// Bind dead letter queue to exchange
	err = r.channel.QueueBind(
		r.config.ConversionQueue+".dlq", // queue name
		r.config.ConversionQueue,        // routing key
		r.config.ConversionQueue+".dlx", // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind dead letter queue: %w", err)
	}

	// Declare notification exchange
	err = r.channel.ExchangeDeclare(
		r.config.NotificationExchange, // name
		"topic",                       // type
		true,                          // durable
		false,                         // auto-deleted
		false,                         // internal
		false,                         // no-wait
		nil,                           // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare notification exchange: %w", err)
	}

	return nil
}



func (r *RabbitMQClient) PublishNotification(ctx context.Context, notification NotificationMessage) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	routingKey := fmt.Sprintf("notification.%s", notification.Type)

	err = r.channel.PublishWithContext(
		ctx,
		r.config.NotificationExchange, // exchange
		routingKey,                     // routing key
		false,                          // mandatory
		false,                          // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish notification: %w", err)
	}

	log.Printf("Published notification: %s for user %s", notification.Type, notification.UserID)
	return nil
}

func (r *RabbitMQClient) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *RabbitMQClient) GetConnection() *amqp091.Connection {
	return r.conn
}