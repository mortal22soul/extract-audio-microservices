package monitoring

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/go-redis/redis/v8"
)

type HealthChecker struct {
	mongoClient  *mongo.Client
	redisClient  *redis.Client
	rabbitConn   *amqp091.Connection
}

type HealthStatus struct {
	Service   string    `json:"service"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type SystemHealth struct {
	Overall   string         `json:"overall"`
	Services  []HealthStatus `json:"services"`
	Timestamp time.Time      `json:"timestamp"`
}

func NewHealthChecker(mongoClient *mongo.Client, redisClient *redis.Client, rabbitConn *amqp091.Connection) *HealthChecker {
	return &HealthChecker{
		mongoClient: mongoClient,
		redisClient: redisClient,
		rabbitConn:  rabbitConn,
	}
}

func (h *HealthChecker) CheckHealth(ctx context.Context) SystemHealth {
	var services []HealthStatus
	overallHealthy := true

	// Check MongoDB
	mongoStatus := h.checkMongoDB(ctx)
	services = append(services, mongoStatus)
	if mongoStatus.Status != "healthy" {
		overallHealthy = false
	}

	// Check Redis
	redisStatus := h.checkRedis(ctx)
	services = append(services, redisStatus)
	if redisStatus.Status != "healthy" {
		overallHealthy = false
	}

	// Check RabbitMQ
	rabbitStatus := h.checkRabbitMQ()
	services = append(services, rabbitStatus)
	if rabbitStatus.Status != "healthy" {
		overallHealthy = false
	}

	overall := "healthy"
	if !overallHealthy {
		overall = "unhealthy"
	}

	return SystemHealth{
		Overall:   overall,
		Services:  services,
		Timestamp: time.Now(),
	}
}

func (h *HealthChecker) checkMongoDB(ctx context.Context) HealthStatus {
	status := HealthStatus{
		Service:   "mongodb",
		Timestamp: time.Now(),
	}

	// Create a context with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.mongoClient.Ping(pingCtx, nil); err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("MongoDB ping failed: %v", err)
		return status
	}

	status.Status = "healthy"
	status.Message = "MongoDB connection is healthy"
	return status
}

func (h *HealthChecker) checkRedis(ctx context.Context) HealthStatus {
	status := HealthStatus{
		Service:   "redis",
		Timestamp: time.Now(),
	}

	// Create a context with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.redisClient.Ping(pingCtx).Err(); err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("Redis ping failed: %v", err)
		return status
	}

	status.Status = "healthy"
	status.Message = "Redis connection is healthy"
	return status
}

func (h *HealthChecker) checkRabbitMQ() HealthStatus {
	status := HealthStatus{
		Service:   "rabbitmq",
		Timestamp: time.Now(),
	}

	if h.rabbitConn == nil || h.rabbitConn.IsClosed() {
		status.Status = "unhealthy"
		status.Message = "RabbitMQ connection is closed"
		return status
	}

	status.Status = "healthy"
	status.Message = "RabbitMQ connection is healthy"
	return status
}

func (h *HealthChecker) StartHealthMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Started health monitoring with interval: %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Health monitoring stopped")
			return
		case <-ticker.C:
			health := h.CheckHealth(ctx)
			
			if health.Overall != "healthy" {
				log.Printf("System health check failed: %+v", health)
				
				// Log individual service issues
				for _, service := range health.Services {
					if service.Status != "healthy" {
						log.Printf("Service %s is unhealthy: %s", service.Service, service.Message)
					}
				}
			} else {
				log.Printf("System health check passed at %v", health.Timestamp)
			}
		}
	}
}

// QueueMetrics provides metrics about queue status
type QueueMetrics struct {
	QueueName     string `json:"queue_name"`
	MessageCount  int    `json:"message_count"`
	ConsumerCount int    `json:"consumer_count"`
	Timestamp     time.Time `json:"timestamp"`
}

func (h *HealthChecker) GetQueueMetrics(queueName string) (*QueueMetrics, error) {
	if h.rabbitConn == nil || h.rabbitConn.IsClosed() {
		return nil, fmt.Errorf("RabbitMQ connection is not available")
	}

	channel, err := h.rabbitConn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	defer channel.Close()

	queue, err := channel.QueueInspect(queueName)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect queue: %w", err)
	}

	return &QueueMetrics{
		QueueName:     queueName,
		MessageCount:  queue.Messages,
		ConsumerCount: queue.Consumers,
		Timestamp:     time.Now(),
	}, nil
}