package auth

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	rateLimiterPrefix   = "auth:ratelimit:"
	maxFailedAttempts   = 5
	lockoutDuration     = 15 * time.Minute
	failedAttemptExpiry = 30 * time.Minute
)

// RateLimiter tracks failed login attempts using Redis
type RateLimiter struct {
	client *redis.Client
	mu     sync.RWMutex
}

// NewRateLimiter creates a new Redis-backed rate limiter
func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{
		client: client,
	}
}

// RecordAttempt increments the failed attempt counter for an identifier.
// Returns true if the identifier is now blocked.
func (rl *RateLimiter) RecordAttempt(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	ctx := context.Background()
	key := rateLimiterPrefix + identifier

	pipe := rl.client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, failedAttemptExpiry)
	if _, err := pipe.Exec(ctx); err != nil {
		fmt.Printf("Failed to record failed attempt in Redis: %v\n", err)
		return false
	}

	count := incrCmd.Val()
	if count >= int64(maxFailedAttempts) {
		// Enforce lockout TTL
		rl.client.Expire(ctx, key, lockoutDuration)
		return true
	}

	return false
}

// RecordSuccess clears the failed attempt counter (called on successful login)
func (rl *RateLimiter) RecordSuccess(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	ctx := context.Background()
	key := rateLimiterPrefix + identifier

	if err := rl.client.Del(ctx, key).Err(); err != nil {
		fmt.Printf("Failed to reset rate limit in Redis: %v\n", err)
	}
}

// IsBlocked checks if an identifier has exceeded the max failed attempts
func (rl *RateLimiter) IsBlocked(identifier string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	ctx := context.Background()
	key := rateLimiterPrefix + identifier

	val, err := rl.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false
	}
	if err != nil {
		fmt.Printf("Failed to check rate limit in Redis: %v\n", err)
		return false
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return false
	}

	return count >= maxFailedAttempts
}

// GetRemainingAttempts returns how many login attempts remain before lockout
func (rl *RateLimiter) GetRemainingAttempts(identifier string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	ctx := context.Background()
	key := rateLimiterPrefix + identifier

	val, err := rl.client.Get(ctx, key).Result()
	if err != nil {
		return maxFailedAttempts
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return maxFailedAttempts
	}

	remaining := maxFailedAttempts - count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetFailedAttempts returns the current failed attempt count for an identifier
func (rl *RateLimiter) GetFailedAttempts(identifier string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	ctx := context.Background()
	key := rateLimiterPrefix + identifier

	val, err := rl.client.Get(ctx, key).Result()
	if err != nil {
		return 0
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}

	return count
}