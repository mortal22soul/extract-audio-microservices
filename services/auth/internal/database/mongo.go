package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Config holds database configuration
type Config struct {
	URI    string
	DBName string
}

// NewConfig creates a new database configuration from environment variables
func NewConfig() *Config {
	return &Config{
		URI:    getEnv("MONGODB_URI", "mongodb://localhost:27017/auth"),
		DBName: getEnv("MONGODB_NAME", "auth"),
	}
}

// Connect establishes a connection to the MongoDB database
func Connect(config *Config) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	// Ping the database
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	log.Println("Successfully connected to MongoDB database")
	return client, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
