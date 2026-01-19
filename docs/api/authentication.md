# 🔒 API Authentication

<div align="center">

**Authentication and Authorization Guide**

[← Endpoints](endpoints.md) • [Rate Limiting →](rate-limiting.md)

</div>

---

## 📋 Table of Contents

- [Overview](#overview)
- [API Key Authentication](#api-key-authentication)
- [JWT Authentication](#jwt-authentication)
- [Creating API Keys](#creating-api-keys)
- [Using API Keys](#using-api-keys)
- [Security Best Practices](#security-best-practices)

---

## 🎯 Overview

NeuronIP supports two authentication methods:

1. **API Key Authentication** - For service-to-service communication
2. **JWT Token Authentication** - For user sessions (web frontend)

---

## 🔑 API Key Authentication

### Header Format

Include your API key in the `Authorization` header:

```
Authorization: Bearer YOUR_API_KEY
```

### Example Request

```bash
curl -X POST http://localhost:8082/api/v1/semantic/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk_live_abc123xyz789" \
  -d '{"query": "test", "limit": 10}'
```

### API Key Format

API keys follow this format:

```
sk_live_<random_string>
sk_test_<random_string>
```

- `sk_live_` - Production keys
- `sk_test_` - Test/development keys

---

## 🎫 JWT Authentication

### Obtaining a JWT Token

1. Authenticate with username/password
2. Receive JWT token in response
3. Include token in subsequent requests

### Using JWT Tokens

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## ➕ Creating API Keys

### Via Dashboard

1. Log in to NeuronIP dashboard
2. Navigate to **Settings** → **API Keys**
3. Click **Create API Key**
4. Enter a name and set permissions
5. Copy the key (shown only once)

### Via API

```bash
curl -X POST http://localhost:8082/api/v1/api-keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_KEY" \
  -d '{
    "name": "My API Key",
    "permissions": ["read", "write"],
    "rate_limit": 1000
  }'
```

---

## 🔐 Security Best Practices

### ✅ Do

- Store API keys securely (environment variables, secrets manager)
- Use different keys for different environments
- Rotate keys regularly
- Use least-privilege permissions
- Monitor key usage

### ❌ Don't

- Commit keys to version control
- Share keys publicly
- Use production keys in development
- Hardcode keys in client-side code
- Reuse keys across multiple services

---

## 🔄 Key Rotation

### Rotating API Keys

1. Create a new API key
2. Update applications to use new key
3. Verify new key works
4. Revoke old key

### Revoking Keys

```bash
curl -X DELETE http://localhost:8082/api/v1/api-keys/{id} \
  -H "Authorization: Bearer YOUR_ADMIN_KEY"
```

---

## 📚 Related Documentation

- [API Overview](overview.md) - API introduction
- [Endpoints](endpoints.md) - API reference
- [Rate Limiting](rate-limiting.md) - Limits and quotas

---

<div align="center">

[← Back to API Docs](README.md)

</div>
