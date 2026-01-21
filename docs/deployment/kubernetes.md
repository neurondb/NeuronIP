# Kubernetes Deployment Guide

This guide covers deploying NeuronIP on Kubernetes.

> **Important**: External services (PostgreSQL, NeuronDB, NeuronMCP, NeuronAgent) are **not** deployed as part of NeuronIP and must be deployed separately. The NeuronIP services connect to these external services via configured endpoints.

## Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured
- Helm 3.x (optional, for Helm chart deployment)
- **External Services** (deployed separately):
  - PostgreSQL 16+ with NeuronDB extension
  - NeuronDB service
  - NeuronMCP service
  - NeuronAgent service
- Ingress controller (nginx recommended)

## Quick Start with Helm

```bash
# Add Helm repository (if using a chart repository)
helm repo add neuronip https://charts.neuronip.ai
helm repo update

# Install NeuronIP
helm install neuronip neuronip/neuronip \
  --namespace neuronip \
  --create-namespace \
  --set database.host=postgres-service \
  --set database.user=neuronip \
  --set database.password=YOUR_PASSWORD
```

## Manual Deployment

### 1. Create Namespace

```bash
kubectl apply -f k8s/namespace.yaml
```

### 2. Create Secrets

Edit `k8s/secrets.yaml` with your actual values, then:

```bash
kubectl apply -f k8s/secrets.yaml
```

### 3. Create ConfigMap

```bash
kubectl apply -f k8s/configmap.yaml
```

### 4. Deploy Services

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

### 5. Set Up Ingress

```bash
kubectl apply -f k8s/ingress.yaml
```

### 6. Enable Autoscaling

```bash
kubectl apply -f k8s/hpa.yaml
```

## Configuration

### Environment Variables

Key environment variables for the API deployment:

**Database Connection (External PostgreSQL):**
- `DB_HOST`: External PostgreSQL host (e.g., `postgres.example.com`)
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password (from Kubernetes secrets)
- `DB_NAME`: Database name (default: neuronip)

**Server Configuration:**
- `SERVER_PORT`: API server port (default: 8082)

**External Service Connections:**
- `NEURONDB_HOST`: External NeuronDB host
- `NEURONDB_PORT`: NeuronDB port
- `NEURONDB_DATABASE`: NeuronDB database name
- `NEURONAGENT_ENDPOINT`: External NeuronAgent API endpoint (e.g., `https://neuronagent.example.com:8080`)
- `NEURONAGENT_API_KEY`: NeuronAgent API key (from Kubernetes secrets)

> **Note**: All external service endpoints must be accessible from the Kubernetes cluster. Configure network policies and DNS as needed.

### Resource Limits

Default resource requests and limits:

- API: 256Mi memory, 250m CPU (requests); 512Mi memory, 500m CPU (limits)
- Frontend: 128Mi memory, 100m CPU (requests); 256Mi memory, 200m CPU (limits)

Adjust in `k8s/deployment.yaml` or Helm values.

## High Availability

### Horizontal Pod Autoscaling

HPA is configured to:
- Scale between 3-10 replicas
- Target 70% CPU utilization
- Target 80% memory utilization

### Database High Availability

For production, configure external PostgreSQL with:
- PostgreSQL with streaming replication
- Connection pooling (PgBouncer)
- Read replicas for read-heavy workloads

> **Note**: Database high availability is managed separately from NeuronIP deployment. Ensure your external PostgreSQL service is configured for high availability.

## Monitoring

### Health Checks

- Liveness probe: `/health` endpoint
- Readiness probe: `/health` endpoint
- Initial delay: 30s (liveness), 10s (readiness)

### Metrics

Prometheus metrics available at `/metrics` endpoint.

## External Service Configuration

### Connecting to External Services

NeuronIP requires external services to be accessible from the Kubernetes cluster:

1. **PostgreSQL**: Configure `DB_HOST` in secrets to point to external PostgreSQL
2. **NeuronDB**: Configure `NEURONDB_HOST` environment variable
3. **NeuronMCP**: Configure MCP service endpoint
4. **NeuronAgent**: Configure `NEURONAGENT_ENDPOINT` environment variable

### Network Configuration

**Option 1: External IP Addresses**
- Configure external services with static IPs
- Update `DB_HOST` and other endpoints to use IP addresses or FQDNs
- Ensure network policies allow outbound connections

**Option 2: ExternalName Services**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: external-postgres
  namespace: neuronip
spec:
  type: ExternalName
  externalName: postgres.example.com
```

**Option 3: Service Endpoints**
- Create Kubernetes Service with Endpoints pointing to external IPs
- Use service name in `DB_HOST` instead of external address

### Testing Connectivity

```bash
# Test from API pod
kubectl exec -n neuronip deployment/neuronip-api -- nc -zv postgres.example.com 5432
kubectl exec -n neuronip deployment/neuronip-api -- nc -zv neurondb.example.com 5432

# Test DNS resolution
kubectl exec -n neuronip deployment/neuronip-api -- nslookup postgres.example.com
```

### Secrets Configuration

Store external service credentials in Kubernetes secrets:

```bash
# Create secret for database
kubectl create secret generic neuronip-secrets \
  --from-literal=db-host=postgres.example.com \
  --from-literal=db-user=neuronip \
  --from-literal=db-password=secure_password \
  -n neuronip
```

See [Production Deployment Guide](production.md) for detailed external service configuration.

---

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n neuronip
```

### View Logs

```bash
kubectl logs -n neuronip deployment/neuronip-api
kubectl logs -n neuronip deployment/neuronip-frontend
```

### Check Events

```bash
kubectl get events -n neuronip --sort-by='.lastTimestamp'
```

## Multi-Region Deployment

For multi-region deployments, see `docs/deployment/multi-region.md`.

---

## 📚 Related Documentation

- [Production Deployment](production.md) - Production deployment guide
- [Production Packaging](packaging.md) - How images are built and packaged
- [Docker Deployment](docker.md) - Docker setup
- [Monitoring](monitoring.md) - Observability and monitoring setup
