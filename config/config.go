package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for our application
type Config struct {
	JWT    JWTConfig
	Server ServerConfig
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret      string
	Issuer      string
	Audience    string
	ExpiryHours int
	Expiry      time.Duration
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	expiryHours := 24 // default
	if hours := os.Getenv("JWT_EXPIRY_HOURS"); hours != "" {
		if h, err := strconv.Atoi(hours); err == nil {
			expiryHours = h
		}
	}

	config := &Config{
		JWT: JWTConfig{
			Secret:      getEnvOrDefault("JWT_SECRET", "default-secret-key"),
			Issuer:      getEnvOrDefault("JWT_ISSUER", "api-gateway"),
			Audience:    getEnvOrDefault("JWT_AUDIENCE", "api-users"),
			ExpiryHours: expiryHours,
			Expiry:      time.Duration(expiryHours) * time.Hour,
		},
		Server: ServerConfig{
			Port: getEnvOrDefault("PORT", "8080"),
		},
	}

	return config, nil
}

// getEnvOrDefault returns the environment variable value or a default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	return getEnvOrDefault(key, defaultValue)
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
