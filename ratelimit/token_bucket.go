package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket represents a token bucket rate limiter
type TokenBucket struct {
	capacity     int           // Maximum number of tokens
	tokens       int           // Current number of tokens
	refillRate   int           // Tokens added per second
	lastRefill   time.Time     // Last time tokens were refilled
	mutex        sync.Mutex    // Protects the bucket state
	refillTicker *time.Ticker  // Periodic refill ticker
	stopChan     chan struct{} // Channel to stop the ticker
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	tb := &TokenBucket{
		capacity:   capacity,
		tokens:     capacity, // Start with full bucket
		refillRate: refillRate,
		lastRefill: time.Now(),
		stopChan:   make(chan struct{}),
	}

	// Start the refill ticker
	tb.refillTicker = time.NewTicker(time.Second)
	go tb.refillLoop()

	return tb
}

// refillLoop continuously refills tokens
func (tb *TokenBucket) refillLoop() {
	for {
		select {
		case <-tb.refillTicker.C:
			tb.refill()
		case <-tb.stopChan:
			tb.refillTicker.Stop()
			return
		}
	}
}

// refill adds tokens to the bucket based on elapsed time
func (tb *TokenBucket) refill() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * tb.refillRate

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}
}

// TryConsume attempts to consume a token from the bucket
func (tb *TokenBucket) TryConsume(tokens int) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// Refill tokens before checking (inline refill to avoid deadlock)
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * tb.refillRate

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	if tb.tokens >= tokens {
		tb.tokens -= tokens
		return true
	}
	return false
}

// GetTokens returns the current number of tokens
func (tb *TokenBucket) GetTokens() int {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// Inline refill to avoid deadlock
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * tb.refillRate

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	return tb.tokens
}

// GetCapacity returns the bucket capacity
func (tb *TokenBucket) GetCapacity() int {
	return tb.capacity
}

// GetRefillRate returns the refill rate
func (tb *TokenBucket) GetRefillRate() int {
	return tb.refillRate
}

// Stop stops the token bucket ticker
func (tb *TokenBucket) Stop() {
	close(tb.stopChan)
}

// RateLimitConfig represents configuration for rate limiting
type RateLimitConfig struct {
	Capacity   int           `json:"capacity"`    // Maximum tokens
	RefillRate int           `json:"refill_rate"` // Tokens per second
	Window     time.Duration `json:"window"`      // Time window for rate limiting
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Capacity:   100,         // 100 requests
		RefillRate: 10,          // 10 requests per second
		Window:     time.Minute, // 1 minute window
	}
}

// RateLimiter manages multiple token buckets
type RateLimiter struct {
	buckets map[string]*TokenBucket
	mutex   sync.RWMutex
	config  *RateLimitConfig
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	return &RateLimiter{
		buckets: make(map[string]*TokenBucket),
		config:  config,
	}
}

// GetBucket gets or creates a token bucket for a key
func (rl *RateLimiter) GetBucket(key string) *TokenBucket {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = NewTokenBucket(rl.config.Capacity, rl.config.RefillRate)
		rl.buckets[key] = bucket
	}

	return bucket
}

// Allow checks if a request is allowed for the given key
func (rl *RateLimiter) Allow(key string, tokens int) bool {
	bucket := rl.GetBucket(key)
	return bucket.TryConsume(tokens)
}

// GetStatus returns the current status of a bucket
func (rl *RateLimiter) GetStatus(key string) (tokens int, capacity int, refillRate int) {
	bucket := rl.GetBucket(key)
	return bucket.GetTokens(), bucket.GetCapacity(), bucket.GetRefillRate()
}

// Cleanup removes old buckets (for memory management)
func (rl *RateLimiter) Cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// This is a simple cleanup - in production, you might want more sophisticated logic
	// For now, we'll keep all buckets as they might be used again
}

// Stop stops all token buckets
func (rl *RateLimiter) Stop() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	for _, bucket := range rl.buckets {
		bucket.Stop()
	}
}

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed    bool          `json:"allowed"`
	Remaining  int           `json:"remaining"`
	ResetTime  time.Time     `json:"reset_time"`
	RetryAfter time.Duration `json:"retry_after"`
}

// CheckRateLimit checks rate limiting and returns detailed result
func (rl *RateLimiter) CheckRateLimit(key string, tokens int) *RateLimitResult {
	bucket := rl.GetBucket(key)
	allowed := bucket.TryConsume(tokens)
	remaining := bucket.GetTokens()

	// Calculate reset time (when bucket will be full again)
	capacity := bucket.GetCapacity()
	refillRate := bucket.GetRefillRate()

	var resetTime time.Time
	var retryAfter time.Duration

	if !allowed {
		// Calculate when enough tokens will be available
		tokensNeeded := tokens - remaining
		secondsToWait := float64(tokensNeeded) / float64(refillRate)
		retryAfter = time.Duration(secondsToWait) * time.Second
		resetTime = time.Now().Add(retryAfter)
	} else {
		// Calculate when bucket will be full
		secondsToFull := float64(capacity-remaining) / float64(refillRate)
		resetTime = time.Now().Add(time.Duration(secondsToFull) * time.Second)
	}

	return &RateLimitResult{
		Allowed:    allowed,
		Remaining:  remaining,
		ResetTime:  resetTime,
		RetryAfter: retryAfter,
	}
}
