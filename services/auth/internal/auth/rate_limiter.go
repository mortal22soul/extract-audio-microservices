package auth

import (
	"sync"
	"time"
)

// LoginAttempt represents a login attempt
type LoginAttempt struct {
	Count     int
	LastTry   time.Time
	BlockedAt time.Time
}

// RateLimiter handles rate limiting for login attempts
type RateLimiter struct {
	attempts map[string]*LoginAttempt
	mutex    sync.RWMutex
	
	maxAttempts   int
	blockDuration time.Duration
	windowSize    time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		attempts:      make(map[string]*LoginAttempt),
		maxAttempts:   5,                // Max 5 attempts
		blockDuration: 15 * time.Minute, // Block for 15 minutes
		windowSize:    5 * time.Minute,  // Reset counter every 5 minutes
	}

	// Start cleanup goroutine
	go rl.cleanup()
	
	return rl
}

// IsBlocked checks if an IP/email is currently blocked
func (rl *RateLimiter) IsBlocked(key string) bool {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	attempt, exists := rl.attempts[key]
	if !exists {
		return false
	}

	// Check if still blocked
	if !attempt.BlockedAt.IsZero() && time.Since(attempt.BlockedAt) < rl.blockDuration {
		return true
	}

	return false
}

// RecordAttempt records a failed login attempt
func (rl *RateLimiter) RecordAttempt(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	attempt, exists := rl.attempts[key]

	if !exists {
		rl.attempts[key] = &LoginAttempt{
			Count:   1,
			LastTry: now,
		}
		return false // Not blocked yet
	}

	// Reset counter if window has passed
	if time.Since(attempt.LastTry) > rl.windowSize {
		attempt.Count = 1
		attempt.LastTry = now
		attempt.BlockedAt = time.Time{} // Clear block
		return false
	}

	// Increment attempt count
	attempt.Count++
	attempt.LastTry = now

	// Block if max attempts reached
	if attempt.Count >= rl.maxAttempts {
		attempt.BlockedAt = now
		return true // Now blocked
	}

	return false // Not blocked yet
}

// RecordSuccess records a successful login (resets counter)
func (rl *RateLimiter) RecordSuccess(key string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	delete(rl.attempts, key)
}

// GetRemainingAttempts returns the number of remaining attempts
func (rl *RateLimiter) GetRemainingAttempts(key string) int {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	attempt, exists := rl.attempts[key]
	if !exists {
		return rl.maxAttempts
	}

	// Reset if window has passed
	if time.Since(attempt.LastTry) > rl.windowSize {
		return rl.maxAttempts
	}

	remaining := rl.maxAttempts - attempt.Count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		
		for key, attempt := range rl.attempts {
			// Remove entries that are old and not blocked, or blocks that have expired
			if (time.Since(attempt.LastTry) > rl.windowSize && attempt.BlockedAt.IsZero()) ||
			   (!attempt.BlockedAt.IsZero() && time.Since(attempt.BlockedAt) > rl.blockDuration) {
				delete(rl.attempts, key)
			}
		}
		
		rl.mutex.Unlock()
	}
}