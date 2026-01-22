# Authentication System

NeuronIP uses API key-based authentication for accessing the platform. This document describes the authentication flow and how to use it.

## Overview

The authentication system consists of:
- **Login Page** (`/login`) - Where users enter their API key
- **Auth Guard** - Protects dashboard routes and redirects unauthenticated users
- **Token Management** - Settings page allows users to update their API key
- **Automatic Redirects** - Seamless flow between authenticated and unauthenticated states

## User Flow

### First Time Access

1. User visits any dashboard route (e.g., `/dashboard`)
2. `AuthGuard` checks for API token in `localStorage`
3. If no token found, user is redirected to `/login`
4. User enters their API key on the login page
5. API key is validated against the backend
6. If valid, token is stored in `localStorage` and user is redirected to dashboard
7. If invalid, error message is shown

### Subsequent Access

- If token exists in `localStorage`, user can access all dashboard routes
- Token is automatically included in all API requests via axios interceptor
- User can update their token in Settings → Security & Authentication

### Logout

- User clicks logout in the header menu
- Token is cleared from `localStorage`
- User is redirected to `/login`

## Components

### Login Page (`/app/login/page.tsx`)

A clean, centered login form that:
- Accepts API key input (password field)
- Validates the key against the backend before storing
- Shows loading state during validation
- Provides helpful information about creating API keys
- Automatically redirects authenticated users

### Auth Guard (`/components/auth/AuthGuard.tsx`)

A client-side component that:
- Wraps the dashboard layout
- Checks for API token on mount and route changes
- Redirects to `/login` if no token found
- Shows loading state during check
- Skips check for `/login` route

### Security Settings (`/components/settings/SecuritySettings.tsx`)

Enhanced settings component that:
- Shows current API key (masked)
- Allows updating the API key
- Validates new keys before saving
- Provides clear/remove functionality
- Shows validation status

## API Integration

### Token Storage

API tokens are stored in browser `localStorage` with the key `api_token`:

```javascript
localStorage.setItem('api_token', 'your-api-key-here')
localStorage.getItem('api_token')
localStorage.removeItem('api_token')
```

### API Client

The API client (`/lib/api/client.ts`) automatically includes the token in all requests:

```typescript
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('api_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})
```

## Creating API Keys

Users can create API keys using:

1. **API Endpoint**: `POST /api/v1/auth/api-keys`
2. **CLI Script**: `scripts/create-api-key.go`
3. **Database**: Direct insert into `neuronip.api_keys` table

## Security Considerations

- API keys are stored in `localStorage` (client-side only)
- Keys are never sent to third parties
- Keys are masked in the UI (only first 4 and last 4 characters shown)
- Keys are validated against the backend before being accepted
- Invalid keys are rejected with clear error messages

## Testing

### Test the Login Flow

1. Clear your token:
   ```javascript
   localStorage.removeItem('api_token')
   ```

2. Visit any dashboard route:
   ```
   http://localhost:3001/dashboard
   ```

3. You should be redirected to:
   ```
   http://localhost:3001/login
   ```

4. Enter a valid API key and click "Sign In"

5. You should be redirected back to the dashboard

### Test Token Management

1. Go to Settings → Security & Authentication
2. You should see your current API key (masked)
3. Click the eye icon to toggle visibility
4. Enter a new API key and click "Update API Key"
5. The key will be validated and saved if valid

### Test Logout

1. Click the user icon in the header
2. Click "Logout"
3. You should be redirected to `/login`
4. Token should be cleared from `localStorage`

## Troubleshooting

### "Invalid API key" Error

- Verify the API key exists in the database
- Check that the key is not expired or revoked
- Ensure the backend API is running and accessible
- Check browser console for network errors

### Redirect Loop

- Clear `localStorage` completely
- Check that `/login` route is accessible
- Verify `AuthGuard` is not checking `/login` route

### Token Not Persisting

- Check browser's `localStorage` settings
- Ensure cookies/localStorage are enabled
- Try in incognito/private mode to rule out extensions

## Future Enhancements

Potential improvements:
- Session expiration and refresh tokens
- Remember me functionality
- OAuth/SSO integration
- Multi-factor authentication
- Token rotation policies
- API key scopes and permissions
