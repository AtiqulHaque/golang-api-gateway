package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	PoolSize int    `json:"pool_size"`
}

// DefaultRedisConfig returns default Redis configuration
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 10,
	}
}

// NewRedisClient creates a new Redis client
func NewRedisClient(config *RedisConfig) *redis.Client {
	if config == nil {
		config = DefaultRedisConfig()
	}

	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})
}

// TestRedisConnection tests the Redis connection
func TestRedisConnection(client *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}

	return nil
}

// RedisManager manages Redis connections and operations
type RedisManager struct {
	client *redis.Client
	config *RedisConfig
}

// NewRedisManager creates a new Redis manager
func NewRedisManager(config *RedisConfig) (*RedisManager, error) {
	client := NewRedisClient(config)

	// Test connection
	if err := TestRedisConnection(client); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisManager{
		client: client,
		config: config,
	}, nil
}

// GetClient returns the Redis client
func (rm *RedisManager) GetClient() *redis.Client {
	return rm.client
}

// Close closes the Redis connection
func (rm *RedisManager) Close() error {
	return rm.client.Close()
}

// HealthCheck checks Redis health
func (rm *RedisManager) HealthCheck(ctx context.Context) error {
	_, err := rm.client.Ping(ctx).Result()
	return err
}

// GetInfo returns Redis server information
func (rm *RedisManager) GetInfo(ctx context.Context) (map[string]string, error) {
	info, err := rm.client.Info(ctx).Result()
	if err != nil {
		return nil, err
	}

	// Parse info string into map
	result := make(map[string]string)
	lines := splitLines(info)
	for _, line := range lines {
		if line == "" || line[0] == '#' {
			continue
		}
		if idx := findChar(line, ':'); idx != -1 {
			key := line[:idx]
			value := line[idx+1:]
			result[key] = value
		}
	}

	return result, nil
}

// Helper functions
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func findChar(s string, c rune) int {
	for i, char := range s {
		if char == c {
			return i
		}
	}
	return -1
}
