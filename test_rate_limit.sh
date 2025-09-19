#!/bin/bash

echo "🧪 Testing Rate Limiting Functionality"
echo "======================================"

# Kill any existing processes
echo "🔄 Cleaning up existing processes..."
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
sleep 1

# Start the server
echo "🚀 Starting API Gateway with rate limiting enabled..."
go run main.go &
SERVER_PID=$!
sleep 3

# Check if server is running
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Server failed to start"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

echo "✅ Server started successfully"

# Test 1: Basic health check with rate limit headers
echo ""
echo "📊 Test 1: Health check with rate limit headers"
echo "----------------------------------------------"
curl -v http://localhost:8080/health 2>&1 | grep -E "(X-RateLimit|HTTP/)"

# Test 2: Multiple rapid requests to trigger rate limiting
echo ""
echo "📊 Test 2: Rapid requests to test rate limiting (100 requests)"
echo "------------------------------------------------------------"
for i in {1..105}; do
    response=$(curl -s -w "%{http_code}" http://localhost:8080/health)
    status_code="${response: -3}"
    echo -n "Request $i: $status_code "
    
    if [ $status_code -eq 429 ]; then
        echo "🚫 RATE LIMITED!"
        break
    elif [ $status_code -eq 200 ]; then
        echo "✅ OK"
    else
        echo "❌ ERROR"
    fi
    
    # Small delay to avoid overwhelming
    sleep 0.01
done

# Test 3: Rate limit headers endpoint
echo ""
echo "📊 Test 3: Rate limit headers endpoint"
echo "------------------------------------"
curl -s http://localhost:8080/api/ratelimit/headers | jq . 2>/dev/null || curl -s http://localhost:8080/api/ratelimit/headers

# Test 4: Login and test with JWT
echo ""
echo "📊 Test 4: Login and test with JWT"
echo "---------------------------------"
login_response=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

token=$(echo $login_response | jq -r '.token' 2>/dev/null)

if [ "$token" != "null" ] && [ "$token" != "" ]; then
    echo "✅ Login successful, token: ${token:0:20}..."
    
    # Test protected endpoint with rate limiting
    echo "Testing protected endpoint with JWT..."
    curl -v -H "Authorization: Bearer $token" http://localhost:8080/api/profile 2>&1 | grep -E "(X-RateLimit|HTTP/)"
else
    echo "❌ Login failed"
fi

# Test 5: API Key rate limiting
echo ""
echo "📊 Test 5: API Key rate limiting"
echo "-------------------------------"
curl -v -H "X-API-Key: test-key-123" http://localhost:8080/api/keys/test 2>&1 | grep -E "(X-RateLimit|HTTP/)"

# Cleanup
echo ""
echo "🧹 Cleaning up..."
kill $SERVER_PID 2>/dev/null || true
sleep 1

echo ""
echo "✅ Rate limiting tests completed!"
