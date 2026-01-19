# ⏱️ Rate Limiting

<div align="center">

**API Rate Limits and Quotas**

[← Authentication](authentication.md)

</div>

---

## 📋 Table of Contents

- [Overview](#overview)
- [Rate Limit Headers](#rate-limit-headers)
- [Tier Limits](#tier-limits)
- [Handling Rate Limits](#handling-rate-limits)
- [Best Practices](#best-practices)

---

## 🎯 Overview

NeuronIP implements rate limiting to ensure fair usage and system stability.

### Rate Limit Headers

Every API response includes rate limit information:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

---

## 📊 Tier Limits

| Tier | Requests/Hour | Requests/Minute | Burst |
|------|---------------|-----------------|-------|
| **Free** | 100 | 10 | 20 |
| **Pro** | 1,000 | 100 | 200 |
| **Enterprise** | Custom | Custom | Custom |

---

## ⚠️ Handling Rate Limits

### Rate Limit Response

When rate limit is exceeded:

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Try again in 60 seconds.",
    "retry_after": 60
  }
}
```

HTTP Status: `429 Too Many Requests`

### Retry Strategy

1. Check `X-RateLimit-Reset` header
2. Wait until reset time
3. Implement exponential backoff
4. Use request queuing for high-volume applications

---

## 💡 Best Practices

- Monitor rate limit headers
- Implement request queuing
- Use caching to reduce API calls
- Batch requests when possible
- Contact support for higher limits

---

<div align="center">

[← Back to API Docs](README.md)

</div>
