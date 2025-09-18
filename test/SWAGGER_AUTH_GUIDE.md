# Swagger UI Authentication Guide

## How to Use JWT Authentication in Swagger UI

### Step 1: Access Swagger UI
1. Open your browser and go to: `http://localhost:8080/swagger/`
2. You should see the API documentation with all available endpoints

### Step 2: Get a JWT Token
1. First, you need to get a JWT token by logging in
2. Use the **POST /login** endpoint with one of these credentials:
   - **Admin**: `admin` / `admin123`
   - **User**: `user` / `user123`
   - **Moderator**: `moderator` / `mod123`

### Step 3: Configure Authentication in Swagger UI
1. Click the **"Authorize"** button at the top right of the Swagger UI
2. In the **BearerAuth** section, enter your JWT token in the format: `Bearer <your-token>`
   - **Important**: Make sure to include the word "Bearer" followed by a space before your token
   - Example: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
3. Click **"Authorize"** to save the token
4. Click **"Close"** to return to the API documentation

**Note**: If you get "authorization header must be in format 'Bearer <token>'" error, make sure you've included the "Bearer " prefix in the authorization field.

### Step 4: Test Protected Endpoints
Now you can test any protected endpoint:
- **GET /api/profile** - Get user profile
- **GET /api/user** - User-only endpoint
- **GET /api/moderator** - Moderator-only endpoint (requires moderator role)
- **GET /api/admin** - Admin-only endpoint (requires admin role)
- **GET /api/mixed** - Admin or Moderator endpoint

### Example JWT Token Format
```
Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGktdXNlcnMiLCJleHAiOjE3MzQ2MjQwMDAsImlhdCI6MTczNDUzNzYwMCwiaXNzIjoiYXBpLWdhdGV3YXkiLCJyb2xlcyI6WyJhZG1pbiJdLCJzdWIiOiJhZG1pbiJ9.example-signature
```

### Troubleshooting
- **"Authentication required"** error: Make sure you've set the JWT token in the Authorize dialog
- **"authorization header must be in format 'Bearer <token>'"** error: 
  - Make sure you've included "Bearer " (with space) before your token
  - Example: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
  - Don't just paste the token without the "Bearer " prefix
- **"Invalid token"** error: Your token may have expired, get a new one from the login endpoint
- **"Insufficient permissions"** error: Your user role doesn't have access to this endpoint

### Common Mistakes
1. **Missing "Bearer " prefix**: Always include "Bearer " before your token
2. **Extra spaces**: Don't add extra spaces around the token
3. **Expired token**: JWT tokens expire after 24 hours, get a new one if needed
4. **Wrong format**: The format must be exactly `Bearer <token>` (case sensitive)

### Available User Roles
- **admin**: Full access to all endpoints
- **moderator**: Access to moderator and user endpoints
- **user**: Access to user endpoints only

### Quick Test Commands
```bash
# Get a token
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Use the token in subsequent requests
curl -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  http://localhost:8080/api/profile
```
