package util

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	rate     int       // tokens per second
	capacity int       // bucket capacity
	tokens   int       // current tokens
	lastTime time.Time // last refill time
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{
		rate:     rps,
		capacity: rps,
		tokens:   rps,
		lastTime: time.Now(),
	}
}

// Allow checks if a request is allowed and consumes a token if so
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastTime)

	// Add tokens based on elapsed time
	tokensToAdd := int(elapsed.Seconds() * float64(rl.rate))
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.capacity {
			rl.tokens = rl.capacity
		}
		rl.lastTime = now
	}

	// Check if we have tokens available
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait() {
	for !rl.Allow() {
		time.Sleep(time.Millisecond * 10)
	}
}
