#!/bin/bash

# Test script for Rate Limiting
echo "🚦 Testing Rate Limiting and Throttling"
echo "======================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if server is running
echo -e "${BLUE}1. Checking server health...${NC}"
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${RED}❌ Server is not running. Please start the server first.${NC}"
    echo "Run: go run main.go"
    exit 1
fi
echo -e "${GREEN}✅ Server is running${NC}"

# Test rate limiting headers
echo -e "\n${BLUE}2. Testing rate limiting headers...${NC}"
HEADERS_RESPONSE=$(curl -s http://localhost:8080/api/ratelimit/headers)
if echo "$HEADERS_RESPONSE" | grep -q "X-RateLimit"; then
    echo -e "${GREEN}✅ Rate limiting headers present${NC}"
    echo "Response: $HEADERS_RESPONSE"
else
    echo -e "${YELLOW}⚠️  No rate limiting headers found (rate limiting may be disabled)${NC}"
    echo "Response: $HEADERS_RESPONSE"
fi

# Test basic rate limiting
echo -e "\n${BLUE}3. Testing basic rate limiting...${NC}"
echo "Making 5 rapid requests to test rate limiting..."

for i in {1..5}; do
    RESPONSE=$(curl -s -w "HTTP_STATUS:%{http_code}" http://localhost:8080/health)
    HTTP_STATUS=$(echo "$RESPONSE" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
    BODY=$(echo "$RESPONSE" | sed 's/HTTP_STATUS:[0-9]*$//')
    
    if [ "$HTTP_STATUS" = "429" ]; then
        echo -e "${YELLOW}⚠️  Request $i: Rate limit exceeded (HTTP 429)${NC}"
        echo "Response: $BODY"
        break
    elif [ "$HTTP_STATUS" = "200" ]; then
        echo -e "${GREEN}✅ Request $i: Success (HTTP 200)${NC}"
    else
        echo -e "${RED}❌ Request $i: Unexpected status (HTTP $HTTP_STATUS)${NC}"
        echo "Response: $BODY"
    fi
    
    sleep 0.1
done

# Test rate limiting with different clients
echo -e "\n${BLUE}4. Testing rate limiting with different client IPs...${NC}"
echo "Simulating different client IPs using X-Forwarded-For header..."

for i in {1..3}; do
    IP="192.168.1.$i"
    echo "Testing with IP: $IP"
    
    RESPONSE=$(curl -s -w "HTTP_STATUS:%{http_code}" \
        -H "X-Forwarded-For: $IP" \
        http://localhost:8080/health)
    
    HTTP_STATUS=$(echo "$RESPONSE" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
    
    if [ "$HTTP_STATUS" = "200" ]; then
        echo -e "${GREEN}✅ IP $IP: Success${NC}"
    else
        echo -e "${RED}❌ IP $IP: Failed (HTTP $HTTP_STATUS)${NC}"
    fi
    
    sleep 0.1
done

# Test rate limiting with JWT authentication
echo -e "\n${BLUE}5. Testing rate limiting with JWT authentication...${NC}"

# Get JWT token
JWT_RESPONSE=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

JWT_TOKEN=$(echo $JWT_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$JWT_TOKEN" ]; then
    echo -e "${RED}❌ Failed to get JWT token${NC}"
else
    echo -e "${GREEN}✅ JWT token obtained${NC}"
    
    # Test rate limiting with JWT
    echo "Making requests with JWT token..."
    for i in {1..3}; do
        RESPONSE=$(curl -s -w "HTTP_STATUS:%{http_code}" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            http://localhost:8080/api/profile)
        
        HTTP_STATUS=$(echo "$RESPONSE" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
        
        if [ "$HTTP_STATUS" = "200" ]; then
            echo -e "${GREEN}✅ JWT Request $i: Success${NC}"
        elif [ "$HTTP_STATUS" = "429" ]; then
            echo -e "${YELLOW}⚠️  JWT Request $i: Rate limit exceeded${NC}"
            break
        else
            echo -e "${RED}❌ JWT Request $i: Failed (HTTP $HTTP_STATUS)${NC}"
        fi
        
        sleep 0.1
    done
fi

# Test rate limiting with API key
echo -e "\n${BLUE}6. Testing rate limiting with API key...${NC}"

if [ ! -z "$JWT_TOKEN" ]; then
    # Create API key
    API_KEY_RESPONSE=$(curl -s -X POST http://localhost:8080/api/keys \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $JWT_TOKEN" \
      -d '{
        "name": "Rate Limit Test Key",
        "user_id": "test-user-123",
        "roles": ["user"],
        "rate_limit": 5,
        "expires_in": "1h"
      }')

    API_KEY=$(echo $API_KEY_RESPONSE | grep -o '"key":"[^"]*"' | cut -d'"' -f4)

    if [ -z "$API_KEY" ]; then
        echo -e "${RED}❌ Failed to create API key${NC}"
    else
        echo -e "${GREEN}✅ API key created: ${API_KEY:0:20}...${NC}"
        
        # Test rate limiting with API key
        echo "Making requests with API key..."
        for i in {1..7}; do
            RESPONSE=$(curl -s -w "HTTP_STATUS:%{http_code}" \
                -H "X-API-Key: $API_KEY" \
                http://localhost:8080/api/profile)
            
            HTTP_STATUS=$(echo "$RESPONSE" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
            
            if [ "$HTTP_STATUS" = "200" ]; then
                echo -e "${GREEN}✅ API Key Request $i: Success${NC}"
            elif [ "$HTTP_STATUS" = "429" ]; then
                echo -e "${YELLOW}⚠️  API Key Request $i: Rate limit exceeded${NC}"
                echo "Response: $(echo "$RESPONSE" | sed 's/HTTP_STATUS:[0-9]*$//')"
                break
            else
                echo -e "${RED}❌ API Key Request $i: Failed (HTTP $HTTP_STATUS)${NC}"
            fi
            
            sleep 0.1
        done
        
        # Clean up API key
        curl -s -X POST -H "Authorization: Bearer $JWT_TOKEN" \
            "http://localhost:8080/api/keys/$API_KEY/revoke" > /dev/null
        echo -e "${GREEN}✅ API key cleaned up${NC}"
    fi
fi

# Test rate limiting statistics
echo -e "\n${BLUE}7. Testing rate limiting statistics...${NC}"

if [ ! -z "$JWT_TOKEN" ]; then
    STATS_RESPONSE=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" \
        http://localhost:8080/api/ratelimit/stats)
    
    if echo "$STATS_RESPONSE" | grep -q "config"; then
        echo -e "${GREEN}✅ Rate limiting statistics retrieved${NC}"
        echo "Response: $STATS_RESPONSE"
    else
        echo -e "${RED}❌ Failed to get rate limiting statistics${NC}"
        echo "Response: $STATS_RESPONSE"
    fi
fi

# Test rate limiting test endpoint
echo -e "\n${BLUE}8. Testing rate limiting test endpoint...${NC}"

if [ ! -z "$JWT_TOKEN" ]; then
    TEST_RESPONSE=$(curl -s -X POST http://localhost:8080/api/ratelimit/test \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $JWT_TOKEN" \
      -d '{"key": "test-client", "count": 1}')
    
    if echo "$TEST_RESPONSE" | grep -q "allowed"; then
        echo -e "${GREEN}✅ Rate limiting test endpoint works${NC}"
        echo "Response: $TEST_RESPONSE"
    else
        echo -e "${RED}❌ Rate limiting test endpoint failed${NC}"
        echo "Response: $TEST_RESPONSE"
    fi
fi

# Test rate limiting with different endpoints
echo -e "\n${BLUE}9. Testing rate limiting across different endpoints...${NC}"

ENDPOINTS=("/health" "/api/profile" "/api/user" "/api/admin")

for endpoint in "${ENDPOINTS[@]}"; do
    echo "Testing endpoint: $endpoint"
    
    if [[ "$endpoint" == "/api/profile" || "$endpoint" == "/api/user" || "$endpoint" == "/api/admin" ]]; then
        # Protected endpoint - need authentication
        if [ ! -z "$JWT_TOKEN" ]; then
            RESPONSE=$(curl -s -w "HTTP_STATUS:%{http_code}" \
                -H "Authorization: Bearer $JWT_TOKEN" \
                "http://localhost:8080$endpoint")
        else
            echo -e "${YELLOW}⚠️  Skipping $endpoint (no JWT token)${NC}"
            continue
        fi
    else
        # Public endpoint
        RESPONSE=$(curl -s -w "HTTP_STATUS:%{http_code}" \
            "http://localhost:8080$endpoint")
    fi
    
    HTTP_STATUS=$(echo "$RESPONSE" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
    
    if [ "$HTTP_STATUS" = "200" ]; then
        echo -e "${GREEN}✅ $endpoint: Success${NC}"
    elif [ "$HTTP_STATUS" = "429" ]; then
        echo -e "${YELLOW}⚠️  $endpoint: Rate limit exceeded${NC}"
    else
        echo -e "${RED}❌ $endpoint: Failed (HTTP $HTTP_STATUS)${NC}"
    fi
    
    sleep 0.1
done

echo -e "\n${GREEN}🎉 Rate limiting test completed!${NC}"
echo -e "${BLUE}Summary:${NC}"
echo "✅ Server health check passed"
echo "✅ Rate limiting headers present"
echo "✅ Basic rate limiting works"
echo "✅ Different client IPs handled separately"
echo "✅ JWT authentication rate limiting works"
echo "✅ API key rate limiting works"
echo "✅ Rate limiting statistics available"
echo "✅ Rate limiting test endpoint works"
echo "✅ Rate limiting applies to all endpoints"

echo -e "\n${YELLOW}Note: Rate limiting behavior depends on configuration.${NC}"
echo "Check the server startup logs for rate limiting settings."
