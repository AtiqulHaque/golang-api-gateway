# API Key Authentication Guide

## Overview

The API Gateway now supports both JWT and API Key authentication methods. API Keys provide an alternative authentication mechanism that's ideal for server-to-server communication, automated scripts, and applications that need long-lived credentials.

## Features

- **🔑 API Key Generation**: Create secure API keys with custom names and expiration
- **⚡ Rate Limiting**: Per-key rate limiting (e.g., 100 requests/minute)
- **🎯 Role-Based Access**: API keys inherit user roles for RBAC
- **🔄 Flexible Authentication**: Routes can require JWT, API Key, or both
- **📊 Management**: Full CRUD operations for API key management
- **🛡️ Security**: Secure key generation with expiration and revocation

## Authentication Methods

### 1. JWT Authentication (Bearer Token)
```bash
curl -H "Authorization: Bearer <jwt-token>" http://localhost:8080/api/profile
```

### 2. API Key Authentication
```bash
curl -H "X-API-Key: <api-key>" http://localhost:8080/api/profile
```

### 3. Both Methods (Flexible)
Routes can accept either JWT or API Key authentication.

## API Key Management

### Create API Key
```bash
curl -X POST http://localhost:8080/api/keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt-token>" \
  -d '{
    "name": "My API Key",
    "user_id": "user123",
    "roles": ["user", "admin"],
    "rate_limit": 100,
    "expires_in": "24h"
  }'
```

**Response:**
```json
{
  "api_key": {
    "key": "ak_9722e22f6bfc740b4b289a3c86a7496c42d9ab872233d75602604387bbcc30b0",
    "name": "My API Key",
    "user_id": "user123",
    "roles": ["user", "admin"],
    "rate_limit": 100,
    "is_active": true,
    "created_at": "2025-09-19T16:09:11.920028+06:00",
    "expires_at": "2025-09-20T16:09:11.920028+06:00"
  },
  "message": "API key created successfully"
}
```

### List API Keys
```bash
curl -H "Authorization: Bearer <jwt-token>" \
  "http://localhost:8080/api/keys?user_id=user123"
```

### Test API Key
```bash
curl -H "X-API-Key: <api-key>" http://localhost:8080/api/keys/test
```

### Revoke API Key
```bash
curl -X POST -H "Authorization: Bearer <jwt-token>" \
  "http://localhost:8080/api/keys/<api-key>/revoke"
```

### Delete API Key
```bash
curl -X DELETE -H "Authorization: Bearer <jwt-token>" \
  "http://localhost:8080/api/keys/<api-key>"
```

### Get Statistics
```bash
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8080/api/keys/stats
```

## Configuration Options

### API Key Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Human-readable name for the API key |
| `user_id` | string | Yes | User ID associated with the key |
| `roles` | array | Yes | Array of roles (e.g., ["user", "admin"]) |
| `rate_limit` | integer | No | Requests per minute (default: 100) |
| `expires_in` | string | No | Expiration duration (e.g., "24h", "7d", "30d") |

### Rate Limiting

API keys support configurable rate limiting:
- **Default**: 100 requests per minute
- **Custom**: Set any value during creation
- **Unlimited**: Set to 0 for no rate limiting
- **Per-key**: Each key has its own rate limit

### Expiration

API keys can have expiration times:
- **Format**: Duration strings like "24h", "7d", "30d"
- **Default**: 24 hours if not specified
- **Automatic cleanup**: Expired keys are automatically removed

## Route Configuration

### Authentication Types

Routes can be configured with different authentication requirements:

```go
// JWT only
router.Use(auth.RequireJWT(jwtManager))

// API Key only  
router.Use(auth.RequireAPIKey(apiKeyStore))

// Either JWT or API Key
router.Use(auth.RequireEither(jwtManager, apiKeyStore))

// Optional authentication (JWT or API Key if provided)
router.Use(auth.OptionalAuth(jwtManager, apiKeyStore))
```

### Current Route Configuration

- **Public Routes**: No authentication required
  - `/health` - Health check
  - `/login` - User login
  - `/api/keys/test` - Test API key

- **JWT + API Key Routes**: Either authentication method accepted
  - `/api/profile` - User profile
  - `/api/refresh` - Refresh token
  - `/api/user` - User endpoint
  - `/api/moderator` - Moderator endpoint
  - `/api/admin` - Admin endpoint
  - `/api/mixed` - Admin or Moderator endpoint

- **JWT Only Routes**: API key management (requires JWT)
  - `/api/keys` - Create/List API keys
  - `/api/keys/stats` - Get statistics
  - `/api/keys/{key}` - Get/Delete specific key
  - `/api/keys/{key}/revoke` - Revoke key

## Security Features

### Key Generation
- **Cryptographically secure**: Uses `crypto/rand` for key generation
- **Prefix**: All keys start with `ak_` for easy identification
- **Length**: 64-character hexadecimal keys (32 bytes)
- **Uniqueness**: Extremely low collision probability

### Access Control
- **Role-based**: API keys inherit user roles
- **User isolation**: Keys are tied to specific users
- **Revocation**: Keys can be deactivated or deleted
- **Expiration**: Automatic cleanup of expired keys

### Rate Limiting
- **Per-key**: Each key has independent rate limits
- **Time window**: 1-minute sliding window
- **Automatic cleanup**: Old rate limit data is cleaned up
- **Configurable**: Set custom limits per key

## Testing

### Quick Test
```bash
make test-apikey
```

### Manual Testing
1. **Get JWT token**: `POST /login`
2. **Create API key**: `POST /api/keys` (with JWT)
3. **Test API key**: `GET /api/keys/test` (with X-API-Key header)
4. **Use protected endpoint**: `GET /api/profile` (with X-API-Key header)

### Swagger UI
- **URL**: `http://localhost:8080/swagger/`
- **Authentication**: Use "Authorize" button to set JWT or API Key
- **JWT**: Enter `Bearer <token>`
- **API Key**: Enter `<api-key>` in X-API-Key field

## Error Handling

### Common Errors

| Error | Status | Description |
|-------|--------|-------------|
| `Authentication required` | 401 | No valid JWT or API key provided |
| `Invalid API key` | 401 | API key is invalid, expired, or inactive |
| `Rate limit exceeded` | 429 | API key has exceeded its rate limit |
| `Insufficient permissions` | 403 | API key doesn't have required roles |

### Error Response Format
```json
{
  "error": "Authentication required",
  "details": "Valid JWT token or API key required"
}
```

## Best Practices

### API Key Management
1. **Use descriptive names** for easy identification
2. **Set appropriate expiration times** for security
3. **Configure rate limits** based on expected usage
4. **Regularly rotate keys** for sensitive applications
5. **Monitor usage** through statistics endpoint

### Security
1. **Store keys securely** (environment variables, secret managers)
2. **Use HTTPS** in production
3. **Implement proper logging** for key usage
4. **Regularly audit** active keys
5. **Revoke unused keys** promptly

### Development
1. **Use JWT for user sessions** (shorter-lived)
2. **Use API keys for service-to-service** (longer-lived)
3. **Test both authentication methods** in your applications
4. **Handle both success and error cases** properly

## Examples

### Node.js Example
```javascript
const axios = require('axios');

// Using API Key
const apiKey = 'ak_9722e22f6bfc740b4b289a3c86a7496c42d9ab872233d75602604387bbcc30b0';

const response = await axios.get('http://localhost:8080/api/profile', {
  headers: {
    'X-API-Key': apiKey
  }
});
```

### Python Example
```python
import requests

api_key = 'ak_9722e22f6bfc740b4b289a3c86a7496c42d9ab872233d75602604387bbcc30b0'

response = requests.get(
    'http://localhost:8080/api/profile',
    headers={'X-API-Key': api_key}
)
```

### cURL Examples
```bash
# Test API key
curl -H "X-API-Key: ak_9722e22f6bfc740b4b289a3c86a7496c42d9ab872233d75602604387bbcc30b0" \
  http://localhost:8080/api/keys/test

# Access protected endpoint
curl -H "X-API-Key: ak_9722e22f6bfc740b4b289a3c86a7496c42d9ab872233d75602604387bbcc30b0" \
  http://localhost:8080/api/profile

# Admin endpoint
curl -H "X-API-Key: ak_9722e22f6bfc740b4b289a3c86a7496c42d9ab872233d75602604387bbcc30b0" \
  http://localhost:8080/api/admin
```

## Monitoring and Analytics

### Statistics Endpoint
```bash
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8080/api/keys/stats
```

**Response:**
```json
{
  "stats": {
    "total_keys": 5,
    "active_keys": 4,
    "expired_keys": 1,
    "inactive_keys": 0
  }
}
```

### Key Information
Each API key includes usage tracking:
- `created_at`: When the key was created
- `last_used_at`: When the key was last used
- `expires_at`: When the key expires
- `is_active`: Whether the key is currently active

This comprehensive API key system provides a robust, secure, and flexible authentication mechanism that complements the existing JWT system perfectly!
