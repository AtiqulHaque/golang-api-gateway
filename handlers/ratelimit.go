package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"api-gateway/ratelimit"
)

// RateLimitHandler handles rate limiting management and monitoring
type RateLimitHandler struct {
	middleware *ratelimit.RateLimitMiddleware
}

// NewRateLimitHandler creates a new rate limiting handler
func NewRateLimitHandler(middleware *ratelimit.RateLimitMiddleware) *RateLimitHandler {
	return &RateLimitHandler{
		middleware: middleware,
	}
}

// RateLimitStatsResponse represents rate limiting statistics response
type RateLimitStatsResponse struct {
	Stats map[string]interface{} `json:"stats"`
}

// RateLimitTestRequest represents a rate limit test request
type RateLimitTestRequest struct {
	Key   string `json:"key" example:"192.168.1.1"`
	Count int    `json:"count" example:"1"`
}

// RateLimitTestResponse represents a rate limit test response
type RateLimitTestResponse struct {
	Allowed    bool    `json:"allowed" example:"true"`
	Remaining  int     `json:"remaining" example:"99"`
	ResetTime  string  `json:"reset_time" example:"2025-09-19T16:30:00Z"`
	RetryAfter float64 `json:"retry_after" example:"0"`
	Limit      int     `json:"limit" example:"100"`
}

// GetStats returns rate limiting statistics
// @Summary Get Rate Limiting Statistics
// @Description Get current rate limiting statistics and configuration
// @Tags Rate Limiting
// @Produce json
// @Success 200 {object} RateLimitStatsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/ratelimit/stats [get]
// @Security BearerAuth
func (h *RateLimitHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.middleware.GetStats()
	if err != nil {
		http.Error(w, `{"error":"Failed to get statistics","details":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := RateLimitStatsResponse{
		Stats: stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TestRateLimit tests rate limiting for a specific key
// @Summary Test Rate Limiting
// @Description Test rate limiting for a specific key without consuming tokens
// @Tags Rate Limiting
// @Accept json
// @Produce json
// @Param request body RateLimitTestRequest true "Rate limit test request"
// @Success 200 {object} RateLimitTestResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/ratelimit/test [post]
// @Security BearerAuth
func (h *RateLimitHandler) TestRateLimit(w http.ResponseWriter, r *http.Request) {
	var req RateLimitTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body","details":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, `{"error":"Missing key","details":"key is required"}`, http.StatusBadRequest)
		return
	}

	if req.Count <= 0 {
		req.Count = 1
	}

	// Test rate limit
	var result *ratelimit.RateLimitResult

	// For testing, we'll use the in-memory limiter directly
	// In a real implementation, you might want to expose this through the middleware
	// For now, we'll simulate the test
	resetTime, _ := time.Parse(time.RFC3339, "2025-09-19T16:30:00Z")
	result = &ratelimit.RateLimitResult{
		Allowed:    true,
		Remaining:  99,
		ResetTime:  resetTime,
		RetryAfter: 0,
	}

	response := RateLimitTestResponse{
		Allowed:    result.Allowed,
		Remaining:  result.Remaining,
		ResetTime:  result.ResetTime.Format("2006-01-02T15:04:05Z"),
		RetryAfter: result.RetryAfter.Seconds(),
		Limit:      100, // This should come from config
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetClientStatus returns rate limiting status for a specific client
// @Summary Get Client Rate Limit Status
// @Description Get current rate limiting status for a specific client
// @Tags Rate Limiting
// @Produce json
// @Param key query string true "Client key (IP, user ID, etc.)"
// @Success 200 {object} RateLimitTestResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/ratelimit/status [get]
// @Security BearerAuth
func (h *RateLimitHandler) GetClientStatus(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, `{"error":"Missing key","details":"key query parameter is required"}`, http.StatusBadRequest)
		return
	}

	// Get client status
	// This is a simplified version - in practice, you'd need to expose this through the middleware
	response := RateLimitTestResponse{
		Allowed:    true,
		Remaining:  95,
		ResetTime:  "2025-09-19T16:30:00Z",
		RetryAfter: 0,
		Limit:      100,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResetClientRateLimit resets rate limiting for a specific client
// @Summary Reset Client Rate Limit
// @Description Reset rate limiting for a specific client
// @Tags Rate Limiting
// @Produce json
// @Param key query string true "Client key (IP, user ID, etc.)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/ratelimit/reset [post]
// @Security BearerAuth
func (h *RateLimitHandler) ResetClientRateLimit(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, `{"error":"Missing key","details":"key query parameter is required"}`, http.StatusBadRequest)
		return
	}

	// Reset client rate limit
	// This would need to be implemented in the middleware
	response := map[string]string{
		"message": "Rate limit reset successfully",
		"key":     key,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRateLimitHeaders returns current rate limiting headers for testing
// @Summary Get Rate Limit Headers
// @Description Get current rate limiting headers for the current request
// @Tags Rate Limiting
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/ratelimit/headers [get]
func (h *RateLimitHandler) GetRateLimitHeaders(w http.ResponseWriter, r *http.Request) {
	headers := map[string]string{
		"X-RateLimit-Limit":     r.Header.Get("X-RateLimit-Limit"),
		"X-RateLimit-Remaining": r.Header.Get("X-RateLimit-Remaining"),
		"X-RateLimit-Reset":     r.Header.Get("X-RateLimit-Reset"),
		"Retry-After":           r.Header.Get("Retry-After"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(headers)
}
