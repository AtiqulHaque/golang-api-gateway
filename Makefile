# API Gateway Makefile

# Variables
APP_NAME = api-gateway
DOCKER_IMAGE = $(APP_NAME):latest
DOCKER_CONTAINER = $(APP_NAME)-container
PORT = 8080

# Colors for output
RED = \033[0;31m
GREEN = \033[0;32m
YELLOW = \033[0;33m
BLUE = \033[0;34m
NC = \033[0m # No Color

.PHONY: help build run stop clean test docker-build docker-run docker-stop docker-clean compose-up compose-down compose-logs dev

# Default target
help: ## Show this help message
	@echo "$(BLUE)API Gateway - Available Commands:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

# Development commands
build: ## Build the Go application
	@echo "$(BLUE)Building Go application...$(NC)"
	go build -o $(APP_NAME) .
	@echo "$(GREEN)✓ Build completed$(NC)"

run: build ## Build and run the application locally
	@echo "$(BLUE)Starting API Gateway on port $(PORT)...$(NC)"
	@echo "$(YELLOW)Swagger UI: http://localhost:$(PORT)/swagger/$(NC)"
	@echo "$(YELLOW)API Docs: http://localhost:$(PORT)/docs$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop$(NC)"
	./$(APP_NAME)

stop: ## Stop the running application
	@echo "$(BLUE)Stopping API Gateway...$(NC)"
	@pkill -f $(APP_NAME) || true
	@echo "$(GREEN)✓ Stopped$(NC)"

clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -f $(APP_NAME)
	@go clean
	@echo "$(GREEN)✓ Cleaned$(NC)"

test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v ./...
	@echo "$(GREEN)✓ Tests completed$(NC)"

# Docker commands
docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE) .
	@echo "$(GREEN)✓ Docker image built: $(DOCKER_IMAGE)$(NC)"

docker-run: docker-build ## Build and run Docker container
	@echo "$(BLUE)Starting Docker container...$(NC)"
	@docker stop $(DOCKER_CONTAINER) 2>/dev/null || true
	@docker rm $(DOCKER_CONTAINER) 2>/dev/null || true
	docker run -d \
		--name $(DOCKER_CONTAINER) \
		-p $(PORT):8080 \
		-e JWT_SECRET=your-secret-key-change-in-production \
		-e JWT_ISSUER=api-gateway \
		-e JWT_AUDIENCE=api-users \
		-e JWT_EXPIRY_HOURS=24 \
		$(DOCKER_IMAGE)
	@echo "$(GREEN)✓ Container started: $(DOCKER_CONTAINER)$(NC)"
	@echo "$(YELLOW)Swagger UI: http://localhost:$(PORT)/swagger/$(NC)"
	@echo "$(YELLOW)API Docs: http://localhost:$(PORT)/docs$(NC)"

docker-stop: ## Stop Docker container
	@echo "$(BLUE)Stopping Docker container...$(NC)"
	@docker stop $(DOCKER_CONTAINER) 2>/dev/null || true
	@docker rm $(DOCKER_CONTAINER) 2>/dev/null || true
	@echo "$(GREEN)✓ Container stopped$(NC)"

docker-logs: ## Show Docker container logs
	@echo "$(BLUE)Showing container logs...$(NC)"
	docker logs -f $(DOCKER_CONTAINER)

docker-clean: docker-stop ## Clean Docker images and containers
	@echo "$(BLUE)Cleaning Docker resources...$(NC)"
	@docker rmi $(DOCKER_IMAGE) 2>/dev/null || true
	@docker system prune -f
	@echo "$(GREEN)✓ Docker cleaned$(NC)"

# Docker Compose commands
compose-up: ## Start services with Docker Compose
	@echo "$(BLUE)Starting services with Docker Compose...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Services started$(NC)"
	@echo "$(YELLOW)API Gateway: http://localhost:$(PORT)$(NC)"
	@echo "$(YELLOW)Swagger UI: http://localhost:$(PORT)/swagger/$(NC)"

compose-up-dev: ## Start development services with hot reload
	@echo "$(BLUE)Starting development services with hot reload...$(NC)"
	docker-compose -f docker-compose.yml -f docker-compose.override.yml up -d
	@echo "$(GREEN)✓ Development services started$(NC)"
	@echo "$(YELLOW)API Gateway: http://localhost:$(PORT)$(NC)"
	@echo "$(YELLOW)Swagger UI: http://localhost:$(PORT)/swagger/$(NC)"
	@echo "$(YELLOW)Hot reload is enabled - file changes will trigger rebuilds$(NC)"

compose-up-prod: ## Start services with Docker Compose (production with nginx)
	@echo "$(BLUE)Starting production services with Docker Compose...$(NC)"
	docker-compose --profile production up -d
	@echo "$(GREEN)✓ Production services started$(NC)"
	@echo "$(YELLOW)API Gateway: http://localhost$(NC)"
	@echo "$(YELLOW)Swagger UI: http://localhost/swagger/$(NC)"

compose-down: ## Stop services with Docker Compose
	@echo "$(BLUE)Stopping services with Docker Compose...$(NC)"
	docker-compose down
	@echo "$(GREEN)✓ Services stopped$(NC)"

compose-down-dev: ## Stop development services
	@echo "$(BLUE)Stopping development services...$(NC)"
	docker-compose -f docker-compose.yml -f docker-compose.override.yml down
	@echo "$(GREEN)✓ Development services stopped$(NC)"

compose-logs: ## Show Docker Compose logs
	@echo "$(BLUE)Showing Docker Compose logs...$(NC)"
	docker-compose logs -f

compose-logs-dev: ## Show development service logs
	@echo "$(BLUE)Showing development service logs...$(NC)"
	docker-compose -f docker-compose.yml -f docker-compose.override.yml logs -f

compose-restart: ## Restart services with Docker Compose
	@echo "$(BLUE)Restarting services with Docker Compose...$(NC)"
	docker-compose restart
	@echo "$(GREEN)✓ Services restarted$(NC)"

# Development workflow
dev: ## Start development environment (Docker with volume mounting)
	@echo "$(BLUE)Starting development environment...$(NC)"
	@echo "$(YELLOW)Note: For hot reload, use 'make dev-local' instead$(NC)"
	@make compose-up
	@echo "$(GREEN)✓ Development environment ready$(NC)"
	@echo "$(YELLOW)API Gateway: http://localhost:$(PORT)$(NC)"
	@echo "$(YELLOW)Swagger UI: http://localhost:$(PORT)/swagger/$(NC)"
	@echo "$(YELLOW)Use 'make compose-logs' to view logs$(NC)"

dev-local: ## Start local development with hot reload (requires air)
	@echo "$(BLUE)Starting local development with hot reload...$(NC)"
	@if ! command -v air >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing air for hot reload...$(NC)"; \
		go install github.com/air-verse/air@latest; \
	fi
	@echo "$(YELLOW)Starting with hot reload...$(NC)"
	@$(shell go env GOPATH)/bin/air -c .air.toml

watch: dev ## Alias for dev command
	@echo "$(YELLOW)Use 'make dev' for development with hot reload$(NC)"

# Utility commands
status: ## Show status of services
	@echo "$(BLUE)Service Status:$(NC)"
	@echo ""
	@echo "$(YELLOW)Local Process:$(NC)"
	@pgrep -f $(APP_NAME) > /dev/null && echo "  $(GREEN)✓ Running$(NC)" || echo "  $(RED)✗ Not running$(NC)"
	@echo ""
	@echo "$(YELLOW)Docker Container:$(NC)"
	@docker ps --filter name=$(DOCKER_CONTAINER) --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "  $(RED)✗ Not running$(NC)"
	@echo ""
	@echo "$(YELLOW)Docker Compose:$(NC)"
	@docker-compose ps 2>/dev/null || echo "  $(RED)✗ Not running$(NC)"

health: ## Check API Gateway health
	@echo "$(BLUE)Checking API Gateway health...$(NC)"
	@curl -s http://localhost:$(PORT)/health | jq . || echo "$(RED)✗ API Gateway not responding$(NC)"

# Quick test
test-api: ## Test API endpoints
	@echo "$(BLUE)Testing API endpoints...$(NC)"
	@chmod +x test_api.sh
	@./test_api.sh

test-swagger: ## Test Swagger UI authentication
	@echo "$(BLUE)Testing Swagger UI authentication...$(NC)"
	@chmod +x test_swagger_auth.sh
	@./test_swagger_auth.sh

# Install dependencies
deps: ## Install Go dependencies
	@echo "$(BLUE)Installing dependencies...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✓ Dependencies installed$(NC)"

# Format code
fmt: ## Format Go code
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

# Lint code
lint: ## Lint Go code
	@echo "$(BLUE)Linting code...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not installed, skipping...$(NC)"; \
	fi

# File watching utilities
watch-files: ## Watch for file changes and show notifications
	@echo "$(BLUE)Watching for file changes...$(NC)"
	@if command -v fswatch >/dev/null 2>&1; then \
		fswatch -o . --exclude='(api-gateway|\.git|tmp|node_modules)' | while read; do \
			echo "$(GREEN)File change detected at $(date)$(NC)"; \
		done; \
	elif command -v inotifywait >/dev/null 2>&1; then \
		inotifywait -r -e modify,create,delete . --exclude='(api-gateway|\.git|tmp|node_modules)' | while read; do \
			echo "$(GREEN)File change detected: $$REPLY$(NC)"; \
		done; \
	else \
		echo "$(YELLOW)No file watcher available. Install fswatch or inotify-tools$(NC)"; \
	fi

install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	@go install github.com/cosmtrek/air@latest
	@if command -v brew >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing fswatch via brew...$(NC)"; \
		brew install fswatch; \
	elif command -v apt-get >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing inotify-tools via apt...$(NC)"; \
		sudo apt-get install -y inotify-tools; \
	fi
	@echo "$(GREEN)✓ Development tools installed$(NC)"

# Development status
dev-status: ## Show development environment status
	@echo "$(BLUE)Development Environment Status:$(NC)"
	@echo ""
	@echo "$(YELLOW)Hot Reload Tools:$(NC)"
	@if command -v air >/dev/null 2>&1; then \
		echo "  $(GREEN)✓ Air installed$(NC)"; \
	else \
		echo "  $(RED)✗ Air not installed$(NC)"; \
	fi
	@if command -v fswatch >/dev/null 2>&1; then \
		echo "  $(GREEN)✓ fswatch available$(NC)"; \
	elif command -v inotifywait >/dev/null 2>&1; then \
		echo "  $(GREEN)✓ inotifywait available$(NC)"; \
	else \
		echo "  $(RED)✗ No file watcher available$(NC)"; \
	fi
	@echo ""
	@echo "$(YELLOW)Docker Services:$(NC)"
	@docker-compose ps 2>/dev/null || echo "  $(RED)✗ No services running$(NC)"
