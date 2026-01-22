# Session-Based Authentication System

## Overview

NeuronIP now supports a complete session-based authentication system with cookie-based sessions, refresh token rotation, and database selection. This system is production-ready, secure, and backward compatible with API key authentication.

## Architecture

### Components

1. **Session Manager** (`api/internal/session/manager.go`)
   - Creates and validates sessions
   - Manages refresh token rotation
   - Handles session revocation
   - Stores database preference per session

2. **Session Middleware** (`api/internal/session/middleware.go`)
   - Validates cookies on each request
   - Automatically refreshes expired access tokens
   - Falls back to API key authentication
   - Adds session to request context

3. **Auth Service** (`api/internal/auth/service.go`)
   - `LoginWithUsername()` - Username/password authentication
   - `RegisterWithUsername()` - User registration
   - `GetCurrentUser()` - Get current user from session

4. **Auth Handlers** (`api/internal/handlers/auth_enhanced.go`)
   - `POST /api/v1/auth/login` - Login endpoint
   - `POST /api/v1/auth/register` - Registration endpoint
   - `GET /api/v1/auth/me` - Get current user
   - `POST /api/v1/auth/logout` - Logout endpoint
   - `POST /api/v1/auth/refresh` - Refresh token endpoint

5. **Session Cleanup** (`api/internal/session/cleanup.go`)
   - Automatic cleanup of expired sessions
   - Removes old refresh tokens
   - Runs hourly in background

6. **Validation** (`api/internal/session/validation.go`)
   - Input validation for all session operations
   - Database name validation
   - Session ID format validation

## Database Schema

### Sessions Table

```sql
CREATE TABLE neuronip.sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES neuronip.users(id),
    database_name TEXT NOT NULL DEFAULT 'neuronip',
    created_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    user_agent_hash TEXT,
    ip_hash TEXT
);
```

### Refresh Tokens Table

```sql
CREATE TABLE neuronip.refresh_tokens (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES neuronip.sessions(id),
    token_hash TEXT NOT NULL UNIQUE,
    rotated_from UUID REFERENCES neuronip.refresh_tokens(id),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL
);
```

### Failed Login Attempts Table

```sql
CREATE TABLE neuronip.failed_login_attempts (
    id UUID PRIMARY KEY,
    identifier TEXT NOT NULL,
    ip_address TEXT,
    attempted_at TIMESTAMPTZ NOT NULL
);
```

## Security Features

### Cookie Security

- **HttpOnly**: Prevents JavaScript access
- **Secure**: Only sent over HTTPS (configurable)
- **SameSite**: CSRF protection (Lax/Strict/None)
- **Domain**: Configurable cookie domain

### Token Security

- **Refresh Token Rotation**: Old tokens revoked on use
- **Token Reuse Detection**: All sessions revoked if reuse detected
- **Hashed Storage**: Tokens stored as SHA256 hashes
- **Expiration**: Configurable TTL for access and refresh tokens

### Input Validation

- Username: 3-100 characters
- Password: 6-1000 characters
- Database name: Validated against allowed list
- Session ID: Format validation
- IP and User Agent: Sanitized and hashed

### Rate Limiting

- Failed login attempt tracking
- Automatic cleanup of old attempts
- Configurable rate limits

## Configuration

### Environment Variables

```bash
# Session Configuration
SESSION_ACCESS_TTL=15m          # Access token TTL (default: 15 minutes)
SESSION_REFRESH_TTL=168h         # Refresh token TTL (default: 7 days)
SESSION_COOKIE_DOMAIN=           # Cookie domain (empty for localhost)
SESSION_COOKIE_SECURE=false      # Secure cookies (true for production)
SESSION_COOKIE_SAME_SITE=Lax     # SameSite policy

# NeuronAI Demo Database
NEURONAI_DEMO_HOST=localhost
NEURONAI_DEMO_PORT=5432
NEURONAI_DEMO_USER=neurondb
NEURONAI_DEMO_PASSWORD=neurondb
NEURONAI_DEMO_DATABASE=neuronai-demo
```

## API Endpoints

### POST /api/v1/auth/login

Login with username and password.

**Request:**
```json
{
  "username": "user123",
  "password": "password123",
  "database": "neuronip"  // Optional: "neuronip" or "neuronai-demo"
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "user123",
    "name": "User Name",
    "role": "analyst"
  },
  "database": "neuronip",
  "token": "session-id"  // For backward compatibility
}
```

**Cookies Set:**
- `access_token`: Session ID (short-lived)
- `refresh_token`: Refresh token (long-lived)

### POST /api/v1/auth/register

Register a new user.

**Request:**
```json
{
  "username": "newuser",
  "password": "password123",
  "database": "neuronip"  // Optional
}
```

**Response:** Same as login

### GET /api/v1/auth/me

Get current authenticated user.

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "User Name",
    "role": "analyst"
  },
  "database": "neuronip"
}
```

### POST /api/v1/auth/logout

Logout and revoke session.

**Response:** 204 No Content

### POST /api/v1/auth/refresh

Refresh access token using refresh token.

**Response:**
```json
{
  "access_token": "new-session-id",
  "refresh_token": "new-refresh-token",
  "database": "neuronip"
}
```

## Frontend Integration

### Login Flow

1. User enters username/password and selects database
2. Frontend calls `POST /api/v1/auth/login`
3. Backend creates session and sets cookies
4. Frontend redirects to dashboard
5. All subsequent requests include cookies automatically

### Authentication Check

- `AuthGuard` component calls `GET /api/v1/auth/me`
- If session invalid, redirects to `/login`
- If session valid, renders protected content

### Logout Flow

1. User clicks logout
2. Frontend calls `POST /api/v1/auth/logout`
3. Backend revokes session and clears cookies
4. Frontend redirects to `/login`

## Database Selection

Users can select between two databases on login:

1. **neuronip**: Main NeuronIP database
2. **neuronai-demo**: Demo database with sample data

The selected database is stored in the session and used for all subsequent requests. This allows users to switch between production and demo environments.

## Backward Compatibility

The system maintains full backward compatibility with API key authentication:

- API key authentication still works via `Authorization: Bearer <key>` header
- Session middleware checks cookies first, falls back to API key
- Frontend can use either method
- Existing API clients continue to work without changes

## Session Lifecycle

1. **Login**: Create session, generate tokens, set cookies
2. **Request**: Validate access token cookie
3. **Expired Access Token**: Use refresh token to get new access token
4. **Refresh**: Rotate refresh token (old revoked, new issued)
5. **Logout**: Revoke session, clear cookies
6. **Token Reuse**: If old refresh token used, revoke all sessions

## Error Handling

All endpoints return structured error responses:

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid username or password",
    "details": {}
  }
}
```

## Production Considerations

1. **HTTPS Required**: Set `SESSION_COOKIE_SECURE=true` in production
2. **Cookie Domain**: Set appropriate domain for your deployment
3. **Rate Limiting**: Configure rate limits for login attempts
4. **Session Cleanup**: Runs automatically every hour
5. **Monitoring**: Monitor session creation/revocation rates
6. **Logging**: All authentication events are logged

## Testing

### Manual Testing

1. Register a new user:
   ```bash
   curl -X POST http://localhost:8082/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","password":"testpass123","database":"neuronip"}' \
     -c cookies.txt
   ```

2. Login:
   ```bash
   curl -X POST http://localhost:8082/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","password":"testpass123","database":"neuronip"}' \
     -c cookies.txt
   ```

3. Get current user:
   ```bash
   curl -X GET http://localhost:8082/api/v1/auth/me \
     -b cookies.txt
   ```

4. Logout:
   ```bash
   curl -X POST http://localhost:8082/api/v1/auth/logout \
     -b cookies.txt
   ```

## Migration

To enable session-based authentication:

1. Run migration:
   ```bash
   scripts/run-migrations.sh
   ```

2. Rebuild services:
   ```bash
   ./run_neuronip.sh build
   ```

3. Restart services:
   ```bash
   ./run_neuronip.sh run
   ```

4. Test login at: `http://localhost:3001/login`

## Troubleshooting

### Cookies Not Set

- Check CORS configuration
- Verify `withCredentials: true` in frontend
- Check cookie domain settings
- Ensure HTTPS in production

### Session Not Found

- Check database connection
- Verify migration ran successfully
- Check session table exists
- Review server logs

### Refresh Token Invalid

- Check token expiration
- Verify token hasn't been revoked
- Check for token reuse (all sessions revoked)
- Review refresh token table
