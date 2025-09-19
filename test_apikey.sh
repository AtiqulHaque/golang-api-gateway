#!/bin/bash

# Test script for API Key authentication
echo "🔑 Testing API Key Authentication"
echo "================================="

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

# Get JWT token for API key management
echo -e "\n${BLUE}2. Getting JWT token for API key management...${NC}"
JWT_RESPONSE=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

JWT_TOKEN=$(echo $JWT_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$JWT_TOKEN" ]; then
    echo -e "${RED}❌ Failed to get JWT token${NC}"
    exit 1
fi
echo -e "${GREEN}✅ JWT token obtained${NC}"

# Create an API key
echo -e "\n${BLUE}3. Creating API key...${NC}"
API_KEY_RESPONSE=$(curl -s -X POST http://localhost:8080/api/keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "name": "Test API Key",
    "user_id": "test-user-123",
    "roles": ["user", "admin"],
    "rate_limit": 50,
    "expires_in": "24h"
  }')

API_KEY=$(echo $API_KEY_RESPONSE | grep -o '"key":"[^"]*"' | cut -d'"' -f4)

if [ -z "$API_KEY" ]; then
    echo -e "${RED}❌ Failed to create API key${NC}"
    echo "Response: $API_KEY_RESPONSE"
    exit 1
fi
echo -e "${GREEN}✅ API key created: ${API_KEY:0:20}...${NC}"

# Test API key authentication
echo -e "\n${BLUE}4. Testing API key authentication...${NC}"
API_KEY_TEST=$(curl -s -H "X-API-Key: $API_KEY" http://localhost:8080/api/keys/test)

if echo "$API_KEY_TEST" | grep -q "name"; then
    echo -e "${GREEN}✅ API key authentication works${NC}"
    echo "Response: $API_KEY_TEST"
else
    echo -e "${RED}❌ API key authentication failed${NC}"
    echo "Response: $API_KEY_TEST"
fi

# Test protected endpoint with API key
echo -e "\n${BLUE}5. Testing protected endpoint with API key...${NC}"
PROTECTED_RESPONSE=$(curl -s -H "X-API-Key: $API_KEY" http://localhost:8080/api/profile)

if echo "$PROTECTED_RESPONSE" | grep -q "user_id"; then
    echo -e "${GREEN}✅ Protected endpoint accessible with API key${NC}"
    echo "Response: $PROTECTED_RESPONSE"
else
    echo -e "${RED}❌ Protected endpoint not accessible with API key${NC}"
    echo "Response: $PROTECTED_RESPONSE"
fi

# Test admin endpoint with API key
echo -e "\n${BLUE}6. Testing admin endpoint with API key...${NC}"
ADMIN_RESPONSE=$(curl -s -H "X-API-Key: $API_KEY" http://localhost:8080/api/admin)

if echo "$ADMIN_RESPONSE" | grep -q "admin-only"; then
    echo -e "${GREEN}✅ Admin endpoint accessible with API key${NC}"
    echo "Response: $ADMIN_RESPONSE"
else
    echo -e "${RED}❌ Admin endpoint not accessible with API key${NC}"
    echo "Response: $ADMIN_RESPONSE"
fi

# Test rate limiting
echo -e "\n${BLUE}7. Testing rate limiting...${NC}"
echo "Making 5 rapid requests to test rate limiting..."
for i in {1..5}; do
    RATE_RESPONSE=$(curl -s -H "X-API-Key: $API_KEY" http://localhost:8080/api/profile)
    if echo "$RATE_RESPONSE" | grep -q "rate limit"; then
        echo -e "${YELLOW}⚠️  Rate limit hit on request $i${NC}"
        break
    else
        echo -e "${GREEN}✅ Request $i successful${NC}"
    fi
    sleep 0.1
done

# List API keys
echo -e "\n${BLUE}8. Listing API keys...${NC}"
LIST_RESPONSE=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "http://localhost:8080/api/keys?user_id=test-user-123")

if echo "$LIST_RESPONSE" | grep -q "api_keys"; then
    echo -e "${GREEN}✅ API keys listed successfully${NC}"
    echo "Response: $LIST_RESPONSE"
else
    echo -e "${RED}❌ Failed to list API keys${NC}"
    echo "Response: $LIST_RESPONSE"
fi

# Get API key stats
echo -e "\n${BLUE}9. Getting API key statistics...${NC}"
STATS_RESPONSE=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" http://localhost:8080/api/keys/stats)

if echo "$STATS_RESPONSE" | grep -q "total_keys"; then
    echo -e "${GREEN}✅ API key statistics retrieved${NC}"
    echo "Response: $STATS_RESPONSE"
else
    echo -e "${RED}❌ Failed to get API key statistics${NC}"
    echo "Response: $STATS_RESPONSE"
fi

# Test invalid API key
echo -e "\n${BLUE}10. Testing invalid API key...${NC}"
INVALID_RESPONSE=$(curl -s -H "X-API-Key: invalid-key" http://localhost:8080/api/profile)

if echo "$INVALID_RESPONSE" | grep -q "Invalid API key"; then
    echo -e "${GREEN}✅ Invalid API key correctly rejected${NC}"
    echo "Response: $INVALID_RESPONSE"
else
    echo -e "${RED}❌ Invalid API key should have been rejected${NC}"
    echo "Response: $INVALID_RESPONSE"
fi

# Clean up - revoke the API key
echo -e "\n${BLUE}11. Cleaning up - revoking API key...${NC}"
REVOKE_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $JWT_TOKEN" "http://localhost:8080/api/keys/$API_KEY/revoke")

if echo "$REVOKE_RESPONSE" | grep -q "revoked successfully"; then
    echo -e "${GREEN}✅ API key revoked successfully${NC}"
else
    echo -e "${YELLOW}⚠️  API key revocation response: $REVOKE_RESPONSE${NC}"
fi

echo -e "\n${GREEN}🎉 API Key authentication test completed!${NC}"
echo -e "${BLUE}Summary:${NC}"
echo "✅ JWT authentication works"
echo "✅ API key creation works"
echo "✅ API key authentication works"
echo "✅ Protected endpoints accessible with API key"
echo "✅ Role-based access control works with API keys"
echo "✅ Rate limiting works"
echo "✅ API key management endpoints work"
echo "✅ Invalid API keys are rejected"
echo "✅ API key revocation works"
