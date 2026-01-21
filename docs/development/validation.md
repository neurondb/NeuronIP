# Validation Guide

This guide covers validation patterns, how to add validators, and frontend validation integration in NeuronIP.

## Overview

NeuronIP uses a two-layer validation approach:
1. **Backend validation** using `go-playground/validator` with struct tags
2. **Frontend validation** for immediate user feedback

## Backend Validation

### Using Validation Middleware

The `ValidateJSON` middleware automatically validates request bodies:

```go
import "github.com/neurondb/NeuronIP/api/internal/middleware"

type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Age      int    `json:"age" validate:"required,min=18,max=120"`
}

router.Handle("/users", middleware.ValidateJSON(&CreateUserRequest{})(
    http.HandlerFunc(handler.CreateUser),
)).Methods("POST")
```

### Validation Tags

Common validation tags:

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Field is required | `validate:"required"` |
| `email` | Valid email address | `validate:"email"` |
| `url` | Valid URL | `validate:"url"` |
| `uuid` | Valid UUID | `validate:"uuid"` |
| `min=n` | Minimum length/value | `validate:"min=8"` |
| `max=n` | Maximum length/value | `validate:"max=100"` |
| `len=n` | Exact length/value | `validate:"len=10"` |
| `numeric` | Numeric value | `validate:"numeric"` |
| `alpha` | Letters only | `validate:"alpha"` |
| `alphanum` | Letters and numbers | `validate:"alphanum"` |
| `gte=n` | Greater than or equal | `validate:"gte=0"` |
| `lte=n` | Less than or equal | `validate:"lte=100"` |
| `gt=n` | Greater than | `validate:"gt=0"` |
| `lt=n` | Less than | `validate:"lt=100"` |
| `oneof=val1 val2` | One of specified values | `validate:"oneof=admin user guest"` |
| `eqfield=Field` | Must equal another field | `validate:"eqfield=Password"` |
| `nefield=Field` | Must not equal another field | `validate:"nefield=Password"` |

### Custom Validators

To add custom validators:

```go
import "github.com/go-playground/validator/v10"

func init() {
    validate.RegisterValidation("customtag", func(fl validator.FieldLevel) bool {
        // Custom validation logic
        return fl.Field().String() == "valid"
    })
}
```

### Validation Error Messages

Error messages are automatically generated. To customize:

```go
// The middleware calls getValidationErrorMessage() which you can extend
func getValidationErrorMessage(field, tag, param string) string {
    switch tag {
    case "customtag":
        return "Custom validation failed"
    // ... other cases
    }
}
```

### Manual Validation

For complex validation, validate in the handler:

```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        handlers.WriteErrorResponseWithContext(w, r, errors.BadRequest("Invalid JSON"))
        return
    }
    
    // Custom validation
    if !isValidDomain(req.Email) {
        handlers.WriteErrorResponseWithContext(w, r, errors.ValidationFailed(
            "Validation failed",
            map[string]string{"email": "Email domain not allowed"},
        ))
        return
    }
    
    // Process request...
}
```

## Frontend Validation

### Form Validation

Use HTML5 validation attributes:

```typescript
<input
  type="email"
  required
  minLength={8}
  pattern="[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$"
/>
```

### React Hook Form

For complex forms, use React Hook Form:

```typescript
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const schema = z.object({
  email: z.string().email("Invalid email address"),
  password: z.string().min(8, "Password must be at least 8 characters"),
})

function MyForm() {
  const { register, handleSubmit, formState: { errors } } = useForm({
    resolver: zodResolver(schema),
  })
  
  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register("email")} />
      {errors.email && <span>{errors.email.message}</span>}
      
      <input {...register("password")} type="password" />
      {errors.password && <span>{errors.password.message}</span>}
    </form>
  )
}
```

### Validation in Wizards

Wizards use step-level validation:

```typescript
const steps: WizardStep[] = [
  {
    id: 'name',
    title: 'Name',
    component: NameStep,
    validate: (data) => {
      return !!(data.name && data.name.trim().length > 0)
    },
  },
]
```

### Real-time Validation

Provide immediate feedback:

```typescript
function EmailInput({ value, onChange }) {
  const [error, setError] = useState('')
  
  const handleChange = (e) => {
    const email = e.target.value
    onChange(email)
    
    if (email && !isValidEmail(email)) {
      setError('Invalid email address')
    } else {
      setError('')
    }
  }
  
  return (
    <div>
      <input value={value} onChange={handleChange} />
      {error && <span className="error">{error}</span>}
    </div>
  )
}
```

## Validation Patterns

### Pattern 1: Required Fields

**Backend:**
```go
type Request struct {
    Name string `json:"name" validate:"required"`
}
```

**Frontend:**
```typescript
<input required />
// or
{!name && <span>Name is required</span>}
```

### Pattern 2: Email Validation

**Backend:**
```go
type Request struct {
    Email string `json:"email" validate:"required,email"`
}
```

**Frontend:**
```typescript
<input type="email" required />
// or custom validation
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
if (!emailRegex.test(email)) {
  setError('Invalid email format')
}
```

### Pattern 3: Password Validation

**Backend:**
```go
type Request struct {
    Password        string `json:"password" validate:"required,min=8"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}
```

**Frontend:**
```typescript
const [password, setPassword] = useState('')
const [confirm, setConfirm] = useState('')
const [error, setError] = useState('')

useEffect(() => {
  if (confirm && password !== confirm) {
    setError('Passwords do not match')
  } else {
    setError('')
  }
}, [password, confirm])
```

### Pattern 4: Numeric Ranges

**Backend:**
```go
type Request struct {
    Age  int `json:"age" validate:"required,min=18,max=120"`
    Score int `json:"score" validate:"gte=0,lte=100"`
}
```

**Frontend:**
```typescript
<input
  type="number"
  min={18}
  max={120}
  value={age}
  onChange={(e) => {
    const val = parseInt(e.target.value)
    if (val < 18 || val > 120) {
      setError('Age must be between 18 and 120')
    }
  }}
/>
```

### Pattern 5: Conditional Validation

**Backend:**
```go
// Validate in handler
if req.Type == "premium" && req.PaymentMethod == "" {
    return errors.ValidationFailed("Validation failed", map[string]string{
        "payment_method": "Payment method required for premium",
    })
}
```

**Frontend:**
```typescript
{type === 'premium' && (
  <div>
    <input
      required={type === 'premium'}
      value={paymentMethod}
      onChange={(e) => setPaymentMethod(e.target.value)}
    />
  </div>
)}
```

## Best Practices

1. **Validate on both frontend and backend**:
   - Frontend: Immediate user feedback
   - Backend: Security and data integrity

2. **Use appropriate validation tags**:
   ```go
   // Good
   Email string `validate:"required,email"`
   
   // Avoid manual checks when tags exist
   ```

3. **Provide clear error messages**:
   ```go
   // Good
   "email": "Please enter a valid email address"
   
   // Bad
   "email": "Invalid"
   ```

4. **Show validation errors immediately**:
   - Validate on blur for better UX
   - Show errors inline with fields

5. **Don't duplicate validation logic**:
   - Use shared validation schemas (Zod, etc.)
   - Keep backend as source of truth

6. **Handle async validation**:
   ```typescript
   const checkEmailExists = async (email: string) => {
     const exists = await apiClient.get(`/users/check-email?email=${email}`)
     if (exists) {
       setError('Email already taken')
     }
   }
   ```

7. **Validate file uploads**:
   ```go
   type Request struct {
       FileSize int64  `validate:"max=10485760"` // 10MB
       FileType string `validate:"oneof=image/jpeg image/png"`
   }
   ```

## Common Validation Scenarios

### UUID Validation

```go
type Request struct {
    ID string `json:"id" validate:"required,uuid"`
}
```

### URL Validation

```go
type Request struct {
    WebhookURL string `json:"webhook_url" validate:"required,url"`
}
```

### Enum Validation

```go
type Request struct {
    Status string `json:"status" validate:"required,oneof=pending active completed"`
}
```

### Date Validation

```go
type Request struct {
    StartDate time.Time `json:"start_date" validate:"required"`
    EndDate   time.Time `json:"end_date" validate:"required,gtefield=StartDate"`
}
```

## Testing Validation

### Backend Testing

```go
func TestCreateUserValidation(t *testing.T) {
    tests := []struct {
        name    string
        request CreateUserRequest
        wantErr bool
    }{
        {
            name: "valid request",
            request: CreateUserRequest{
                Email:    "test@example.com",
                Password: "password123",
            },
            wantErr: false,
        },
        {
            name: "missing email",
            request: CreateUserRequest{
                Password: "password123",
            },
            wantErr: true,
        },
        // ... more tests
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validate.Struct(tt.request)
            if (err != nil) != tt.wantErr {
                t.Errorf("validation error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Frontend Testing

```typescript
describe('Email validation', () => {
  it('shows error for invalid email', () => {
    render(<EmailInput />)
    const input = screen.getByRole('textbox')
    
    fireEvent.change(input, { target: { value: 'invalid-email' } })
    fireEvent.blur(input)
    
    expect(screen.getByText('Invalid email address')).toBeInTheDocument()
  })
})
```

## Related Documentation

- [Error Handling Guide](./error-handling.md)
- [Request Flow Guide](../architecture/request-flow.md)
- [Wizard Development Guide](./wizards.md)
