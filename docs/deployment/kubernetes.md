# Kubernetes Deployment Guide

This guide covers deploying NeuronIP on Kubernetes.

## Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured
- Helm 3.x (optional, for Helm chart deployment)
- PostgreSQL 16+ with NeuronDB extension
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

- `DB_HOST`: PostgreSQL host
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name (default: neuronip)
- `SERVER_PORT`: API server port (default: 8082)
- `NEURONAGENT_ENDPOINT`: NeuronAgent API endpoint
- `NEURONAGENT_API_KEY`: NeuronAgent API key

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

For production, use:
- PostgreSQL with streaming replication
- Connection pooling (PgBouncer)
- Read replicas for read-heavy workloads

## Monitoring

### Health Checks

- Liveness probe: `/health` endpoint
- Readiness probe: `/health` endpoint
- Initial delay: 30s (liveness), 10s (readiness)

### Metrics

Prometheus metrics available at `/metrics` endpoint.

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
