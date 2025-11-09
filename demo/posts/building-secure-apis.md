---
title: Building Secure APIs with Environment Variables
date: 2024-01-29
author: Marcus Rodriguez
excerpt: Best practices for using environment variables to secure your API endpoints and third-party integrations.
tags: [api, security, nodejs]
---

# Building Secure APIs with Environment Variables

When building modern web applications, properly managing API keys and secrets is crucial for security. Here's how we approach it in our blog platform.

## Environment Variables in Practice

Our blog uses several environment variables for different services:

### 1. Analytics Integration

```javascript
const ANALYTICS_KEY_GOOGLE = process.env.ANALYTICS_KEY_GOOGLE;
const ANALYTICS_KEY_MIXPANEL = process.env.ANALYTICS_KEY_MIXPANEL;
```

These keys allow us to:
- Track user behavior and engagement
- Measure content performance
- A/B test different features

### 2. Payment Processing

```javascript
const STRIPE_API_KEY = process.env.STRIPE_API_KEY;
```

This enables:
- Premium subscriptions
- One-time payments
- Subscription management

### 3. Backend Authentication

```javascript
const BACKEND_SECRET = process.env.BACKEND_SECRET_KEY;
```

Used for:
- Admin API endpoints
- Internal service communication
- Webhook verification

## Security Best Practices

### Never Commit Secrets

```bash
# .gitignore
.env
.env.local
.env.*.local
```

### Use Strong, Unique Keys

Generate strong secrets using:
```bash
openssl rand -base64 32
```

### Implement Rate Limiting

Protect your APIs from abuse:
```javascript
app.use('/api/stats', requireAuth, rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 100
}));
```

### Validate Environment Variables on Startup

```javascript
const requiredEnvVars = [
  'ANALYTICS_KEY_GOOGLE',
  'ANALYTICS_KEY_MIXPANEL',
  'STRIPE_API_KEY',
  'BACKEND_SECRET_KEY'
];

requiredEnvVars.forEach(varName => {
  if (!process.env[varName]) {
    console.warn(`Warning: ${varName} is not set`);
  }
});
```

## Team Access Management

Different team members need different levels of access:

- **Engineers**: Full access to all secrets for debugging and development
- **Designers**: Analytics keys for user behavior analysis
- **Marketers**: Analytics keys for campaign tracking

## Using envv for Secret Management

With envv, we can:

1. **Encrypt secrets** using industry-standard encryption
2. **Control access** with fine-grained permissions
3. **Rotate keys** without breaking deployments
4. **Audit usage** for compliance

Example workflow:
```bash
# Engineer adds a new secret
envv set STRIPE_API_KEY sk_live_...

# Designer accesses analytics
envv get ANALYTICS_KEY_GOOGLE

# Marketer views available keys
envv list
```

## Conclusion

Proper environment variable management is essential for application security. By following these practices and using tools like envv, you can build secure applications while maintaining team productivity.