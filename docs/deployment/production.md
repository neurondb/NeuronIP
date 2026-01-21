# 🚀 Production Deployment

<div align="center">

**Production deployment guide for NeuronIP**

[← Docker](docker.md) • [Packaging](packaging.md) • [Kubernetes](kubernetes.md) • [Monitoring →](monitoring.md)

</div>

---

## 📋 Table of Contents

- [Architecture Overview](#architecture-overview)
- [Prerequisites](#prerequisites)
- [Deployment Options](#deployment-options)
- [External Service Configuration](#external-service-configuration)
- [CI/CD and Release Process](#cicd-and-release-process)
- [Security Checklist](#security-checklist)
- [Performance Optimization](#performance-optimization)
- [Monitoring and Observability](#monitoring-and-observability)

---

## Architecture Overview

NeuronIP uses a **containerized microservices architecture** with the following components:

### Core Services (Packaged with NeuronIP)

- **Backend API**: Go application (port 8082)
  - Statically compiled binary in Alpine Linux container
  - Multi-stage Docker build for minimal image size (~10-20MB)
  - Handles all business logic and API endpoints

- **Frontend**: Next.js application (port 3000)
  - Production-optimized Next.js build
  - Served via Node.js Alpine container
  - Communicates only with the API service

### External Dependencies (Deployed Separately)

> **Important**: The following services are **not** packaged with NeuronIP and must be deployed separately on different machines or containers:

- **PostgreSQL**: Primary database for NeuronIP application data
- **NeuronDB**: Additional database service for specialized data operations
- **NeuronMCP**: MCP (Model Context Protocol) service
- **NeuronAgent**: Agent service for AI workflows

**Connection Pattern:**
```
Frontend → API → PostgreSQL
              → NeuronDB
              → NeuronMCP
              → NeuronAgent
```

The API service connects to all external services via configured endpoints. The frontend only communicates with the API service.

---

## Prerequisites

### Infrastructure Requirements

- **Container Runtime**: Docker 20.10+ or Kubernetes 1.24+
- **Network Access**: Connectivity to external services (PostgreSQL, NeuronDB, NeuronMCP, NeuronAgent)
- **SSL/TLS Certificates**: For production HTTPS endpoints
- **Secrets Management**: Secure storage for database credentials and API keys

### External Services

Before deploying NeuronIP, ensure the following external services are available:

1. **PostgreSQL 16+**
   - Database: `neuronip`
   - User credentials configured
   - Network accessible from NeuronIP API containers

2. **NeuronDB**
   - Service endpoint accessible
   - Database configured
   - Network connectivity verified

3. **NeuronMCP**
   - MCP service running
   - Endpoint URL configured
   - Authentication configured (if required)

4. **NeuronAgent**
   - Agent service running
   - API endpoint accessible
   - API key configured (if required)

### Environment Configuration

- Environment variables configured for all services
- Secrets stored securely (Kubernetes secrets, Docker secrets, or external vault)
- Network policies configured (if using Kubernetes)
- Monitoring and logging infrastructure ready

---

## Deployment Options

NeuronIP supports multiple deployment strategies for production:

### Option 1: Docker Compose

**Best for:** Single-server deployments, development, testing, small production environments

#### Quick Start

```bash
# Build and start services
docker compose up -d --build

# Verify services
docker compose ps
curl http://localhost:8082/health
```

#### Production Configuration

**Create `docker-compose.prod.yml`:**
```yaml
version: '3.8'

services:
  neuronip-api:
    image: neuronip/api:v1.2.3  # Use specific version
    restart: always
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8082/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    environment:
      DB_HOST: ${DB_HOST}  # External PostgreSQL
      DB_PORT: ${DB_PORT:-5432}
      DB_USER: ${DB_USER}
      DB_PASSWORD: ${DB_PASSWORD}  # From .env or secrets
      DB_NAME: ${DB_NAME:-neuronip}
      NEURONDB_HOST: ${NEURONDB_HOST}
      NEURONDB_PORT: ${NEURONDB_PORT:-5432}
      NEURONAGENT_ENDPOINT: ${NEURONAGENT_ENDPOINT}
    networks:
      - neuronip-network

  neuronip-frontend:
    image: neuronip/frontend:v1.2.3
    restart: always
    deploy:
      resources:
        limits:
          cpus: '0.2'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:3000/"]
      interval: 30s
      timeout: 10s
      retries: 3
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    environment:
      NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8082/api/v1}
    depends_on:
      neuronip-api:
        condition: service_healthy
    networks:
      - neuronip-network

networks:
  neuronip-network:
    external: true  # Or create: docker network create neuronip-network
```

**Deploy with production config:**
```bash
# Use production override
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Or use environment-specific file
docker compose -f docker-compose.prod.yml up -d
```

#### Environment Variables

**Create `.env.production`:**
```bash
# Database Configuration
DB_HOST=postgres.production.example.com
DB_PORT=5432
DB_USER=neuronip
DB_PASSWORD=secure_password_here
DB_NAME=neuronip

# External Services
NEURONDB_HOST=neurondb.production.example.com
NEURONDB_PORT=5432
NEURONDB_DATABASE=neurondb
NEURONAGENT_ENDPOINT=https://neuronagent.production.example.com:8080

# Frontend Configuration
NEXT_PUBLIC_API_URL=https://api.neuronip.example.com/api/v1

# Security
NEURONAGENT_API_KEY=your_api_key_here
```

**Load environment:**
```bash
# Load from file
docker compose --env-file .env.production up -d

# Or export before running
export $(cat .env.production | xargs)
docker compose up -d
```

#### Using Docker Secrets

**Create secrets:**
```bash
# Create secret files
echo "secure_password" | docker secret create db_password -
echo "api_key_here" | docker secret create neuronagent_api_key -
```

**Update docker-compose.yml:**
```yaml
services:
  neuronip-api:
    secrets:
      - db_password
      - neuronagent_api_key
    environment:
      DB_PASSWORD_FILE: /run/secrets/db_password
      NEURONAGENT_API_KEY_FILE: /run/secrets/neuronagent_api_key

secrets:
  db_password:
    external: true
  neuronagent_api_key:
    external: true
```

#### Resource Limits

**Configure in docker-compose.yml:**
```yaml
services:
  neuronip-api:
    deploy:
      resources:
        limits:
          cpus: '0.5'      # 50% of one CPU
          memory: 512M     # 512 MB RAM
        reservations:
          cpus: '0.25'     # Guaranteed 25% CPU
          memory: 256M     # Guaranteed 256 MB RAM
```

#### Health Checks

**Configure health checks:**
```yaml
services:
  neuronip-api:
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8082/health"]
      interval: 30s        # Check every 30 seconds
      timeout: 10s         # Timeout after 10 seconds
      retries: 3           # Retry 3 times before marking unhealthy
      start_period: 40s    # Allow 40 seconds for startup
```

**Limitations:**
- Limited scalability (single server)
- No high availability (single point of failure)
- Manual scaling required
- No automatic load balancing

**Production Considerations:**
- ✅ Use Docker secrets for sensitive data
- ✅ Configure resource limits
- ✅ Set up health checks
- ✅ Enable logging aggregation
- ✅ Use specific image versions (not `latest`)
- ✅ Configure restart policies
- ✅ Set up monitoring
- ✅ Regular backups

### Option 2: Kubernetes (Direct Manifests)

**Best for:** Production Kubernetes clusters with full control

```bash
# Apply all manifests
kubectl apply -f k8s/

# Check deployment status
kubectl get deployments -n neuronip
kubectl get pods -n neuronip
```

**Features:**
- API: 3 replicas with resource limits
- Frontend: 2 replicas with resource limits
- Health probes configured
- Secrets management via Kubernetes secrets
- Horizontal Pod Autoscaler (HPA) enabled

**Configuration:**
- Update `k8s/secrets.yaml` with database credentials
- Configure external service endpoints in `k8s/deployment.yaml`
- Set up Ingress for external access

See [Kubernetes Deployment Guide](kubernetes.md) for detailed instructions.

### Option 3: Helm Chart

**Best for:** Production Kubernetes with multiple environments, CI/CD pipelines, enterprise deployments

#### Installation

**Basic installation:**
```bash
# Install with default values
helm install neuronip ./helm/neuronip \
  --namespace neuronip \
  --create-namespace
```

**Installation with custom values:**
```bash
# Install with overrides
helm install neuronip ./helm/neuronip \
  --namespace neuronip \
  --create-namespace \
  --set database.host=postgres.production.example.com \
  --set database.user=neuronip \
  --set image.api.tag=v1.2.3 \
  --set image.frontend.tag=v1.2.3 \
  --set replicaCount=5 \
  --set autoscaling.enabled=true \
  --set autoscaling.minReplicas=3 \
  --set autoscaling.maxReplicas=10
```

**Installation with values file:**
```bash
# Create production values file
cat > values-production.yaml <<EOF
replicaCount: 5

image:
  api:
    repository: neuronip/api
    tag: v1.2.3
    pullPolicy: IfNotPresent
  frontend:
    repository: neuronip/frontend
    tag: v1.2.3
    pullPolicy: IfNotPresent

database:
  host: postgres.production.example.com
  port: 5432
  name: neuronip
  user: neuronip

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

resources:
  api:
    requests:
      memory: "512Mi"
      cpu: "500m"
    limits:
      memory: "1Gi"
      cpu: "1000m"
EOF

# Install with values file
helm install neuronip ./helm/neuronip \
  --namespace neuronip \
  --create-namespace \
  -f values-production.yaml
```

#### Upgrades

**Upgrade to new version:**
```bash
# Upgrade with new image version
helm upgrade neuronip ./helm/neuronip \
  --namespace neuronip \
  --set image.api.tag=v1.3.0 \
  --set image.frontend.tag=v1.3.0

# Upgrade with values file
helm upgrade neuronip ./helm/neuronip \
  --namespace neuronip \
  -f values-production.yaml
```

**Rollback:**
```bash
# List release history
helm history neuronip -n neuronip

# Rollback to previous version
helm rollback neuronip -n neuronip

# Rollback to specific revision
helm rollback neuronip 2 -n neuronip
```

#### Configuration Examples

**Development environment:**
```yaml
# values-dev.yaml
replicaCount: 1

image:
  api:
    tag: latest
  frontend:
    tag: latest

autoscaling:
  enabled: false

resources:
  api:
    requests:
      memory: "128Mi"
      cpu: "100m"
    limits:
      memory: "256Mi"
      cpu: "200m"
```

**Staging environment:**
```yaml
# values-staging.yaml
replicaCount: 2

image:
  api:
    tag: v1.2.3
  frontend:
    tag: v1.2.3

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5

ingress:
  enabled: true
  hosts:
    - host: api-staging.neuronip.example.com
    - host: app-staging.neuronip.example.com
```

**Production environment:**
```yaml
# values-production.yaml
replicaCount: 3

image:
  api:
    repository: neuronip/api
    tag: v1.2.3
    pullPolicy: Always  # Always pull for production
  frontend:
    repository: neuronip/frontend
    tag: v1.2.3
    pullPolicy: Always

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

resources:
  api:
    requests:
      memory: "512Mi"
      cpu: "500m"
    limits:
      memory: "2Gi"
      cpu: "2000m"
  frontend:
    requests:
      memory: "256Mi"
      cpu: "200m"
    limits:
      memory: "512Mi"
      cpu: "500m"

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
  hosts:
    - host: api.neuronip.example.com
    - host: app.neuronip.example.com
  tls:
    - secretName: neuronip-tls
      hosts:
        - api.neuronip.example.com
        - app.neuronip.example.com
```

#### Secrets Management

**Create secrets:**
```bash
# Create secret for database credentials
kubectl create secret generic neuronip-secrets \
  --from-literal=db-host=postgres.production.example.com \
  --from-literal=db-user=neuronip \
  --from-literal=db-password=secure_password \
  --namespace neuronip

# Create secret from file
kubectl create secret generic neuronip-secrets \
  --from-file=db-password=./secrets/db-password.txt \
  --namespace neuronip
```

**Use with Helm:**
```yaml
# In values.yaml or via --set
database:
  host: postgres.production.example.com
  # Password from secret
  passwordSecret: neuronip-secrets
  passwordKey: db-password
```

#### Benefits

- ✅ Environment-specific configurations via `values.yaml`
- ✅ Version management (Helm releases)
- ✅ Easy upgrades and rollbacks
- ✅ Templated deployments (DRY principle)
- ✅ Dependency management (if using Helm charts)
- ✅ Configuration validation
- ✅ Release history tracking

#### Advanced Features

**Dry-run (preview changes):**
```bash
helm install neuronip ./helm/neuronip \
  --namespace neuronip \
  --dry-run \
  --debug
```

**Template rendering (debug):**
```bash
# See rendered templates
helm template neuronip ./helm/neuronip \
  --namespace neuronip \
  -f values-production.yaml

# Validate templates
helm lint ./helm/neuronip
```

**Dependency management:**
```yaml
# Chart.yaml
dependencies:
  - name: postgresql
    version: 12.0.0
    repository: https://charts.bitnami.com/bitnami
    condition: postgresql.enabled
```

**Install dependencies:**
```bash
helm dependency update ./helm/neuronip
helm install neuronip ./helm/neuronip --namespace neuronip
```

---

## External Service Configuration

### Database Connection (PostgreSQL)

**Environment Variables:**
```bash
DB_HOST=postgres.example.com          # External PostgreSQL host
DB_PORT=5432                          # PostgreSQL port
DB_USER=neuronip                      # Database user
DB_PASSWORD=secure_password           # From secrets
DB_NAME=neuronip                      # Database name
```

**Kubernetes Secrets:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: neuronip-secrets
  namespace: neuronip
type: Opaque
stringData:
  db-host: postgres.example.com
  db-user: neuronip
  db-password: secure_password
```

### NeuronDB Configuration

**Environment Variables:**
```bash
NEURONDB_HOST=neurondb.example.com    # NeuronDB host
NEURONDB_PORT=5432                    # NeuronDB port
NEURONDB_DATABASE=neurondb            # NeuronDB database
```

### NeuronAgent Configuration

**Environment Variables:**
```bash
NEURONAGENT_ENDPOINT=https://neuronagent.example.com:8080
NEURONAGENT_API_KEY=your_api_key      # From secrets
```

### Network Connectivity

Ensure network connectivity from NeuronIP API containers to external services:

**Docker Compose:**
- Services must be on the same network or accessible via hostname/IP
- Use external network: `neurondb-network`

**Kubernetes:**
- External services accessible via:
  - External IP addresses
  - Kubernetes ExternalName services
  - Ingress with proper routing
- Network policies configured to allow outbound connections

**Testing Connectivity:**
```bash
# From API container
docker compose exec neuronip-api nc -zv postgres.example.com 5432
docker compose exec neuronip-api nc -zv neurondb.example.com 5432

# From Kubernetes pod
kubectl exec -n neuronip deployment/neuronip-api -- nc -zv postgres.example.com 5432
```

---

## CI/CD and Release Process

### Image Registry Strategy

NeuronIP Docker images are published to **Docker Hub**:

- **API Image**: `neuronip/api`
- **Frontend Image**: `neuronip/frontend`

**Tagging Strategy:**
- `latest`: Most recent release
- `v1.2.3`: Semantic version tags
- `main-abc123`: Branch-commit SHA tags
- `pr-123`: Pull request tags

### Automated Builds (GitHub Actions)

**Build Workflow** (`.github/workflows/build.yml`):
- Triggers: Push to `main` branch or version tags (`v*`)
- Actions:
  - Builds Docker images with Docker Buildx
  - Pushes to Docker Hub
  - Uses GitHub Actions cache for faster builds
  - Tags images with multiple strategies

**Release Workflow** (`.github/workflows/release.yml`):
- Triggers: Version tags (`v*.*.*`)
- Actions:
  1. Extracts version from git tag
  2. Generates changelog
  3. Builds Go binary with version metadata
  4. Builds Next.js production bundle
  5. Builds and pushes Docker images with version tags
  6. Creates GitHub Release
  7. Uploads binary artifacts

### Release Process

**Manual Release:**
```bash
# Run release script
./scripts/release.sh patch  # or minor, major

# Script will:
# 1. Calculate new version
# 2. Create git tag
# 3. Push tag (triggers release workflow)
```

**Version Metadata:**
- Go binary includes: version, build date, git commit
- Docker images tagged with semantic version
- GitHub Release created automatically

### Deployment from Registry

**Pull and Deploy:**
```bash
# Docker Compose
docker compose pull
docker compose up -d

# Kubernetes
kubectl set image deployment/neuronip-api \
  api=neuronip/api:v1.2.3 \
  -n neuronip

kubectl set image deployment/neuronip-frontend \
  frontend=neuronip/frontend:v1.2.3 \
  -n neuronip
```

**Helm Upgrade:**
```bash
helm upgrade neuronip ./helm/neuronip \
  --set image.api.tag=v1.2.3 \
  --set image.frontend.tag=v1.2.3
```

---

## Security Checklist

### Container Security

- ✅ Multi-stage builds (minimal attack surface)
- ✅ Non-root containers (Alpine base images)
- ✅ Minimal base images (Alpine Linux)
- ✅ No build tools in runtime images
- ✅ Static binary compilation (Go backend)

### Secrets Management

- ✅ Credentials stored in Kubernetes secrets
- ✅ Environment variables for non-sensitive config
- ✅ No secrets in Docker images
- ✅ Secrets rotation capability

### Network Security

- ✅ TLS/SSL via cert-manager (Kubernetes)
- ✅ Ingress with TLS termination
- ✅ Network policies (if configured)
- ✅ Internal service communication via ClusterIP

### Security Improvements to Consider

- [ ] Image scanning in CI/CD pipeline
- [ ] Signed container images
- [ ] Security context policies (Kubernetes)
- [ ] Network policies for external service access
- [ ] Regular security updates
- [ ] Vulnerability scanning
- [ ] Runtime security monitoring

---

## Performance Optimization

### Resource Limits

**API Service:**
- Requests: 256Mi memory, 250m CPU
- Limits: 512Mi memory, 500m CPU

**Frontend Service:**
- Requests: 128Mi memory, 100m CPU
- Limits: 256Mi memory, 200m CPU

**Adjust based on:**
- Expected traffic load
- Response time requirements
- Available cluster resources

### Autoscaling

**Horizontal Pod Autoscaler (HPA):**
- Min replicas: 3
- Max replicas: 10
- CPU target: 70%
- Memory target: 80%
- Scale-up: Aggressive (100% increase, 2 pods max per 15s)
- Scale-down: Conservative (50% decrease per 60s, 5min stabilization)

**Configuration:**
```bash
# View HPA status
kubectl get hpa -n neuronip

# Adjust HPA
kubectl edit hpa neuronip-api-hpa -n neuronip
```

### Build Optimizations

**Current:**
- Layer caching in Docker builds
- GitHub Actions cache for dependencies
- Multi-stage builds for smaller images

**Potential Improvements:**
- Next.js standalone output mode (reduce frontend image size)
- BuildKit cache mounts (faster builds)
- Multi-architecture builds (ARM64 support)
- Image compression

---

## Monitoring and Observability

### Health Checks

**Configured Endpoints:**
- API: `/health` (liveness and readiness)
- Frontend: `/` (liveness and readiness)

**Probe Configuration:**
- Liveness: 30s initial delay, 10s period
- Readiness: 10s initial delay, 5s period

**Manual Health Check:**
```bash
# API
curl http://localhost:8082/health

# From Kubernetes
kubectl get pods -n neuronip
kubectl describe pod <pod-name> -n neuronip
```

### Metrics

**Available Endpoints:**
- Prometheus metrics: `/metrics` (if configured)

**Missing (to be implemented):**
- Prometheus metrics endpoints
- Distributed tracing
- APM integration
- Custom business metrics

### Logging

**Current:**
- Structured logging (JSON format)
- Log levels configurable
- Container logs accessible

**To Implement:**
- Log aggregation (ELK, Loki, etc.)
- Centralized log storage
- Log retention policies
- Log analysis and alerting

### External Service Monitoring

Monitor external service health:
- PostgreSQL connection pool status
- NeuronDB availability
- NeuronMCP service health
- NeuronAgent response times

---

## Troubleshooting

### Common Issues

**API cannot connect to external services:**
```bash
# Verify network connectivity
kubectl exec -n neuronip deployment/neuronip-api -- nc -zv <service-host> <port>

# Check DNS resolution
kubectl exec -n neuronip deployment/neuronip-api -- nslookup <service-host>

# Verify credentials
kubectl get secret neuronip-secrets -n neuronip -o yaml
```

**Frontend cannot reach API:**
```bash
# Check API service
kubectl get svc neuronip-api -n neuronip

# Verify NEXT_PUBLIC_API_URL
kubectl get deployment neuronip-frontend -n neuronip -o yaml | grep NEXT_PUBLIC_API_URL

# Test from frontend pod
kubectl exec -n neuronip deployment/neuronip-frontend -- wget -O- http://neuronip-api:8082/health
```

**High resource usage:**
```bash
# Check resource usage
kubectl top pods -n neuronip

# View HPA status
kubectl get hpa -n neuronip

# Adjust resource limits if needed
kubectl edit deployment neuronip-api -n neuronip
```

---

## 📚 Related Documentation

- [Docker Deployment](docker.md) - Docker setup and configuration
- [Production Packaging](packaging.md) - How images are built and packaged
- [Kubernetes Deployment](kubernetes.md) - Kubernetes deployment guide
- [Monitoring](monitoring.md) - Observability and monitoring setup
- [Configuration Reference](../reference/configuration.md) - Complete configuration options

---

<div align="center">

[← Back to Documentation](../README.md)

</div>
