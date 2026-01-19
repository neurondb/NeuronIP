# NeuronIP - AI-Native Enterprise Intelligence Platform

NeuronIP is a comprehensive enterprise intelligence platform that combines five core capabilities into a unified system:

- **Semantic Knowledge Search** - Search your entire knowledge base by meaning
- **Data Warehouse Q&A** - Ask questions and get SQL + charts + explanation
- **Customer Support Memory** - Automate support with AI agents and long-term memory
- **Compliance & Audit Analytics** - Policy matching, anomaly detection, semantic filtering
- **Agent Workflows** - Long-term memory and workflow execution powered by NeuronDB

## Architecture

NeuronIP is built as a standalone application with:
- **Backend**: Go-based REST API
- **Frontend**: Next.js with TypeScript
- **Database**: PostgreSQL with NeuronDB extension
- **Integrations**: NeuronDB, NeuronAgent, NeuronMCP

## Quick Start

### Prerequisites

- Docker and Docker Compose
- PostgreSQL 16+ with NeuronDB extension
- Go 1.24+ (for building backend)
- Node.js 18+ (for building frontend)

### Using Docker Compose

From the repository root:

```bash
# Start all services including NeuronIP
docker compose up -d

# Access NeuronIP
# Frontend: http://localhost:3001
# Backend API: http://localhost:8082
```

### Manual Setup

#### Backend

```bash
cd NeuronIP/api

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5433
export DB_USER=neuronip
export DB_PASSWORD=neuronip
export DB_NAME=neuronip

# Initialize database
psql -d neuronip -f ../neuronip.sql

# Run server
go run cmd/server/main.go
```

#### Frontend

```bash
cd NeuronIP/frontend

# Install dependencies
npm install

# Run development server
npm run dev
```

## API Documentation

### Health Check

```bash
curl http://localhost:8082/health
```

### Semantic Search

```bash
curl -X POST http://localhost:8082/api/v1/semantic/search \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "machine learning algorithms",
    "limit": 10
  }'
```

## Configuration

See `api/internal/config/config.go` for all configuration options.

Key environment variables:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - Database connection
- `SERVER_PORT` - API server port (default: 8082)
- `NEURONAGENT_ENDPOINT` - NeuronAgent API endpoint
- `NEURONAGENT_API_KEY` - NeuronAgent API key

## License

See [LICENSE](../LICENSE) file for license information.
