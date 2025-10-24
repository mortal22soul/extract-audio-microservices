package utils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "github.com/lib/pq"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	PostgresURL    string
	MongoURL       string
	RedisURL       string
	RabbitMQURL    string
	AuthServiceURL string
	AnalyticsURL   string
	GatewayURL     string
	RealtimeURL    string
	JWTSecret      string
}

// GetTestConfig returns test configuration from environment variables
func GetTestConfig() *TestConfig {
	return &TestConfig{
		PostgresURL:    getEnv("TEST_POSTGRES_URL", "postgres://test_user:test_pass@localhost:5433/video_converter_test?sslmode=disable"),
		MongoURL:       getEnv("TEST_MONGO_URL", "mongodb://test_user:test_pass@localhost:27018/video_converter_test"),
		RedisURL:       getEnv("TEST_REDIS_URL", "redis://:test_pass@localhost:6380"),
		RabbitMQURL:    getEnv("TEST_RABBITMQ_URL", "amqp://test_user:test_pass@localhost:5673/"),
		AuthServiceURL: getEnv("TEST_AUTH_SERVICE_URL", "localhost:50051"),
		AnalyticsURL:   getEnv("TEST_ANALYTICS_URL", "localhost:50052"),
		GatewayURL:     getEnv("TEST_GATEWAY_URL", "http://localhost:8080"),
		RealtimeURL:    getEnv("TEST_REALTIME_URL", "http://localhost:3001"),
		JWTSecret:      getEnv("TEST_JWT_SECRET", "test_jwt_secret_key_for_integration_tests"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestUser represents a test user
type TestUser struct {
	ID        string
	Email     string
	Password  string
	FirstName string
	LastName  string
	Token     string
}

// CreateTestUser creates a test user and returns user data with JWT token
func CreateTestUser(t *testing.T, config *TestConfig) *TestUser {
	user := &TestUser{
		Email:     fmt.Sprintf("test_%d@example.com", time.Now().UnixNano()),
		Password:  "TestPassword123!",
		FirstName: "Test",
		LastName:  "User",
	}

	// Generate JWT token for the user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "test_user_id",
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	user.Token = tokenString
	user.ID = "test_user_id"

	return user
}

// DatabaseConnections holds database connections for tests
type DatabaseConnections struct {
	Postgres *sql.DB
	MongoDB  *mongo.Client
	Redis    *redis.Client
	RabbitMQ *amqp.Connection
}

// SetupDatabases establishes connections to all test databases
func SetupDatabases(t *testing.T, config *TestConfig) *DatabaseConnections {
	ctx := context.Background()

	// PostgreSQL connection
	postgres, err := sql.Open("postgres", config.PostgresURL)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	if err := postgres.PingContext(ctx); err != nil {
		t.Fatalf("Failed to ping PostgreSQL: %v", err)
	}

	// MongoDB connection
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoURL))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := mongoClient.Ping(ctx, nil); err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// Redis connection
	redisOpts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		t.Fatalf("Failed to parse Redis URL: %v", err)
	}

	redisClient := redis.NewClient(redisOpts)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to ping Redis: %v", err)
	}

	// RabbitMQ connection
	rabbitConn, err := amqp.Dial(config.RabbitMQURL)
	if err != nil {
		t.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	return &DatabaseConnections{
		Postgres: postgres,
		MongoDB:  mongoClient,
		Redis:    redisClient,
		RabbitMQ: rabbitConn,
	}
}

// CleanupDatabases closes all database connections and cleans up test data
func (db *DatabaseConnections) CleanupDatabases(t *testing.T) {
	ctx := context.Background()

	// Clean PostgreSQL
	if db.Postgres != nil {
		_, err := db.Postgres.ExecContext(ctx, "TRUNCATE TABLE users, user_sessions CASCADE")
		if err != nil {
			log.Printf("Failed to clean PostgreSQL: %v", err)
		}
		db.Postgres.Close()
	}

	// Clean MongoDB
	if db.MongoDB != nil {
		database := db.MongoDB.Database("video_converter_test")
		collections := []string{"videos", "conversion_jobs", "analytics"}
		for _, collection := range collections {
			err := database.Collection(collection).Drop(ctx)
			if err != nil {
				log.Printf("Failed to drop MongoDB collection %s: %v", collection, err)
			}
		}
		db.MongoDB.Disconnect(ctx)
	}

	// Clean Redis
	if db.Redis != nil {
		err := db.Redis.FlushDB(ctx).Err()
		if err != nil {
			log.Printf("Failed to flush Redis: %v", err)
		}
		db.Redis.Close()
	}

	// Close RabbitMQ
	if db.RabbitMQ != nil {
		db.RabbitMQ.Close()
	}
}

// GRPCConnections holds gRPC client connections
type GRPCConnections struct {
	AuthConn      *grpc.ClientConn
	AnalyticsConn *grpc.ClientConn
}

// SetupGRPCConnections establishes gRPC connections to services
func SetupGRPCConnections(t *testing.T, config *TestConfig) *GRPCConnections {
	// Auth service connection
	authConn, err := grpc.Dial(config.AuthServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to auth service: %v", err)
	}

	// Analytics service connection
	analyticsConn, err := grpc.Dial(config.AnalyticsURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to analytics service: %v", err)
	}

	return &GRPCConnections{
		AuthConn:      authConn,
		AnalyticsConn: analyticsConn,
	}
}

// CleanupGRPCConnections closes all gRPC connections
func (g *GRPCConnections) CleanupGRPCConnections() {
	if g.AuthConn != nil {
		g.AuthConn.Close()
	}
	if g.AnalyticsConn != nil {
		g.AnalyticsConn.Close()
	}
}

// WaitForServices waits for all services to be ready
func WaitForServices(t *testing.T, config *TestConfig) {
	maxRetries := 30
	retryInterval := 2 * time.Second

	services := map[string]string{
		"Gateway":   config.GatewayURL + "/health",
		"Realtime":  config.RealtimeURL + "/health",
	}

	for serviceName, healthURL := range services {
		for i := 0; i < maxRetries; i++ {
			// Simple HTTP health check would go here
			// For now, just wait a bit
			time.Sleep(retryInterval)
			log.Printf("Waiting for %s service to be ready... (%d/%d)", serviceName, i+1, maxRetries)
		}
	}

	// Wait for gRPC services
	grpcServices := map[string]string{
		"Auth":      config.AuthServiceURL,
		"Analytics": config.AnalyticsURL,
	}

	for serviceName, address := range grpcServices {
		for i := 0; i < maxRetries; i++ {
			conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err == nil {
				conn.Close()
				log.Printf("%s gRPC service is ready", serviceName)
				break
			}
			if i == maxRetries-1 {
				t.Fatalf("Failed to connect to %s gRPC service after %d retries", serviceName, maxRetries)
			}
			time.Sleep(retryInterval)
		}
	}
}

// CreateTestVideoFile creates a test video file for upload tests
func CreateTestVideoFile(t *testing.T) []byte {
	// This would create a minimal test video file
	// For now, return dummy data
	return []byte("dummy video content for testing")
}