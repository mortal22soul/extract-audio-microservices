package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

type ProgressUpdate struct {
	JobID    string `json:"job_id"`
	UserID   string `json:"user_id"`
	Progress int    `json:"progress"`
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
}

func NewRedisClient(redisURL string) (*RedisClient, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Connected to Redis: %s", redisURL)

	return &RedisClient{
		client: client,
	}, nil
}

func (rc *RedisClient) PublishProgress(ctx context.Context, update ProgressUpdate) error {
	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal progress update: %w", err)
	}

	channel := "conversion:progress"
	if update.Status == "completed" {
		channel = "conversion:complete"
	} else if update.Status == "failed" {
		channel = "conversion:error"
	}

	err = rc.client.Publish(ctx, channel, data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish to Redis: %w", err)
	}

	log.Printf("Published progress update for job %s: %d%%", update.JobID, update.Progress)
	return nil
}

func (rc *RedisClient) SetJobProgress(ctx context.Context, jobID string, progress int) error {
	key := fmt.Sprintf("conversion:%s", jobID)
	data := map[string]interface{}{
		"progress": progress,
		"status":   "processing",
	}

	err := rc.client.HMSet(ctx, key, data).Err()
	if err != nil {
		return fmt.Errorf("failed to set job progress: %w", err)
	}

	// Set expiration for cleanup
	rc.client.Expire(ctx, key, 3600) // 1 hour

	return nil
}

func (rc *RedisClient) Close() error {
	return rc.client.Close()
}

func (rc *RedisClient) GetClient() *redis.Client {
	return rc.client
}