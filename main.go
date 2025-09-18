package main

import (
	"fmt"
	"log"
	"net/http"

	"api-gateway/auth"
	"api-gateway/config"
	_ "api-gateway/docs" // Import docs package for Swagger
	"api-gateway/handlers"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println("Configuration loaded successfully")
	fmt.Println("JWT Secret:", cfg.JWT.Secret)
	fmt.Println("JWT Issuer:", cfg.JWT.Issuer)
	fmt.Println("JWT Audience:", cfg.JWT.Audience)
	fmt.Println("JWT Expiry:", cfg.JWT.Expiry)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.Issuer,
		cfg.JWT.Audience,
		cfg.JWT.Expiry,
	)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(jwtManager)
	protectedHandler := handlers.NewProtectedHandler()
	swaggerHandler := handlers.NewSwaggerHandler()

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

	// Protected routes (authentication required)
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(auth.AuthMiddleware(jwtManager))

	// Authentication endpoints
	protected.HandleFunc("/profile", authHandler.Profile).Methods("GET")
	protected.HandleFunc("/refresh", authHandler.RefreshToken).Methods("POST")

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
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Apply CORS to all routes
	router.Use(corsHandler)

	// Start server
	port := cfg.Server.Port
	fmt.Printf("API Gateway starting on port %s\n", port)
	fmt.Printf("Swagger UI: http://localhost:%s/swagger/\n", port)
	fmt.Printf("API Docs: http://localhost:%s/docs\n", port)

	log.Fatal(http.ListenAndServe(":"+port, router))
}
