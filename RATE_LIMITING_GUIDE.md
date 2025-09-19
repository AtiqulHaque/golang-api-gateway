# Rate Limiting and Throttling Guide

## Overview

The API Gateway implements comprehensive rate limiting and throttling using the **Token Bucket Algorithm** with support for both in-memory and Redis-based distributed rate limiting. This system provides fine-grained control over API usage and protects against abuse.

## Features

- **🪣 Token Bucket Algorithm**: Smooth rate limiting with burst capacity
- **🌐 Distributed Scaling**: Redis support for multi-instance deployments
- **🎯 Multiple Client Identification**: IP, JWT subject, API key, or user ID based
- **⚡ High Performance**: In-memory fallback with Redis for scaling
- **📊 Monitoring**: Real-time statistics and status endpoints
- **🔧 Configurable**: Flexible configuration per client type
- **🚦 HTTP 429 Responses**: Standard rate limit exceeded responses

## Algorithm: Token Bucket

The Token Bucket algorithm provides smooth rate limiting with burst capacity:

- **Capacity**: Maximum number of tokens in the bucket
- **Refill Rate**: Tokens added per second
- **Burst**: Allows short bursts up to capacity
- **Smooth**: Distributes requests evenly over time

### Example
- **Capacity**: 100 tokens
- **Refill Rate**: 10 tokens/second
- **Burst**: Can handle 100 requests immediately
- **Sustained**: Can handle 10 requests/second continuously

## Client Identification Methods

### 1. IP Address (Default)
```bash
# Rate limiting per IP address
curl http://localhost:8080/api/profile
```

### 2. JWT Subject
```bash
# Rate limiting per JWT user
curl -H "Authorization: Bearer <jwt-token>" http://localhost:8080/api/profile
```

### 3. API Key
```bash
# Rate limiting per API key
curl -H "X-API-Key: <api-key>" http://localhost:8080/api/profile
```

### 4. User ID
```bash
# Rate limiting per authenticated user
curl -H "Authorization: Bearer <jwt-token>" http://localhost:8080/api/profile
```

## Configuration

### Environment Variables

```bash
# Enable/disable rate limiting
RATE_LIMIT_ENABLED=true

# Client identification method
RATE_LIMIT_IDENTIFIER=ip  # ip, jwt, apikey, user

# Rate limiting parameters
RATE_LIMIT_CAPACITY=100
RATE_LIMIT_REFILL_RATE=10
RATE_LIMIT_WINDOW=1m

# Redis configuration (for distributed scaling)
RATE_LIMIT_USE_REDIS=false
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10

# Skip certain requests
RATE_LIMIT_SKIP_SUCCESS=false
RATE_LIMIT_SKIP_FAILED=false
```

### Configuration Examples

#### Basic IP-based Rate Limiting
```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_IDENTIFIER=ip
RATE_LIMIT_CAPACITY=100
RATE_LIMIT_REFILL_RATE=10
RATE_LIMIT_WINDOW=1m
```

#### JWT-based Rate Limiting
```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_IDENTIFIER=jwt
RATE_LIMIT_CAPACITY=50
RATE_LIMIT_REFILL_RATE=5
RATE_LIMIT_WINDOW=1m
```

#### Redis-based Distributed Rate Limiting
```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_IDENTIFIER=ip
RATE_LIMIT_CAPACITY=1000
RATE_LIMIT_REFILL_RATE=100
RATE_LIMIT_WINDOW=1m
RATE_LIMIT_USE_REDIS=true
REDIS_HOST=redis-cluster.example.com
REDIS_PORT=6379
REDIS_PASSWORD=secret
```

## API Endpoints

### Rate Limiting Management

#### Get Rate Limiting Statistics
```bash
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8080/api/ratelimit/stats
```

**Response:**
```json
{
  "stats": {
    "config": {
      "identifier": "ip",
      "capacity": 100,
      "refill_rate": 10,
      "window": "1m0s",
      "use_redis": false,
      "skip_successful": false,
      "skip_failed": false
    },
    "in_memory": {
      "buckets": 5
    }
  }
}
```

#### Test Rate Limiting
```bash
curl -X POST -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{"key": "test-client", "count": 1}' \
  http://localhost:8080/api/ratelimit/test
```

#### Get Client Status
```bash
curl -H "Authorization: Bearer <jwt-token>" \
  "http://localhost:8080/api/ratelimit/status?key=192.168.1.1"
```

#### Reset Client Rate Limit
```bash
curl -X POST -H "Authorization: Bearer <jwt-token>" \
  "http://localhost:8080/api/ratelimit/reset?key=192.168.1.1"
```

#### Get Rate Limit Headers
```bash
curl http://localhost:8080/api/ratelimit/headers
```

## HTTP Headers

### Request Headers
- **X-Forwarded-For**: Client IP for proxy scenarios
- **X-Real-IP**: Client IP for proxy scenarios
- **Authorization**: JWT token for user-based rate limiting
- **X-API-Key**: API key for key-based rate limiting

### Response Headers
- **X-RateLimit-Limit**: Maximum requests allowed
- **X-RateLimit-Remaining**: Remaining requests in current window
- **X-RateLimit-Reset**: Time when the rate limit resets (Unix timestamp)
- **Retry-After**: Seconds to wait before retrying (only on 429)

## Rate Limit Responses

### HTTP 429 - Too Many Requests
```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests",
  "retry_after": 60,
  "reset_time": "2025-09-19T16:30:00Z",
  "limit": 100,
  "remaining": 0
}
```

### Headers on 429 Response
```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1734567000
Retry-After: 60
Content-Type: application/json
```

## Use Cases

### 1. API Protection
```bash
# Protect against DDoS and abuse
RATE_LIMIT_IDENTIFIER=ip
RATE_LIMIT_CAPACITY=1000
RATE_LIMIT_REFILL_RATE=100
```

### 2. User-based Limits
```bash
# Different limits per user
RATE_LIMIT_IDENTIFIER=jwt
RATE_LIMIT_CAPACITY=100
RATE_LIMIT_REFILL_RATE=10
```

### 3. API Key Tiers
```bash
# Different limits per API key
RATE_LIMIT_IDENTIFIER=apikey
RATE_LIMIT_CAPACITY=10000
RATE_LIMIT_REFILL_RATE=1000
```

### 4. Microservice Protection
```bash
# Protect internal services
RATE_LIMIT_IDENTIFIER=ip
RATE_LIMIT_CAPACITY=100
RATE_LIMIT_REFILL_RATE=20
RATE_LIMIT_SKIP_SUCCESS=true
```

## Monitoring and Alerting

### Key Metrics
- **Request Rate**: Requests per second per client
- **Rate Limit Hits**: Number of 429 responses
- **Bucket Utilization**: Percentage of capacity used
- **Client Distribution**: Number of unique clients

### Monitoring Endpoints
```bash
# Get current statistics
curl http://localhost:8080/api/ratelimit/stats

# Monitor specific client
curl "http://localhost:8080/api/ratelimit/status?key=192.168.1.1"
```

### Logging
Rate limiting events are logged with:
- Client identifier
- Request timestamp
- Rate limit decision
- Remaining tokens
- Reset time

## Performance Considerations

### In-Memory Rate Limiting
- **Pros**: Fast, no network overhead
- **Cons**: Not shared between instances
- **Use Case**: Single instance deployments

### Redis-based Rate Limiting
- **Pros**: Shared between instances, persistent
- **Cons**: Network overhead, Redis dependency
- **Use Case**: Multi-instance deployments

### Optimization Tips
1. **Use appropriate capacity**: Balance burst vs sustained rate
2. **Choose right identifier**: IP for DDoS protection, JWT for user limits
3. **Monitor performance**: Watch Redis latency and memory usage
4. **Tune refill rate**: Match your expected traffic patterns

## Testing

### Quick Test
```bash
make test-ratelimit
```

### Manual Testing
```bash
# Test basic rate limiting
for i in {1..10}; do
  curl -w "HTTP_STATUS:%{http_code}\n" http://localhost:8080/health
  sleep 0.1
done

# Test with different IPs
curl -H "X-Forwarded-For: 192.168.1.1" http://localhost:8080/health
curl -H "X-Forwarded-For: 192.168.1.2" http://localhost:8080/health
```

### Load Testing
```bash
# Test with Apache Bench
ab -n 1000 -c 10 http://localhost:8080/health

# Test with curl in parallel
for i in {1..100}; do
  curl http://localhost:8080/health &
done
wait
```

## Troubleshooting

### Common Issues

#### Rate Limiting Not Working
1. Check if `RATE_LIMIT_ENABLED=true`
2. Verify configuration values
3. Check server logs for errors
4. Test with `/api/ratelimit/headers` endpoint

#### Redis Connection Issues
1. Verify Redis is running
2. Check connection parameters
3. Test Redis connectivity
4. Check Redis logs

#### Unexpected 429 Responses
1. Check rate limit configuration
2. Verify client identification
3. Check if multiple clients share same identifier
4. Review rate limit statistics

### Debug Commands
```bash
# Check rate limiting status
curl http://localhost:8080/api/ratelimit/headers

# Get statistics
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/ratelimit/stats

# Test specific client
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/ratelimit/status?key=192.168.1.1"
```

## Best Practices

### Configuration
1. **Start conservative**: Begin with lower limits and increase
2. **Monitor usage**: Watch for false positives
3. **Use appropriate identifiers**: Match your use case
4. **Set reasonable windows**: Balance responsiveness vs accuracy

### Monitoring
1. **Track metrics**: Monitor rate limit hits and client behavior
2. **Set up alerts**: Alert on high rate limit hit rates
3. **Regular reviews**: Adjust limits based on usage patterns
4. **Log analysis**: Analyze logs for abuse patterns

### Security
1. **Protect Redis**: Secure Redis connections in production
2. **Monitor abuse**: Watch for unusual patterns
3. **Regular updates**: Keep rate limiting logic updated
4. **Backup configs**: Keep configuration backups

## Examples

### Node.js Client
```javascript
const axios = require('axios');

async function makeRequest() {
  try {
    const response = await axios.get('http://localhost:8080/api/profile', {
      headers: {
        'Authorization': 'Bearer ' + token
      }
    });
    return response.data;
  } catch (error) {
    if (error.response && error.response.status === 429) {
      const retryAfter = error.response.headers['retry-after'];
      console.log(`Rate limited. Retry after ${retryAfter} seconds`);
      await new Promise(resolve => setTimeout(resolve, retryAfter * 1000));
      return makeRequest(); // Retry
    }
    throw error;
  }
}
```

### Python Client
```python
import requests
import time

def make_request():
    try:
        response = requests.get(
            'http://localhost:8080/api/profile',
            headers={'Authorization': f'Bearer {token}'}
        )
        return response.json()
    except requests.exceptions.HTTPError as e:
        if e.response.status_code == 429:
            retry_after = float(e.response.headers.get('Retry-After', 60))
            print(f"Rate limited. Retrying after {retry_after} seconds")
            time.sleep(retry_after)
            return make_request()  # Retry
        raise
```

### cURL Examples
```bash
# Basic request
curl http://localhost:8080/api/profile

# With JWT
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/profile

# With API key
curl -H "X-API-Key: <key>" http://localhost:8080/api/profile

# Check rate limit headers
curl -I http://localhost:8080/api/profile
```

This comprehensive rate limiting system provides robust protection against abuse while maintaining high performance and flexibility for different deployment scenarios!
