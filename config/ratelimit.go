package config

import (
	"time"
)

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool          `json:"enabled"`
	Identifier  string        `json:"identifier"` // "ip", "jwt", "apikey", "user"
	Capacity    int           `json:"capacity"`
	RefillRate  int           `json:"refill_rate"`
	Window      time.Duration `json:"window"`
	UseRedis    bool          `json:"use_redis"`
	Redis       RedisConfig   `json:"redis"`
	SkipSuccess bool          `json:"skip_success"`
	SkipFailed  bool          `json:"skip_failed"`
}

// RedisConfig represents Redis configuration for rate limiting
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	PoolSize int    `json:"pool_size"`
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Enabled:    true,
		Identifier: "ip",
		Capacity:   100,
		RefillRate: 10,
		Window:     time.Minute,
		UseRedis:   false,
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
		SkipSuccess: false,
		SkipFailed:  false,
	}
}

// LoadRateLimitConfig loads rate limiting configuration from environment
func LoadRateLimitConfig() *RateLimitConfig {
	config := DefaultRateLimitConfig()

	// Load from environment variables
	config.Enabled = getEnvBool("RATE_LIMIT_ENABLED", true)
	if !config.Enabled {
		return config
	}

	config.Identifier = getEnvString("RATE_LIMIT_IDENTIFIER", "ip")
	config.Capacity = getEnvInt("RATE_LIMIT_CAPACITY", 100)
	config.RefillRate = getEnvInt("RATE_LIMIT_REFILL_RATE", 10)
	config.Window = getEnvDuration("RATE_LIMIT_WINDOW", time.Minute)
	config.UseRedis = getEnvBool("RATE_LIMIT_USE_REDIS", false)
	config.SkipSuccess = getEnvBool("RATE_LIMIT_SKIP_SUCCESS", false)
	config.SkipFailed = getEnvBool("RATE_LIMIT_SKIP_FAILED", false)

	// Redis configuration
	config.Redis.Host = getEnvString("REDIS_HOST", "localhost")
	config.Redis.Port = getEnvInt("REDIS_PORT", 6379)
	config.Redis.Password = getEnvString("REDIS_PASSWORD", "")
	config.Redis.DB = getEnvInt("REDIS_DB", 0)
	config.Redis.PoolSize = getEnvInt("REDIS_POOL_SIZE", 10)

	return config
}
