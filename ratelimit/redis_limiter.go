package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter implements distributed rate limiting using Redis
type RedisRateLimiter struct {
	client *redis.Client
	config *RateLimitConfig
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client, config *RateLimitConfig) *RedisRateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	return &RedisRateLimiter{
		client: client,
		config: config,
	}
}

// RedisBucketData represents bucket data stored in Redis
type RedisBucketData struct {
	Tokens     int       `json:"tokens"`
	LastRefill time.Time `json:"last_refill"`
}

// Allow checks if a request is allowed using Redis
func (rl *RedisRateLimiter) Allow(ctx context.Context, key string, tokens int) (*RateLimitResult, error) {
	// Use Lua script for atomic operations
	script := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local refillRate = tonumber(ARGV[2])
		local tokens = tonumber(ARGV[3])
		local now = tonumber(ARGV[4])
		
		-- Get current bucket data
		local data = redis.call('GET', key)
		local bucket
		
		if data then
			bucket = cjson.decode(data)
		else
			bucket = {
				tokens = capacity,
				last_refill = now
			}
		end
		
		-- Calculate tokens to add based on elapsed time
		local elapsed = now - bucket.last_refill
		local tokensToAdd = math.floor(elapsed * refillRate)
		
		-- Refill tokens
		bucket.tokens = bucket.tokens + tokensToAdd
		if bucket.tokens > capacity then
			bucket.tokens = capacity
		end
		bucket.last_refill = now
		
		-- Check if we can consume tokens
		local allowed = false
		if bucket.tokens >= tokens then
			bucket.tokens = bucket.tokens - tokens
			allowed = true
		end
		
		-- Store updated bucket data
		redis.call('SET', key, cjson.encode(bucket), 'EX', 3600)
		
		-- Calculate reset time
		local resetTime
		local retryAfter = 0
		
		if allowed then
			-- When bucket will be full
			resetTime = now + (capacity - bucket.tokens) / refillRate
		else
			-- When enough tokens will be available
			retryAfter = (tokens - bucket.tokens) / refillRate
			resetTime = now + retryAfter
		end
		
		return {allowed, bucket.tokens, resetTime, retryAfter}
	`

	now := time.Now().Unix()
	result, err := rl.client.Eval(ctx, script, []string{key},
		rl.config.Capacity,
		rl.config.RefillRate,
		tokens,
		now).Result()

	if err != nil {
		return nil, fmt.Errorf("redis rate limit check failed: %w", err)
	}

	// Parse result
	results, ok := result.([]interface{})
	if !ok || len(results) != 4 {
		return nil, fmt.Errorf("invalid redis script result")
	}

	allowed, _ := results[0].(int64)
	remaining, _ := results[1].(int64)
	resetTimeUnix, _ := results[2].(float64)
	retryAfterFloat, _ := results[3].(float64)

	return &RateLimitResult{
		Allowed:    allowed == 1,
		Remaining:  int(remaining),
		ResetTime:  time.Unix(int64(resetTimeUnix), 0),
		RetryAfter: time.Duration(retryAfterFloat) * time.Second,
	}, nil
}

// GetStatus gets the current status of a bucket from Redis
func (rl *RedisRateLimiter) GetStatus(ctx context.Context, key string) (int, int, int, error) {
	data, err := rl.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// Bucket doesn't exist, return full capacity
		return rl.config.Capacity, rl.config.Capacity, rl.config.RefillRate, nil
	}
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get bucket status: %w", err)
	}

	var bucket RedisBucketData
	if err := json.Unmarshal([]byte(data), &bucket); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to unmarshal bucket data: %w", err)
	}

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(bucket.LastRefill)
	tokensToAdd := int(elapsed.Seconds()) * rl.config.RefillRate

	tokens := bucket.Tokens + tokensToAdd
	if tokens > rl.config.Capacity {
		tokens = rl.config.Capacity
	}

	return tokens, rl.config.Capacity, rl.config.RefillRate, nil
}

// Reset resets a bucket in Redis
func (rl *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	return rl.client.Del(ctx, key).Err()
}

// Cleanup removes expired keys (Redis TTL handles this automatically)
func (rl *RedisRateLimiter) Cleanup(ctx context.Context) error {
	// Redis TTL handles cleanup automatically
	// This method is here for interface compatibility
	return nil
}

// GetStats returns statistics about rate limiting
func (rl *RedisRateLimiter) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// Get all rate limit keys
	keys, err := rl.client.Keys(ctx, "rate_limit:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit keys: %w", err)
	}

	stats := map[string]interface{}{
		"total_buckets": len(keys),
		"config": map[string]interface{}{
			"capacity":    rl.config.Capacity,
			"refill_rate": rl.config.RefillRate,
			"window":      rl.config.Window.String(),
		},
	}

	return stats, nil
}
