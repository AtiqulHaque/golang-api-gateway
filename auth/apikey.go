package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// APIKey represents an API key with metadata
type APIKey struct {
	Key        string    `json:"key"`
	Name       string    `json:"name"`
	UserID     string    `json:"user_id"`
	Roles      []string  `json:"roles"`
	RateLimit  int       `json:"rate_limit"` // requests per minute
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// APIKeyStore manages API keys in memory
type APIKeyStore struct {
	keys       map[string]*APIKey
	mu         sync.RWMutex
	rateLimits map[string][]time.Time // key -> timestamps of requests
	rateMu     sync.RWMutex
}

// NewAPIKeyStore creates a new API key store
func NewAPIKeyStore() *APIKeyStore {
	store := &APIKeyStore{
		keys:       make(map[string]*APIKey),
		rateLimits: make(map[string][]time.Time),
	}

	fmt.Println(store.keys)
	// Start cleanup routine for expired keys and rate limits
	go store.cleanupRoutine()

	return store
}

// GenerateAPIKey generates a new API key
func (s *APIKeyStore) GenerateAPIKey(name, userID string, roles []string, rateLimit int, expiresIn time.Duration) (*APIKey, error) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	key := &APIKey{
		Key:       "ak_" + hex.EncodeToString(keyBytes),
		Name:      name,
		UserID:    userID,
		Roles:     roles,
		RateLimit: rateLimit,
		IsActive:  true,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(expiresIn),
	}

	s.mu.Lock()
	s.keys[key.Key] = key
	s.mu.Unlock()

	return key, nil
}

// ValidateAPIKey validates an API key and checks rate limits
func (s *APIKeyStore) ValidateAPIKey(key string) (*APIKey, error) {
	s.mu.RLock()
	apiKey, exists := s.keys[key]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("invalid API key")
	}

	if !apiKey.IsActive {
		return nil, fmt.Errorf("API key is inactive")
	}

	if time.Now().After(apiKey.ExpiresAt) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Check rate limit
	if apiKey.RateLimit > 0 {
		if err := s.checkRateLimit(key, apiKey.RateLimit); err != nil {
			return nil, err
		}
	}

	// Update last used time
	s.mu.Lock()
	apiKey.LastUsedAt = time.Now()
	s.mu.Unlock()

	return apiKey, nil
}

// checkRateLimit checks if the API key is within its rate limit
func (s *APIKeyStore) checkRateLimit(key string, limit int) error {
	s.rateMu.Lock()
	defer s.rateMu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Minute) // Check last minute

	// Clean old timestamps
	timestamps := s.rateLimits[key]
	var validTimestamps []time.Time
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if within limit
	if len(validTimestamps) >= limit {
		return fmt.Errorf("rate limit exceeded: %d requests per minute", limit)
	}

	// Add current request
	validTimestamps = append(validTimestamps, now)
	s.rateLimits[key] = validTimestamps

	return nil
}

// GetAPIKey retrieves an API key by key string
func (s *APIKeyStore) GetAPIKey(key string) (*APIKey, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apiKey, exists := s.keys[key]
	return apiKey, exists
}

// ListAPIKeys returns all API keys for a user
func (s *APIKeyStore) ListAPIKeys(userID string) []*APIKey {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var userKeys []*APIKey
	for _, key := range s.keys {
		if key.UserID == userID {
			userKeys = append(userKeys, key)
		}
	}

	return userKeys
}

// RevokeAPIKey deactivates an API key
func (s *APIKeyStore) RevokeAPIKey(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	apiKey, exists := s.keys[key]
	if !exists {
		return fmt.Errorf("API key not found")
	}

	apiKey.IsActive = false
	return nil
}

// DeleteAPIKey permanently removes an API key
func (s *APIKeyStore) DeleteAPIKey(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.keys[key]; !exists {
		return fmt.Errorf("API key not found")
	}

	delete(s.keys, key)

	// Clean up rate limit data
	s.rateMu.Lock()
	delete(s.rateLimits, key)
	s.rateMu.Unlock()

	return nil
}

// cleanupRoutine periodically cleans up expired keys and old rate limit data
func (s *APIKeyStore) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		// Clean up expired keys
		s.mu.Lock()
		for key, apiKey := range s.keys {
			if now.After(apiKey.ExpiresAt) {
				delete(s.keys, key)
			}
		}
		s.mu.Unlock()

		// Clean up old rate limit data
		s.rateMu.Lock()
		cutoff := now.Add(-time.Hour) // Keep only last hour of rate limit data
		for key, timestamps := range s.rateLimits {
			var validTimestamps []time.Time
			for _, ts := range timestamps {
				if ts.After(cutoff) {
					validTimestamps = append(validTimestamps, ts)
				}
			}
			if len(validTimestamps) == 0 {
				delete(s.rateLimits, key)
			} else {
				s.rateLimits[key] = validTimestamps
			}
		}
		s.rateMu.Unlock()
	}
}

// GetStats returns statistics about API key usage
func (s *APIKeyStore) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	activeKeys := 0
	expiredKeys := 0
	now := time.Now()

	for _, key := range s.keys {
		if key.IsActive {
			if now.After(key.ExpiresAt) {
				expiredKeys++
			} else {
				activeKeys++
			}
		}
	}

	return map[string]interface{}{
		"total_keys":    len(s.keys),
		"active_keys":   activeKeys,
		"expired_keys":  expiredKeys,
		"inactive_keys": len(s.keys) - activeKeys - expiredKeys,
	}
}
