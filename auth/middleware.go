package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// UserContextKey is the key for storing user claims in context
	UserContextKey ContextKey = "user_claims"
)

// AuthMiddleware creates middleware for JWT authentication
func AuthMiddleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for certain paths
			if shouldSkipAuth(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			tokenString, err := ExtractTokenFromHeader(authHeader)
			if err != nil {
				writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err.Error())
				return
			}

			// Validate token
			claims, err := jwtManager.ValidateToken(tokenString)
			if err != nil {
				writeErrorResponse(w, http.StatusUnauthorized, "Invalid token", err.Error())
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RBACMiddleware creates middleware for role-based access control
func RBACMiddleware(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user claims from context
			claims, ok := r.Context().Value(UserContextKey).(*Claims)
			if !ok {
				writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", "User claims not found in context")
				return
			}

			// Check if user has any of the required roles
			if !hasAnyRole(claims.Roles, requiredRoles) {
				writeErrorResponse(w, http.StatusForbidden, "Access denied", "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClaimsFromContext extracts user claims from request context
func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*Claims)
	return claims, ok
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := GetClaimsFromContext(ctx)
	if !ok {
		return "", false
	}
	return claims.UserID, true
}

// GetUserRolesFromContext extracts user roles from request context
func GetUserRolesFromContext(ctx context.Context) ([]string, bool) {
	claims, ok := GetClaimsFromContext(ctx)
	if !ok {
		return nil, false
	}
	return claims.Roles, true
}

// HasRole checks if user has a specific role
func HasRole(ctx context.Context, role string) bool {
	roles, ok := GetUserRolesFromContext(ctx)
	if !ok {
		return false
	}
	return hasAnyRole(roles, []string{role})
}

// hasAnyRole checks if user has any of the required roles
func hasAnyRole(userRoles []string, requiredRoles []string) bool {
	for _, requiredRole := range requiredRoles {
		for _, userRole := range userRoles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

// shouldSkipAuth determines if authentication should be skipped for a path
func shouldSkipAuth(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/login",
		"/register",
		"/docs",
		"/swagger",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// writeErrorResponse writes a JSON error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{
		"error":   message,
		"details": details,
	}

	json.NewEncoder(w).Encode(response)
}
