package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"api-gateway/auth"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response payload
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// UserInfo represents user information
type UserInfo struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// AuthHandler handles authentication-related endpoints
type AuthHandler struct {
	jwtManager *auth.JWTManager
	// In a real application, you would have a user service/database
	// For demo purposes, we'll use mock data
	users map[string]UserData
}

// UserData represents user data for authentication
type UserData struct {
	ID       string
	Username string
	Email    string
	Password string
	Roles    []string
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(jwtManager *auth.JWTManager) *AuthHandler {
	// Mock user data - in production, this would come from a database
	users := map[string]UserData{
		"admin": {
			ID:       "1",
			Username: "admin",
			Email:    "admin@example.com",
			Password: "admin123", // In production, this would be hashed
			Roles:    []string{"admin", "user"},
		},
		"user": {
			ID:       "2",
			Username: "user",
			Email:    "user@example.com",
			Password: "user123",
			Roles:    []string{"user"},
		},
		"moderator": {
			ID:       "3",
			Username: "moderator",
			Email:    "moderator@example.com",
			Password: "mod123",
			Roles:    []string{"moderator", "user"},
		},
	}

	return &AuthHandler{
		jwtManager: jwtManager,
		users:      users,
	}
}

// Login handles user login
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Login successful"
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Router /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate user credentials
	user, exists := h.users[req.Username]
	if !exists || user.Password != req.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(user.ID, user.Username, user.Email, user.Roles)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(24 * time.Hour) // This should match your JWT expiry

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Roles:    user.Roles,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Profile returns the current user's profile
// @Summary Get user profile
// @Description Get current user profile information
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserInfo "User profile retrieved successfully"
// @Failure 401 {object} ErrorResponse "Authentication required"
// @Router /api/profile [get]
func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.GetUserFromContext(r)
	if userCtx == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	userInfo := UserInfo{
		ID:       userCtx.UserID,
		Username: userCtx.Username,
		Email:    userCtx.Email,
		Roles:    userCtx.Roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// RefreshToken handles token refresh
// @Summary Refresh token
// @Description Refresh JWT token
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} LoginResponse "Token refreshed successfully"
// @Failure 401 {object} ErrorResponse "Authentication required"
// @Router /api/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.GetUserFromContext(r)
	if userCtx == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Generate new token with same claims
	token, err := h.jwtManager.GenerateToken(userCtx.UserID, userCtx.Username, userCtx.Email, userCtx.Roles)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserInfo{
			ID:       userCtx.UserID,
			Username: userCtx.Username,
			Email:    userCtx.Email,
			Roles:    userCtx.Roles,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
