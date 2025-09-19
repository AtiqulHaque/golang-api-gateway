package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"api-gateway/auth"
)

// APIKeyHandler handles API key management
type APIKeyHandler struct {
	apiKeyStore *auth.APIKeyStore
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(apiKeyStore *auth.APIKeyStore) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyStore: apiKeyStore,
	}
}

// CreateAPIKeyRequest represents the request to create an API key
type CreateAPIKeyRequest struct {
	Name      string   `json:"name" example:"My API Key"`
	UserID    string   `json:"user_id" example:"user123"`
	Roles     []string `json:"roles" example:"user,admin"`
	RateLimit int      `json:"rate_limit" example:"100"`
	ExpiresIn string   `json:"expires_in" example:"24h"`
}

// CreateAPIKeyResponse represents the response for creating an API key
type CreateAPIKeyResponse struct {
	APIKey    *auth.APIKey `json:"api_key"`
	Message   string       `json:"message" example:"API key created successfully"`
	CreatedAt time.Time    `json:"created_at"`
}

// ListAPIKeysResponse represents the response for listing API keys
type ListAPIKeysResponse struct {
	APIKeys []*auth.APIKey `json:"api_keys"`
	Count   int            `json:"count"`
}

// APIKeyStatsResponse represents the response for API key statistics
type APIKeyStatsResponse struct {
	Stats map[string]interface{} `json:"stats"`
}

// CreateAPIKey creates a new API key
// @Summary Create API Key
// @Description Create a new API key with specified roles and rate limits
// @Tags API Keys
// @Accept json
// @Produce json
// @Param request body CreateAPIKeyRequest true "API Key creation request"
// @Success 201 {object} CreateAPIKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/keys [post]
// @Security BearerAuth
func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body","details":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" || req.UserID == "" || len(req.Roles) == 0 {
		http.Error(w, `{"error":"Missing required fields","details":"name, user_id, and roles are required"}`, http.StatusBadRequest)
		return
	}

	// Parse expiration time
	expiresIn := 24 * time.Hour // Default to 24 hours
	if req.ExpiresIn != "" {
		var err error
		expiresIn, err = time.ParseDuration(req.ExpiresIn)
		if err != nil {
			http.Error(w, `{"error":"Invalid expires_in format","details":"Use format like '24h', '7d', '30d'"}`, http.StatusBadRequest)
			return
		}
	}

	// Set default rate limit if not provided
	rateLimit := req.RateLimit
	if rateLimit <= 0 {
		rateLimit = 100 // Default to 100 requests per minute
	}

	// Create API key
	apiKey, err := h.apiKeyStore.GenerateAPIKey(req.Name, req.UserID, req.Roles, rateLimit, expiresIn)
	if err != nil {
		http.Error(w, `{"error":"Failed to create API key","details":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := CreateAPIKeyResponse{
		APIKey:    apiKey,
		Message:   "API key created successfully",
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ListAPIKeys lists API keys for a user
// @Summary List API Keys
// @Description List all API keys for the authenticated user
// @Tags API Keys
// @Produce json
// @Success 200 {object} ListAPIKeysResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/keys [get]
// @Security BearerAuth
func (h *APIKeyHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	// Get user ID from query parameter or context
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		// Try to get from context if available
		// This would require the request to go through auth middleware first
		http.Error(w, `{"error":"Missing user_id","details":"user_id query parameter is required"}`, http.StatusBadRequest)
		return
	}

	apiKeys := h.apiKeyStore.ListAPIKeys(userID)

	response := ListAPIKeysResponse{
		APIKeys: apiKeys,
		Count:   len(apiKeys),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAPIKey retrieves a specific API key
// @Summary Get API Key
// @Description Get details of a specific API key
// @Tags API Keys
// @Produce json
// @Param key path string true "API Key"
// @Success 200 {object} auth.APIKey
// @Failure 404 {object} ErrorResponse
// @Router /api/keys/{key} [get]
// @Security BearerAuth
func (h *APIKeyHandler) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/api/keys/"):]
	if key == "" {
		http.Error(w, `{"error":"Missing API key","details":"API key parameter is required"}`, http.StatusBadRequest)
		return
	}

	apiKey, exists := h.apiKeyStore.GetAPIKey(key)
	if !exists {
		http.Error(w, `{"error":"API key not found","details":"The specified API key does not exist"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiKey)
}

// RevokeAPIKey revokes an API key
// @Summary Revoke API Key
// @Description Revoke (deactivate) an API key
// @Tags API Keys
// @Produce json
// @Param key path string true "API Key"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Router /api/keys/{key}/revoke [post]
// @Security BearerAuth
func (h *APIKeyHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/api/keys/"):]
	key = key[:len(key)-len("/revoke")]

	if key == "" {
		http.Error(w, `{"error":"Missing API key","details":"API key parameter is required"}`, http.StatusBadRequest)
		return
	}

	err := h.apiKeyStore.RevokeAPIKey(key)
	if err != nil {
		http.Error(w, `{"error":"Failed to revoke API key","details":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	response := map[string]string{
		"message": "API key revoked successfully",
		"key":     key,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteAPIKey permanently deletes an API key
// @Summary Delete API Key
// @Description Permanently delete an API key
// @Tags API Keys
// @Produce json
// @Param key path string true "API Key"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Router /api/keys/{key} [delete]
// @Security BearerAuth
func (h *APIKeyHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/api/keys/"):]

	if key == "" {
		http.Error(w, `{"error":"Missing API key","details":"API key parameter is required"}`, http.StatusBadRequest)
		return
	}

	err := h.apiKeyStore.DeleteAPIKey(key)
	if err != nil {
		http.Error(w, `{"error":"Failed to delete API key","details":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	response := map[string]string{
		"message": "API key deleted successfully",
		"key":     key,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAPIKeyStats returns statistics about API keys
// @Summary Get API Key Statistics
// @Description Get statistics about API key usage
// @Tags API Keys
// @Produce json
// @Success 200 {object} APIKeyStatsResponse
// @Router /api/keys/stats [get]
// @Security BearerAuth
func (h *APIKeyHandler) GetAPIKeyStats(w http.ResponseWriter, r *http.Request) {
	stats := h.apiKeyStore.GetStats()

	response := APIKeyStatsResponse{
		Stats: stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TestAPIKey tests an API key
// @Summary Test API Key
// @Description Test if an API key is valid and get its details
// @Tags API Keys
// @Produce json
// @Param X-API-Key header string true "API Key"
// @Success 200 {object} auth.APIKey
// @Failure 401 {object} ErrorResponse
// @Router /api/keys/test [get]
func (h *APIKeyHandler) TestAPIKey(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, `{"error":"Missing API key","details":"X-API-Key header is required"}`, http.StatusBadRequest)
		return
	}

	key, err := h.apiKeyStore.ValidateAPIKey(apiKey)
	if err != nil {
		http.Error(w, `{"error":"Invalid API key","details":"`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(key)
}
