package auth

import (
	"sync"
	"time"
)

// BlacklistedToken represents a blacklisted token
type BlacklistedToken struct {
	TokenHash string
	ExpiresAt time.Time
}

// TokenBlacklist manages blacklisted tokens
type TokenBlacklist struct {
	tokens map[string]time.Time // token hash -> expiry time
	mutex  sync.RWMutex
}

// NewTokenBlacklist creates a new token blacklist
func NewTokenBlacklist() *TokenBlacklist {
	bl := &TokenBlacklist{
		tokens: make(map[string]time.Time),
	}

	// Start cleanup goroutine
	go bl.cleanup()

	return bl
}

// BlacklistToken adds a token to the blacklist
func (bl *TokenBlacklist) BlacklistToken(tokenHash string, expiresAt time.Time) {
	bl.mutex.Lock()
	defer bl.mutex.Unlock()

	bl.tokens[tokenHash] = expiresAt
}

// IsBlacklisted checks if a token is blacklisted
func (bl *TokenBlacklist) IsBlacklisted(tokenHash string) bool {
	bl.mutex.RLock()
	defer bl.mutex.RUnlock()

	expiresAt, exists := bl.tokens[tokenHash]
	if !exists {
		return false
	}

	// If token has expired, it's effectively not blacklisted anymore
	if time.Now().After(expiresAt) {
		return false
	}

	return true
}

// cleanup removes expired tokens from the blacklist
func (bl *TokenBlacklist) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		bl.mutex.Lock()
		now := time.Now()

		for tokenHash, expiresAt := range bl.tokens {
			if now.After(expiresAt) {
				delete(bl.tokens, tokenHash)
			}
		}

		bl.mutex.Unlock()
	}
}

// GetBlacklistedCount returns the number of blacklisted tokens
func (bl *TokenBlacklist) GetBlacklistedCount() int {
	bl.mutex.RLock()
	defer bl.mutex.RUnlock()

	return len(bl.tokens)
}