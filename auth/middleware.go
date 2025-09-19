package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// AuthType represents the type of authentication
type AuthType int

const (
	AuthTypeJWT AuthType = iota
	AuthTypeAPIKey
	AuthTypeBoth
)

// AuthConfig configures authentication requirements
type AuthConfig struct {
	Type     AuthType
	Required bool
}

// UserContext represents the authenticated user context
type UserContext struct {
	UserID   string
	Username string
	Email    string
	Roles    []string
	AuthType string // "jwt" or "apikey"
	APIKey   *APIKey
}

// contextKey is a custom type for context keys
type contextKey string

const userContextKey contextKey = "user"

// AuthMiddleware creates a middleware that supports both JWT and API Key authentication
func AuthMiddleware(jwtManager *JWTManager, apiKeyStore *APIKeyStore, config AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userCtx *UserContext

			// Try JWT authentication first if required
			if config.Type == AuthTypeJWT || config.Type == AuthTypeBoth {
				userCtx, _ = authenticateJWT(r, jwtManager)
				if userCtx != nil {
					userCtx.AuthType = "jwt"
					r = r.WithContext(context.WithValue(r.Context(), userContextKey, userCtx))
					next.ServeHTTP(w, r)
					return
				}
			}

			// Try API Key authentication if JWT failed or if API Key is required
			if config.Type == AuthTypeAPIKey || config.Type == AuthTypeBoth {
				userCtx, _ = authenticateAPIKey(r, apiKeyStore)
				if userCtx != nil {
					userCtx.AuthType = "apikey"
					r = r.WithContext(context.WithValue(r.Context(), userContextKey, userCtx))
					next.ServeHTTP(w, r)
					return
				}
			}

			// If authentication is required and both methods failed
			if config.Required {
				http.Error(w, `{"error":"Authentication required","details":"Valid JWT token or API key required"}`, http.StatusUnauthorized)
				return
			}

			// If authentication is not required, continue without user context
			next.ServeHTTP(w, r)
		})
	}
}

// authenticateJWT attempts to authenticate using JWT
func authenticateJWT(r *http.Request, jwtManager *JWTManager) (*UserContext, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("no authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return &UserContext{
		UserID:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Roles:    claims.Roles,
	}, nil
}

// authenticateAPIKey attempts to authenticate using API Key
func authenticateAPIKey(r *http.Request, apiKeyStore *APIKeyStore) (*UserContext, error) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided")
	}

	key, err := apiKeyStore.ValidateAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	return &UserContext{
		UserID:   key.UserID,
		Username: key.Name,
		Roles:    key.Roles,
		APIKey:   key,
	}, nil
}

// GetUserFromContext extracts user context from request context
func GetUserFromContext(r *http.Request) *UserContext {
	userCtx, ok := r.Context().Value(userContextKey).(*UserContext)
	if !ok {
		return nil
	}
	return userCtx
}

// RBACMiddleware creates role-based access control middleware
func RBACMiddleware(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx := GetUserFromContext(r)
			if userCtx == nil {
				http.Error(w, `{"error":"Authentication required","details":"User context not found"}`, http.StatusUnauthorized)
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, requiredRole := range requiredRoles {
				for _, userRole := range userCtx.Roles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				http.Error(w, `{"error":"Insufficient permissions","details":"Required roles: `+strings.Join(requiredRoles, ", ")+`"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireJWT creates middleware that requires JWT authentication
func RequireJWT(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return AuthMiddleware(jwtManager, nil, AuthConfig{Type: AuthTypeJWT, Required: true})
}

// RequireAPIKey creates middleware that requires API Key authentication
func RequireAPIKey(apiKeyStore *APIKeyStore) func(http.Handler) http.Handler {
	return AuthMiddleware(nil, apiKeyStore, AuthConfig{Type: AuthTypeAPIKey, Required: true})
}

// RequireEither creates middleware that requires either JWT or API Key authentication
func RequireEither(jwtManager *JWTManager, apiKeyStore *APIKeyStore) func(http.Handler) http.Handler {
	return AuthMiddleware(jwtManager, apiKeyStore, AuthConfig{Type: AuthTypeBoth, Required: true})
}

// OptionalAuth creates middleware that accepts JWT or API Key but doesn't require authentication
func OptionalAuth(jwtManager *JWTManager, apiKeyStore *APIKeyStore) func(http.Handler) http.Handler {
	return AuthMiddleware(jwtManager, apiKeyStore, AuthConfig{Type: AuthTypeBoth, Required: false})
}
