package handlers

import (
	"encoding/json"
	"net/http"

	"api-gateway/auth"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Authentication required"`
	Details string `json:"details" example:"Invalid token"`
}

// ProtectedResponse represents a protected endpoint response
type ProtectedResponse struct {
	Message string   `json:"message" example:"This is a protected endpoint"`
	User    string   `json:"user" example:"admin"`
	Roles   []string `json:"roles" example:"admin,user"`
}

// ProtectedHandler handles protected endpoints
type ProtectedHandler struct{}

// NewProtectedHandler creates a new protected handler
func NewProtectedHandler() *ProtectedHandler {
	return &ProtectedHandler{}
}

// AdminOnly handles admin-only endpoints
// @Summary Admin endpoint
// @Description Access admin-only endpoint (requires admin role)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ProtectedResponse "Access granted"
// @Failure 401 {object} ErrorResponse "Authentication required"
// @Failure 403 {object} ErrorResponse "Insufficient permissions"
// @Router /api/admin [get]
func (h *ProtectedHandler) AdminOnly(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"message": "This is an admin-only endpoint",
		"user":    claims.Username,
		"roles":   claims.Roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ModeratorOnly handles moderator-only endpoints
// @Summary Moderator endpoint
// @Description Access moderator-only endpoint (requires moderator role)
// @Tags Moderator
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ProtectedResponse "Access granted"
// @Failure 401 {object} ErrorResponse "Authentication required"
// @Failure 403 {object} ErrorResponse "Insufficient permissions"
// @Router /api/moderator [get]
func (h *ProtectedHandler) ModeratorOnly(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"message": "This is a moderator-only endpoint",
		"user":    claims.Username,
		"roles":   claims.Roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UserOnly handles user-only endpoints
// @Summary User endpoint
// @Description Access user-level endpoint (requires authentication)
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ProtectedResponse "Access granted"
// @Failure 401 {object} ErrorResponse "Authentication required"
// @Router /api/user [get]
func (h *ProtectedHandler) UserOnly(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"message": "This is a user-only endpoint",
		"user":    claims.Username,
		"roles":   claims.Roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MixedRoles handles endpoints that require multiple roles
// @Summary Mixed roles endpoint
// @Description Access endpoint requiring admin or moderator role
// @Tags Mixed
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ProtectedResponse "Access granted"
// @Failure 401 {object} ErrorResponse "Authentication required"
// @Failure 403 {object} ErrorResponse "Insufficient permissions"
// @Router /api/mixed [get]
func (h *ProtectedHandler) MixedRoles(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"message": "This endpoint requires admin or moderator role",
		"user":    claims.Username,
		"roles":   claims.Roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthCheck handles health check endpoint (no auth required)
// @Summary Health check
// @Description Health check endpoint to verify service status
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "Service is healthy"
// @Router /health [get]
func (h *ProtectedHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "api-gateway",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
