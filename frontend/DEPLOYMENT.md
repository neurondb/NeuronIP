# Deployment Guide

Complete guide for deploying the NeuronIP frontend to production.

## Prerequisites

- Node.js 18+ installed
- Environment variables configured
- Build tools available

## Build Process

### 1. Install Dependencies

```bash
npm ci
```

### 2. Environment Setup

Create `.env.production` with production values:

```bash
NEXT_PUBLIC_API_URL=https://api.neuronip.com
NEXT_PUBLIC_WS_URL=wss://api.neuronip.com/ws
NEXT_PUBLIC_ENABLE_ANALYTICS=true
NEXT_PUBLIC_ENABLE_SENTRY=true
NEXT_PUBLIC_SENTRY_DSN=your-sentry-dsn
NEXT_PUBLIC_SENTRY_ENVIRONMENT=production
```

### 3. Build

```bash
npm run build
```

### 4. Start Production Server

```bash
npm start
```

## Docker Deployment

### Dockerfile

The project includes a Dockerfile. Build and run:

```bash
docker build -t neuronip-frontend .
docker run -p 3000:3000 neuronip-frontend
```

### Docker Compose

```yaml
version: '3.8'
services:
  frontend:
    build: .
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=${API_URL}
    restart: unless-stopped
```

## Vercel Deployment

### Automatic Deployment

1. Connect your repository to Vercel
2. Configure environment variables in Vercel dashboard
3. Deploy automatically on push to main branch

### Manual Deployment

```bash
npm install -g vercel
vercel --prod
```

## Environment Variables

Required for production:

- `NEXT_PUBLIC_API_URL` - Backend API URL
- `NEXT_PUBLIC_WS_URL` - WebSocket URL (optional)
- `NEXT_PUBLIC_ENABLE_SENTRY` - Enable error tracking
- `NEXT_PUBLIC_SENTRY_DSN` - Sentry DSN (if enabled)

## Performance Optimization

### 1. Enable Compression

Already configured in `next.config.js`.

### 2. CDN Configuration

Configure CDN for static assets:
- Images
- Fonts
- Static files

### 3. Caching Headers

Set appropriate cache headers:
- Static assets: 1 year
- HTML: No cache
- API responses: As needed

## Monitoring

### Error Tracking

Sentry is integrated. Configure in environment variables.

### Analytics

Enable analytics by setting `NEXT_PUBLIC_ENABLE_ANALYTICS=true`.

## Security

### 1. Content Security Policy

Add CSP headers in `next.config.js`:

```js
headers: async () => [
  {
    source: '/:path*',
    headers: [
      {
        key: 'Content-Security-Policy',
        value: "default-src 'self'; script-src 'self' 'unsafe-eval' 'unsafe-inline';"
      }
    ]
  }
]
```

### 2. Environment Variables

Never commit `.env` files. Use secure secret management.

## Troubleshooting

### Build Failures

1. Check Node.js version (18+)
2. Clear `.next` directory
3. Delete `node_modules` and reinstall
4. Check for TypeScript errors

### Runtime Errors

1. Check environment variables
2. Verify API connectivity
3. Check browser console
4. Review Sentry logs

## CI/CD

### GitHub Actions Example

```yaml
name: Deploy
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - run: npm ci
      - run: npm run build
      - run: npm run test
      - name: Deploy
        run: vercel --prod --token ${{ secrets.VERCEL_TOKEN }}
```

## Rollback

If deployment fails:

1. Revert to previous version
2. Check error logs
3. Fix issues
4. Redeploy

## Health Checks

Configure health check endpoint:

```tsx
// app/api/health/route.ts
export async function GET() {
  return Response.json({ status: 'ok' })
}
```

## Scaling

### Horizontal Scaling

- Use load balancer
- Multiple instances
- Session management

### Vertical Scaling

- Increase server resources
- Optimize bundle size
- Enable caching
