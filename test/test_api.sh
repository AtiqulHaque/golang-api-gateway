#!/bin/bash

# API Gateway Test Script
# This script demonstrates the authentication and authorization features

BASE_URL="http://localhost:8080"

echo "🚀 Testing API Gateway Authentication and Authorization"
echo "======================================================"

# Test health check (no auth required)
echo -e "\n1. Testing health check (no auth required):"
curl -s "$BASE_URL/health" | jq .

# Test Swagger documentation endpoints
echo -e "\n2. Testing Swagger documentation endpoints:"
echo -e "\n   a) Swagger JSON:"
curl -s "$BASE_URL/swagger/doc.json" | jq '.info'

echo -e "\n   b) Swagger UI (checking if accessible):"
curl -s -I "$BASE_URL/swagger/" | head -1

# Test login with admin user
echo -e "\n3. Testing login with admin user:"
ADMIN_RESPONSE=$(curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}')
echo "$ADMIN_RESPONSE" | jq .

# Extract admin token
ADMIN_TOKEN=$(echo "$ADMIN_RESPONSE" | jq -r '.token')
echo "Admin token: ${ADMIN_TOKEN:0:50}..."

# Test login with regular user
echo -e "\n4. Testing login with regular user:"
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "password": "user123"}')
echo "$USER_RESPONSE" | jq .

# Extract user token
USER_TOKEN=$(echo "$USER_RESPONSE" | jq -r '.token')
echo "User token: ${USER_TOKEN:0:50}..."

# Test protected endpoints with admin token
echo -e "\n5. Testing protected endpoints with admin token:"

echo -e "\n   a) User profile:"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "$BASE_URL/api/profile" | jq .

echo -e "\n   b) User endpoint:"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "$BASE_URL/api/user" | jq .

echo -e "\n   c) Admin endpoint:"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "$BASE_URL/api/admin" | jq .

echo -e "\n   d) Mixed roles endpoint:"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "$BASE_URL/api/mixed" | jq .

# Test protected endpoints with user token
echo -e "\n6. Testing protected endpoints with user token:"

echo -e "\n   a) User profile:"
curl -s -H "Authorization: Bearer $USER_TOKEN" "$BASE_URL/api/profile" | jq .

echo -e "\n   b) User endpoint:"
curl -s -H "Authorization: Bearer $USER_TOKEN" "$BASE_URL/api/user" | jq .

echo -e "\n   c) Admin endpoint (should fail):"
curl -s -H "Authorization: Bearer $USER_TOKEN" "$BASE_URL/api/admin" | jq .

echo -e "\n   d) Mixed roles endpoint (should fail):"
curl -s -H "Authorization: Bearer $USER_TOKEN" "$BASE_URL/api/mixed" | jq .

# Test without token
echo -e "\n7. Testing protected endpoint without token (should fail):"
curl -s "$BASE_URL/api/profile" | jq .

# Test with invalid token
echo -e "\n8. Testing protected endpoint with invalid token (should fail):"
curl -s -H "Authorization: Bearer invalid-token" "$BASE_URL/api/profile" | jq .

echo -e "\n✅ Test completed!"
