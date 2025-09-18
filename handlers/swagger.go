package handlers

import (
	"net/http"
	"strings"

	"api-gateway/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// SwaggerHandler handles Swagger documentation endpoints
type SwaggerHandler struct{}

// NewSwaggerHandler creates a new Swagger handler
func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

// SwaggerUI serves the Swagger UI
func (h *SwaggerHandler) SwaggerUI(w http.ResponseWriter, r *http.Request) {
	// Remove the /swagger prefix from the request path
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/swagger")
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	handler := httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	)
	handler.ServeHTTP(w, r)
}

// SwaggerJSON serves the Swagger JSON
func (h *SwaggerHandler) SwaggerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(docs.SwaggerInfo.ReadDoc()))
}
