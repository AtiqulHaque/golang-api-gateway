package ratelimit

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ClientIdentifier represents different ways to identify clients
type ClientIdentifier int

const (
	ClientByIP ClientIdentifier = iota
	ClientByJWTSubject
	ClientByAPIKey
	ClientByUserID
)

// RateLimitMiddlewareConfig represents configuration for rate limiting middleware
type RateLimitMiddlewareConfig struct {
	Identifier     ClientIdentifier           `json:"identifier"`
	Config         *RateLimitConfig           `json:"config"`
	UseRedis       bool                       `json:"use_redis"`
	RedisConfig    *RedisConfig               `json:"redis_config"`
	SkipSuccessful bool                       `json:"skip_successful"` // Don't count successful requests
	SkipFailed     bool                       `json:"skip_failed"`     // Don't count failed requests
	CustomKeyFunc  func(*http.Request) string `json:"-"`               // Custom key generation function
}

// DefaultRateLimitMiddlewareConfig returns default configuration
func DefaultRateLimitMiddlewareConfig() *RateLimitMiddlewareConfig {
	return &RateLimitMiddlewareConfig{
		Identifier:     ClientByIP,
		Config:         DefaultRateLimitConfig(),
		UseRedis:       false,
		RedisConfig:    DefaultRedisConfig(),
		SkipSuccessful: false,
		SkipFailed:     false,
	}
}

// RateLimitMiddleware creates rate limiting middleware
type RateLimitMiddleware struct {
	config       *RateLimitMiddlewareConfig
	limiter      *RateLimiter
	redisLimiter *RedisRateLimiter
	redisManager *RedisManager
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(config *RateLimitMiddlewareConfig) (*RateLimitMiddleware, error) {
	if config == nil {
		config = DefaultRateLimitMiddlewareConfig()
	}

	rl := &RateLimitMiddleware{
		config: config,
	}

	// Initialize in-memory limiter
	rl.limiter = NewRateLimiter(config.Config)

	// Initialize Redis limiter if configured
	if config.UseRedis {
		var err error
		rl.redisManager, err = NewRedisManager(config.RedisConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Redis: %w", err)
		}

		rl.redisLimiter = NewRedisRateLimiter(rl.redisManager.GetClient(), config.Config)
	}

	return rl, nil
}

// Middleware returns the HTTP middleware function
func (rl *RateLimitMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate client key
			key := rl.generateClientKey(r)

			// Check rate limit
			var result *RateLimitResult
			var err error

			if rl.config.UseRedis && rl.redisLimiter != nil {
				ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
				defer cancel()
				result, err = rl.redisLimiter.Allow(ctx, key, 1)
			} else {
				result = rl.limiter.CheckRateLimit(key, 1)
			}

			if err != nil {
				// If Redis fails, log error but allow request
				fmt.Printf("Rate limit check failed: %v\n", err)
				next.ServeHTTP(w, r)
				return
			}

			// Add rate limit headers
			rl.addRateLimitHeaders(w, result)

			if !result.Allowed {
				// Rate limit exceeded
				rl.writeRateLimitResponse(w, result)
				return
			}

			// Create a custom response writer to track status codes
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     200,
			}

			// Call next handler
			next.ServeHTTP(rw, r)

			// Check if we should count this request based on status code
			_ = rl.shouldCountRequest(rw.statusCode)
		})
	}
}

// generateClientKey generates a unique key for the client
func (rl *RateLimitMiddleware) generateClientKey(r *http.Request) string {
	// Use custom key function if provided
	if rl.config.CustomKeyFunc != nil {
		key := rl.config.CustomKeyFunc(r)
		if key != "" {
			return key
		}
	}

	var key string
	switch rl.config.Identifier {
	case ClientByIP:
		key = rl.getClientIP(r)
	case ClientByJWTSubject:
		key = rl.getJWTSubject(r)
	case ClientByAPIKey:
		key = rl.getAPIKey(r)
	case ClientByUserID:
		key = rl.getUserID(r)
	default:
		key = rl.getClientIP(r)
	}

	// Ensure we always return a valid key
	if key == "" {
		key = "unknown:" + rl.getClientIP(r)
	}

	return key
}

// getClientIP extracts the client IP address
func (rl *RateLimitMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// getJWTSubject extracts the JWT subject
func (rl *RateLimitMiddleware) getJWTSubject(r *http.Request) string {
	// For JWT-based rate limiting, we need to extract the JWT from the header directly
	// since the authentication middleware might not have run yet
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		// Use a hash of the JWT token as the key to avoid storing the full token
		// For now, we'll use a simple approach: take first 20 chars of the token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if len(token) > 20 {
			token = token[:20]
		}
		return "jwt:" + token
	}
	// If no JWT available, fall back to IP
	return rl.getClientIP(r)
}

// getAPIKey extracts the API key
func (rl *RateLimitMiddleware) getAPIKey(r *http.Request) string {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" {
		return "apikey:" + apiKey
	}
	// If no API key available, fall back to IP
	return rl.getClientIP(r)
}

// getUserID extracts the user ID from context
func (rl *RateLimitMiddleware) getUserID(r *http.Request) string {
	// For user-based rate limiting, we need to extract from headers directly
	// since the authentication middleware might not have run yet
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		// Use a hash of the JWT token as the key
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if len(token) > 20 {
			token = token[:20]
		}
		return "user:" + token
	}
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" {
		// Use first 20 chars of API key
		if len(apiKey) > 20 {
			apiKey = apiKey[:20]
		}
		return "user:" + apiKey
	}
	// If no authentication available, fall back to IP
	return rl.getClientIP(r)
}

// shouldCountRequest determines if a request should be counted based on status code
func (rl *RateLimitMiddleware) shouldCountRequest(statusCode int) bool {
	if rl.config.SkipSuccessful && statusCode >= 200 && statusCode < 300 {
		return false
	}
	if rl.config.SkipFailed && statusCode >= 400 {
		return false
	}
	return true
}

// addRateLimitHeaders adds rate limiting headers to the response
func (rl *RateLimitMiddleware) addRateLimitHeaders(w http.ResponseWriter, result *RateLimitResult) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.config.Config.Capacity))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

	if !result.Allowed {
		w.Header().Set("Retry-After", strconv.FormatFloat(result.RetryAfter.Seconds(), 'f', 0, 64))
	}
}

// writeRateLimitResponse writes a 429 response
func (rl *RateLimitMiddleware) writeRateLimitResponse(w http.ResponseWriter, result *RateLimitResult) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)

	_ = map[string]interface{}{
		"error":       "Rate limit exceeded",
		"message":     "Too many requests",
		"retry_after": result.RetryAfter.Seconds(),
		"reset_time":  result.ResetTime.Format(time.RFC3339),
		"limit":       rl.config.Config.Capacity,
		"remaining":   result.Remaining,
	}

	fmt.Fprintf(w, `{"error":"Rate limit exceeded","message":"Too many requests","retry_after":%.0f,"reset_time":"%s","limit":%d,"remaining":%d}`,
		result.RetryAfter.Seconds(),
		result.ResetTime.Format(time.RFC3339),
		rl.config.Config.Capacity,
		result.Remaining)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetStats returns rate limiting statistics
func (rl *RateLimitMiddleware) GetStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"config": map[string]interface{}{
			"identifier":      rl.config.Identifier,
			"capacity":        rl.config.Config.Capacity,
			"refill_rate":     rl.config.Config.RefillRate,
			"window":          rl.config.Config.Window.String(),
			"use_redis":       rl.config.UseRedis,
			"skip_successful": rl.config.SkipSuccessful,
			"skip_failed":     rl.config.SkipFailed,
		},
	}

	if rl.config.UseRedis && rl.redisLimiter != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		redisStats, err := rl.redisLimiter.GetStats(ctx)
		if err != nil {
			stats["redis_error"] = err.Error()
		} else {
			stats["redis"] = redisStats
		}
	} else {
		stats["in_memory"] = map[string]interface{}{
			"buckets": len(rl.limiter.buckets),
		}
	}

	return stats, nil
}

// Close closes the rate limiter and cleans up resources
func (rl *RateLimitMiddleware) Close() error {
	if rl.limiter != nil {
		rl.limiter.Stop()
	}

	if rl.redisManager != nil {
		return rl.redisManager.Close()
	}

	return nil
}
