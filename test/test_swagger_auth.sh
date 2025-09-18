#!/bin/bash

# Test script to demonstrate Swagger UI authentication
echo "🔐 Testing Swagger UI Authentication"
echo "=================================="

# Get a JWT token
echo "1. Getting JWT token..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

# Extract token from response
TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ Failed to get token"
    exit 1
fi

echo "✅ Token obtained: ${TOKEN:0:50}..."

# Test with correct Bearer format
echo ""
echo "2. Testing with correct Bearer format..."
RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/profile)

if echo "$RESPONSE" | grep -q "username"; then
    echo "✅ Bearer format works correctly"
    echo "Response: $RESPONSE"
else
    echo "❌ Bearer format failed"
    echo "Response: $RESPONSE"
fi

# Test with incorrect format (without Bearer)
echo ""
echo "3. Testing with incorrect format (without Bearer)..."
RESPONSE=$(curl -s -H "Authorization: $TOKEN" http://localhost:8080/api/profile)

if echo "$RESPONSE" | grep -q "authorization header must be in format"; then
    echo "✅ Server correctly rejects token without Bearer prefix"
    echo "Response: $RESPONSE"
else
    echo "❌ Server should have rejected token without Bearer prefix"
    echo "Response: $RESPONSE"
fi

echo ""
echo "📝 Instructions for Swagger UI:"
echo "1. Go to http://localhost:8080/swagger/"
echo "2. Click 'Authorize' button"
echo "3. Enter: Bearer $TOKEN"
echo "4. Click 'Authorize' and 'Close'"
echo "5. Test any protected endpoint"
