# Error Handling Guide

This guide covers error handling patterns, best practices, and error code reference for NeuronIP.

## Overview

NeuronIP uses a structured error handling system that ensures:
- Consistent error format across all endpoints
- Request ID tracking for debugging
- User-friendly error messages
- Proper HTTP status codes

## Error Structure

All API errors follow this structure:

```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid input provided",
    "details": {
      "field": "email",
      "reason": "Invalid email format",
      "request_id": "550e8400-e29b-41d4-a716-446655440000"
    }
  }
}
```

## Error Codes

### Client Errors (4xx)

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `BAD_REQUEST` | 400 | Invalid request format or parameters |
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_FAILED` | 400 | Request validation failed |
| `TOO_MANY_REQUESTS` | 429 | Rate limit exceeded |
| `TIMEOUT` | 504 | Request timeout |

### Server Errors (5xx)

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INTERNAL_SERVER_ERROR` | 500 | Unexpected server error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |

## Creating Errors

### Using Error Constructors

```go
import "github.com/neurondb/NeuronIP/api/internal/errors"

// Bad request
err := errors.BadRequest("Invalid input")

// With details
err := errors.ValidationFailed("Validation failed", map[string]string{
    "email": "Invalid email format",
    "password": "Password too short",
})

// Unauthorized
err := errors.Unauthorized("Invalid API key")

// Not found
err := errors.NotFound("User")

// Internal server error
err := errors.InternalServer("Database connection failed")
```

### Wrapping Errors

```go
// Wrap existing errors
err := errors.Wrap(originalErr, errors.ErrCodeInternalServer, "Failed to process request")

// Check if error is APIError
if apiErr := errors.AsAPIError(err); apiErr != nil {
    // Handle API error
}
```

## Writing Error Responses

### In Handlers

Always use context-aware error handlers:

```go
import "github.com/neurondb/NeuronIP/api/internal/handlers"

func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    // Process request...
    
    if err != nil {
        // Use WriteErrorWithRequest for automatic context handling
        handlers.WriteErrorWithRequest(w, r, err)
        return
    }
    
    // Or use WriteErrorResponseWithContext for APIError
    handlers.WriteErrorResponseWithContext(w, r, errors.BadRequest("Invalid input"))
}
```

### Error Response Functions

- `WriteError(w, err)` - Basic error writing (no context)
- `WriteErrorWithRequest(w, r, err)` - Includes request context
- `WriteErrorResponse(w, apiErr)` - Direct APIError (no context)
- `WriteErrorResponseWithContext(w, r, apiErr)` - Includes request context

**Always use the context-aware versions** to ensure request IDs are included.

## Error Handling Patterns

### Pattern 1: Service Layer Errors

```go
// Service layer
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.GetUser(ctx, id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.NotFound("User")
        }
        return nil, errors.Wrap(err, errors.ErrCodeInternalServer, "Failed to get user")
    }
    return user, nil
}

// Handler
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    user, err := h.service.GetUser(r.Context(), id)
    if err != nil {
        handlers.WriteErrorWithRequest(w, r, err)
        return
    }
    json.NewEncoder(w).Encode(user)
}
```

### Pattern 2: Validation Errors

```go
// Using validation middleware
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

router.Handle("/users", middleware.ValidateJSON(&CreateUserRequest{})(
    http.HandlerFunc(handler.CreateUser),
))

// Manual validation
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        handlers.WriteErrorResponseWithContext(w, r, errors.BadRequest("Invalid JSON"))
        return
    }
    
    if req.Email == "" {
        handlers.WriteErrorResponseWithContext(w, r, errors.ValidationFailed(
            "Validation failed",
            map[string]string{"email": "Email is required"},
        ))
        return
    }
}
```

### Pattern 3: Database Errors

```go
func (s *Service) CreateUser(ctx context.Context, user *User) error {
    err := s.repo.CreateUser(ctx, user)
    if err != nil {
        // Check for specific database errors
        if pgErr, ok := err.(*pgconn.PgError); ok {
            switch pgErr.Code {
            case "23505": // Unique violation
                return errors.BadRequest("User already exists")
            case "23503": // Foreign key violation
                return errors.BadRequest("Invalid reference")
            }
        }
        return errors.Wrap(err, errors.ErrCodeInternalServer, "Failed to create user")
    }
    return nil
}
```

## Frontend Error Handling

### Error Parsing

The frontend automatically parses backend errors:

```typescript
import { formatError, getErrorMessage } from '@/lib/api/errorHandler'

try {
  await apiClient.post('/endpoint', data)
} catch (error) {
  const formatted = formatError(error)
  console.error(formatted.title, formatted.message)
  // Display formatted.error to user
}
```

### Error Display

```typescript
import { formatError } from '@/lib/api/errorHandler'

const error = formatError(apiError)
// error.title: "Error Title"
// error.message: "User-friendly message"
// error.recovery: Array of recovery actions
// error.requestId: Request ID for support
```

## Best Practices

1. **Always include request context**:
   ```go
   handlers.WriteErrorWithRequest(w, r, err)
   ```

2. **Use appropriate error codes**:
   - `BAD_REQUEST` for invalid input
   - `UNAUTHORIZED` for auth issues
   - `NOT_FOUND` for missing resources
   - `INTERNAL_SERVER_ERROR` for unexpected errors

3. **Provide helpful error messages**:
   ```go
   // Bad
   errors.BadRequest("Error")
   
   // Good
   errors.BadRequest("Email address is required")
   ```

4. **Include validation details**:
   ```go
   errors.ValidationFailed("Validation failed", map[string]string{
       "email": "Invalid email format",
       "age": "Must be 18 or older",
   })
   ```

5. **Log errors with context**:
   ```go
   logging.ErrorContext(ctx, "Failed to process request", err)
   ```

6. **Don't expose internal details**:
   ```go
   // Bad - exposes internal structure
   errors.InternalServer(err.Error())
   
   // Good - generic message
   errors.InternalServer("An unexpected error occurred")
   ```

7. **Wrap errors for context**:
   ```go
   return errors.Wrap(err, errors.ErrCodeInternalServer, "Failed to save user")
   ```

## Error Code Reference

### Complete Error Code List

```go
const (
    ErrCodeBadRequest       ErrorCode = "BAD_REQUEST"
    ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
    ErrCodeForbidden        ErrorCode = "FORBIDDEN"
    ErrCodeNotFound         ErrorCode = "NOT_FOUND"
    ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
    ErrCodeTooManyRequests  ErrorCode = "TOO_MANY_REQUESTS"
    ErrCodeTimeout          ErrorCode = "TIMEOUT"
    ErrCodeInternalServer   ErrorCode = "INTERNAL_SERVER_ERROR"
    ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)
```

## Debugging Errors

1. **Check request ID** in error response
2. **Search logs** using request ID:
   ```bash
   grep <request-id> logs/app.log
   ```
3. **Review error details** in response
4. **Check error code** for category
5. **Verify error handling** in handler code

## Related Documentation

- [Request Flow Guide](../architecture/request-flow.md)
- [Validation Guide](./validation.md)
- [Backend Architecture](../architecture/backend.md)
