package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenBlacklistPrefix = "auth:blacklist:"

// TokenBlacklist manages blacklisted tokens using Redis
type TokenBlacklist struct {
	client *redis.Client
}

// NewTokenBlacklist creates a new Redis-backed token blacklist
func NewTokenBlacklist(client *redis.Client) *TokenBlacklist {
	return &TokenBlacklist{
		client: client,
	}
}

// BlacklistToken adds a token hash to the blacklist with an expiry
func (bl *TokenBlacklist) BlacklistToken(tokenHash string, expiresAt time.Time) {
	ctx := context.Background()
	key := tokenBlacklistPrefix + tokenHash

	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Token already expired, no need to blacklist
		return
	}

	if err := bl.client.Set(ctx, key, "1", ttl).Err(); err != nil {
		fmt.Printf("Failed to blacklist token in Redis: %v\n", err)
	}
}

// IsBlacklisted checks if a token hash is in the blacklist
func (bl *TokenBlacklist) IsBlacklisted(tokenHash string) bool {
	ctx := context.Background()
	key := tokenBlacklistPrefix + tokenHash

	exists, err := bl.client.Exists(ctx, key).Result()
	if err != nil {
		fmt.Printf("Failed to check token blacklist in Redis: %v\n", err)
		return false
	}

	return exists > 0
}

// GetBlacklistedCount returns the approximate number of blacklisted tokens
func (bl *TokenBlacklist) GetBlacklistedCount() int {
	ctx := context.Background()

	var count int
	var cursor uint64
	for {
		keys, nextCursor, err := bl.client.Scan(ctx, cursor, tokenBlacklistPrefix+"*", 100).Result()
		if err != nil {
			break
		}
		count += len(keys)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return count
}