# API Gateway with JWT Authentication and RBAC

A secure Go API gateway with JWT authentication and role-based access control (RBAC).

## Features

- **JWT Authentication**: Configurable secret keys, issuer, and audience validation
- **Token Validation**: Expiration, issuer, and audience checks
- **RBAC Support**: Role-based access control for different endpoints
- **Middleware**: Reusable authentication and authorization middleware
- **Helper Functions**: Easy extraction of claims from JWT tokens
- **Configuration**: Environment-based configuration management
- **API Documentation**: Interactive Swagger UI with comprehensive endpoint documentation

## Project Structure

```
api-gateway/
├── auth/
│   ├── jwt.go          # JWT token generation and validation
│   └── middleware.go   # Authentication and RBAC middleware
├── config/
│   └── config.go       # Configuration management
├── docs/
│   ├── docs.go         # Swagger documentation
│   └── swagger.json    # OpenAPI specification
├── handlers/
│   ├── auth.go         # Authentication endpoints
│   ├── protected.go    # Protected endpoints with role examples
│   └── swagger.go      # Swagger documentation handler
├── main.go             # Main application entry point
├── test_api.sh         # API testing script
├── go.mod              # Go module dependencies
└── README.md           # This file
```

## Quick Start

### Option 1: Docker (Recommended)

1. **Using Docker Compose** (easiest):
   ```bash
   make compose-up
   ```

2. **Using Docker directly**:
   ```bash
   make docker-run
   ```

3. **Using Makefile** (development):
   ```bash
   make dev
   ```

### Option 2: Local Development

1. **Install dependencies**:
   ```bash
   make deps
   ```

2. **Set environment variables** (optional):
   ```bash
   export JWT_SECRET="your-secret-key-here"
   export JWT_ISSUER="api-gateway"
   export JWT_AUDIENCE="api-users"
   export JWT_EXPIRY_HOURS="24"
   export PORT="8080"
   ```

3. **Run the application**:
   ```bash
   make run
   ```

### Available Make Commands

Run `make help` to see all available commands:

```bash
# Development Commands
make dev           # Start development environment with hot reload
make dev-local     # Start local development with hot reload (requires air)
make watch         # Alias for dev command
make watch-files   # Watch for file changes and show notifications
make dev-status    # Show development environment status

# Docker Commands
make compose-up    # Start with Docker Compose
make compose-up-dev # Start development services with hot reload
make compose-up-prod # Start production services with nginx
make docker-run    # Build and run with Docker
make docker-build  # Build Docker image

# Utility Commands
make build         # Build the Go application
make run           # Build and run locally
make test-api      # Test API endpoints
make status        # Show service status
make health        # Check API health
make install-tools # Install development tools
```

## API Endpoints

### Public Endpoints
- `POST /login` - User login
- `GET /health` - Health check
- `GET /swagger/` - Interactive Swagger UI documentation
- `GET /docs` - Redirect to Swagger UI
- `GET /swagger/doc.json` - OpenAPI specification (JSON)

### Protected Endpoints (require authentication)
- `GET /api/profile` - Get user profile
- `POST /api/refresh` - Refresh JWT token
- `GET /api/user` - User endpoint (any authenticated user)
- `GET /api/moderator` - Moderator only (requires moderator role)
- `GET /api/admin` - Admin only (requires admin role)
- `GET /api/mixed` - Admin or Moderator (requires admin or moderator role)

## Authentication

### Login
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'
```

### Using JWT Token
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/profile
```

## Test Users

| Username  | Password | Roles           |
|-----------|----------|-----------------|
| admin     | admin123 | admin, user     |
| moderator | mod123   | moderator, user |
| user      | user123  | user            |

## JWT Configuration

The JWT configuration can be set via environment variables:

- `JWT_SECRET`: Secret key for signing tokens (default: "default-secret-key")
- `JWT_ISSUER`: Token issuer (default: "api-gateway")
- `JWT_AUDIENCE`: Token audience (default: "api-users")
- `JWT_EXPIRY_HOURS`: Token expiry in hours (default: 24)
- `PORT`: Server port (default: "8080")

## Usage Examples

### 1. Basic Authentication Middleware

```go
// Apply authentication to all routes
router.Use(auth.AuthMiddleware(jwtManager))
```

### 2. Role-Based Access Control

```go
// Require specific roles
router.HandleFunc("/admin", handler.AdminOnly).Methods("GET").
    Middleware(auth.RBACMiddleware("admin"))

// Require multiple roles (OR logic)
router.HandleFunc("/moderator", handler.ModeratorOnly).Methods("GET").
    Middleware(auth.RBACMiddleware("admin", "moderator"))
```

### 3. Extract Claims from Context

```go
// Get user claims
claims, ok := auth.GetClaimsFromContext(r.Context())
if ok {
    userID := claims.UserID
    roles := claims.Roles
}

// Get specific user info
userID, ok := auth.GetUserIDFromContext(r.Context())
roles, ok := auth.GetUserRolesFromContext(r.Context())

// Check specific role
isAdmin := auth.HasRole(r.Context(), "admin")
```

## Security Features

- **Token Validation**: Comprehensive JWT validation including expiration, issuer, and audience
- **Role-Based Access**: Fine-grained access control based on user roles
- **Context Integration**: Seamless integration with Go's context package
- **Error Handling**: Proper HTTP status codes and error messages
- **CORS Support**: Built-in CORS middleware for web applications

## API Documentation

The API includes comprehensive Swagger documentation accessible at:

- **Swagger UI**: `http://localhost:8080/swagger/` - Interactive documentation interface
- **OpenAPI JSON**: `http://localhost:8080/swagger/doc.json` - Machine-readable API specification
- **Quick Access**: `http://localhost:8080/docs` - Redirects to Swagger UI

The documentation includes:
- Complete endpoint descriptions
- Request/response schemas
- Authentication requirements
- Role-based access control details
- Interactive testing capabilities

### Using Swagger UI with Authentication
1. **Get a JWT Token**: Use the `/login` endpoint with valid credentials
2. **Authorize in Swagger**: Click "Authorize" button and enter `Bearer <your-token>`
3. **Test Protected Endpoints**: All protected endpoints will now work with your token

**Available Test Credentials:**
- Admin: `admin` / `admin123`
- User: `user` / `user123`
- Moderator: `moderator` / `mod123`

📖 **Detailed Guide**: See [SWAGGER_AUTH_GUIDE.md](SWAGGER_AUTH_GUIDE.md) for step-by-step instructions

## Docker Environment

The project includes a complete Docker setup for easy deployment and development:

### Docker Files
- `Dockerfile` - Multi-stage build for production-ready image
- `Dockerfile.dev` - Development image with hot reload support
- `docker-compose.yml` - Service orchestration with optional nginx
- `docker-compose.override.yml` - Development overrides with hot reload
- `docker-compose.dev.yml` - Alternative development configuration
- `.dockerignore` - Excludes unnecessary files from Docker context
- `nginx.conf` - Reverse proxy configuration for production
- `.air.toml` - Air configuration for hot reload

### Docker Features
- **Multi-stage build** for optimized image size
- **Non-root user** for security
- **Health checks** for container monitoring
- **Environment variable** configuration
- **Nginx reverse proxy** for production deployments
- **Hot reload** for development with file watching
- **Volume mounting** for live code updates

### Development Features
- **File watching** with automatic rebuilds
- **Hot reload** using Air (Go hot reload tool)
- **Volume mounting** for instant code changes
- **Development tools** auto-installation
- **Multiple development modes** (Docker vs local)

### Development Workflow

1. **Start development environment**:
   ```bash
   make dev
   ```

2. **Make code changes** - The server will automatically rebuild and restart

3. **View logs** to see rebuild notifications:
   ```bash
   make compose-logs-dev
   ```

4. **Check development status**:
   ```bash
   make dev-status
   ```

5. **Stop development environment**:
   ```bash
   make compose-down-dev
   ```

### File Watching

The development environment automatically detects changes to:
- `.go` files (Go source code)
- `.html` files (templates)
- `.toml` files (configuration)
- `.json` files (configuration)

Changes trigger:
- Automatic rebuild of the application
- Container restart with new code
- Log notifications of the rebuild process

### Production Deployment

1. **With Nginx** (recommended for production):
   ```bash
   make compose-up-prod
   ```

2. **Environment Configuration**:
   ```bash
   # Create .env file
   echo "JWT_SECRET=your-production-secret-key" > .env
   echo "JWT_ISSUER=your-domain.com" >> .env
   echo "JWT_AUDIENCE=your-api-users" >> .env
   
   # Start with environment file
   docker-compose --env-file .env up -d
   ```

## Dependencies

- `github.com/golang-jwt/jwt/v5` - JWT token handling
- `github.com/gorilla/mux` - HTTP router
- `github.com/joho/godotenv` - Environment variable loading
- `github.com/swaggo/swag` - Swagger documentation generation
- `github.com/swaggo/http-swagger` - Swagger UI serving
- `github.com/swaggo/files` - Static file serving for Swagger

## License

MIT License
