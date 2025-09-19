package main

import (
	"fmt"
	"log"
	"net/http"

	"api-gateway/auth"
	"api-gateway/config"
	_ "api-gateway/docs" // Import docs package for Swagger
	"api-gateway/handlers"
	"api-gateway/ratelimit"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.Issuer,
		cfg.JWT.Audience,
		cfg.JWT.Expiry,
	)

	// Initialize API key store
	apiKeyStore := auth.NewAPIKeyStore()

	// Initialize rate limiting
	rateLimitConfig := config.LoadRateLimitConfig()
	var rateLimitMiddleware *ratelimit.RateLimitMiddleware
	if rateLimitConfig.Enabled {
		// Convert config to middleware config
		identifier := ratelimit.ClientByIP
		switch rateLimitConfig.Identifier {
		case "jwt":
			identifier = ratelimit.ClientByJWTSubject
		case "apikey":
			identifier = ratelimit.ClientByAPIKey
		case "user":
			identifier = ratelimit.ClientByUserID
		}

		middlewareConfig := &ratelimit.RateLimitMiddlewareConfig{
			Identifier: identifier,
			Config: &ratelimit.RateLimitConfig{
				Capacity:   rateLimitConfig.Capacity,
				RefillRate: rateLimitConfig.RefillRate,
				Window:     rateLimitConfig.Window,
			},
			UseRedis: rateLimitConfig.UseRedis,
			RedisConfig: &ratelimit.RedisConfig{
				Host:     rateLimitConfig.Redis.Host,
				Port:     rateLimitConfig.Redis.Port,
				Password: rateLimitConfig.Redis.Password,
				DB:       rateLimitConfig.Redis.DB,
				PoolSize: rateLimitConfig.Redis.PoolSize,
			},
			SkipSuccessful: rateLimitConfig.SkipSuccess,
			SkipFailed:     rateLimitConfig.SkipFailed,
		}

		var err error
		rateLimitMiddleware, err = ratelimit.NewRateLimitMiddleware(middlewareConfig)
		if err != nil {
			log.Fatalf("Failed to initialize rate limiting: %v", err)
		}
	}
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(jwtManager)
	protectedHandler := handlers.NewProtectedHandler()
	swaggerHandler := handlers.NewSwaggerHandler()
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyStore)
	var rateLimitHandler *handlers.RateLimitHandler
	if rateLimitMiddleware != nil {
		rateLimitHandler = handlers.NewRateLimitHandler(rateLimitMiddleware)
	}

	// Setup routes
	router := mux.NewRouter()

	// Public routes (no authentication required)
	router.HandleFunc("/health", protectedHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/login", authHandler.Login).Methods("POST")

	// Swagger documentation routes
	router.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	}).Methods("GET")
	router.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/swagger.html")
	}).Methods("GET")
	router.HandleFunc("/swagger/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/swagger.html")
	}).Methods("GET")
	router.HandleFunc("/swagger/doc.json", swaggerHandler.SwaggerJSON).Methods("GET")
	router.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	}).Methods("GET")

	// Alternative Swagger UI endpoint
	router.HandleFunc("/swagger-ui", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	}).Methods("GET")

	// API Key test endpoint (no authentication required)
	router.HandleFunc("/api/keys/test", apiKeyHandler.TestAPIKey).Methods("GET")

	// Rate limiting endpoints
	if rateLimitHandler != nil {
		router.HandleFunc("/api/ratelimit/headers", rateLimitHandler.GetRateLimitHeaders).Methods("GET")

		// Rate limiting management endpoints (JWT required)
		rateLimitRoutes := router.PathPrefix("/api/ratelimit").Subrouter()
		rateLimitRoutes.Use(auth.RequireJWT(jwtManager))
		rateLimitRoutes.HandleFunc("/stats", rateLimitHandler.GetStats).Methods("GET")
		rateLimitRoutes.HandleFunc("/test", rateLimitHandler.TestRateLimit).Methods("POST")
		rateLimitRoutes.HandleFunc("/status", rateLimitHandler.GetClientStatus).Methods("GET")
		rateLimitRoutes.HandleFunc("/reset", rateLimitHandler.ResetClientRateLimit).Methods("POST")
	}

	// Protected routes (JWT or API Key authentication required)
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(auth.RequireEither(jwtManager, apiKeyStore))

	// Authentication endpoints
	protected.HandleFunc("/profile", authHandler.Profile).Methods("GET")
	protected.HandleFunc("/refresh", authHandler.RefreshToken).Methods("POST")

	// API Key management endpoints (JWT only)
	apiKeyRoutes := router.PathPrefix("/api/keys").Subrouter()
	apiKeyRoutes.Use(auth.RequireJWT(jwtManager))
	apiKeyRoutes.HandleFunc("", apiKeyHandler.CreateAPIKey).Methods("POST")
	apiKeyRoutes.HandleFunc("", apiKeyHandler.ListAPIKeys).Methods("GET")
	apiKeyRoutes.HandleFunc("/stats", apiKeyHandler.GetAPIKeyStats).Methods("GET")
	apiKeyRoutes.HandleFunc("/{key}", apiKeyHandler.GetAPIKey).Methods("GET")
	apiKeyRoutes.HandleFunc("/{key}/revoke", apiKeyHandler.RevokeAPIKey).Methods("POST")
	apiKeyRoutes.HandleFunc("/{key}", apiKeyHandler.DeleteAPIKey).Methods("DELETE")

	// Role-based protected routes
	protected.HandleFunc("/user", protectedHandler.UserOnly).Methods("GET")

	// Moderator-only routes
	moderatorRoutes := protected.PathPrefix("/moderator").Subrouter()
	moderatorRoutes.Use(auth.RBACMiddleware("moderator"))
	moderatorRoutes.HandleFunc("", protectedHandler.ModeratorOnly).Methods("GET")

	// Admin-only routes
	adminRoutes := protected.PathPrefix("/admin").Subrouter()
	adminRoutes.Use(auth.RBACMiddleware("admin"))
	adminRoutes.HandleFunc("", protectedHandler.AdminOnly).Methods("GET")

	// Mixed role routes (admin or moderator)
	mixedRoutes := protected.PathPrefix("/mixed").Subrouter()
	mixedRoutes.Use(auth.RBACMiddleware("admin", "moderator"))
	mixedRoutes.HandleFunc("", protectedHandler.MixedRoles).Methods("GET")

	// Add CORS middleware
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Apply rate limiting middleware if enabled
	if rateLimitMiddleware != nil {
		router.Use(rateLimitMiddleware.Middleware())
	}

	// Apply CORS to all routes
	router.Use(corsHandler)

	// Start server
	port := cfg.Server.Port
	//
	if rateLimitConfig.Enabled {
		// fmt.Printf("🚦 Rate Limiting:\n")
		// fmt.Printf("   Identifier: %s\n", rateLimitConfig.Identifier)
		// fmt.Printf("   Capacity: %d requests\n", rateLimitConfig.Capacity)
		// fmt.Printf("   Refill Rate: %d requests/second\n", rateLimitConfig.RefillRate)
		// fmt.Printf("   Window: %s\n", rateLimitConfig.Window)
		// if rateLimitConfig.UseRedis {
		// 	fmt.Printf("   Backend: Redis (%s:%d)\n", rateLimitConfig.Redis.Host, rateLimitConfig.Redis.Port)
		// } else {
		// 	fmt.Printf("   Backend: In-Memory\n")
		// }
		// fmt.Printf("   GET  /api/ratelimit/headers - Get rate limit headers\n")
		// fmt.Printf("   GET  /api/ratelimit/stats - Rate limiting statistics (JWT required)\n")
	}
	fmt.Printf("🌐 Swagger UI: http://localhost:%s/swagger/\n", port)
	fmt.Printf("📚 API Docs: http://localhost:%s/docs\n", port)

	log.Fatal(http.ListenAndServe(":"+port, router))
}
