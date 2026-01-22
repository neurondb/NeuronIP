# 🚀 NeuronIP Quick Start

## Run Everything

The easiest way to run NeuronIP is using the `run_neuronip.sh` script:

```bash
./run_neuronip.sh
```

This will:
- ✅ Build Docker images (if needed)
- ✅ Create network
- ✅ Start server (neuronip-server) on port 8082
- ✅ Start UI (neuronip-ui) on port 3001
- ✅ Check health and show status

## Available Commands

```bash
# Start everything (default)
./run_neuronip.sh
./run_neuronip.sh run

# Build images only
./run_neuronip.sh build

# View status
./run_neuronip.sh status

# View logs
./run_neuronip.sh logs server    # Server logs
./run_neuronip.sh logs ui        # UI logs
./run_neuronip.sh logs all       # All logs

# Stop services
./run_neuronip.sh stop

# Restart services
./run_neuronip.sh restart

# Clean up (remove containers and images)
./run_neuronip.sh clean

# Help
./run_neuronip.sh help
```

## Custom Configuration

You can override defaults using environment variables:

```bash
# Custom database
DB_HOST=myhost DB_USER=myuser DB_PASSWORD=mypass ./run_neuronip.sh

# Custom ports
NEURONIP_API_PORT=9090 NEURONIP_UI_PORT=4000 ./run_neuronip.sh

# All options
DB_HOST=localhost \
DB_PORT=5432 \
DB_USER=postgres \
DB_PASSWORD=secret \
DB_NAME=neuronip \
NEURONIP_API_PORT=8082 \
NEURONIP_UI_PORT=3001 \
./run_neuronip.sh
```

## Service Names

- **neuronip-server** - Backend API (was: neuronip-api)
- **neuronip-ui** - Frontend UI (was: neuronip-frontend)

## Access URLs

After running, access:
- **UI**: http://localhost:3001
- **API**: http://localhost:8082
- **Health**: http://localhost:8082/health

## View Demo Data

1. Open http://localhost:3001/dashboard/warehouse
2. Press F12 → Console
3. Run: `localStorage.setItem('api_token', 'test-key-82f13cedd19abec5bdd9ffad70f3f774'); location.reload();`
4. You'll see 3 warehouse schemas and 8 data sources!

## Troubleshooting

**Services won't start?**
```bash
# Check Docker is running
docker info

# View logs
./run_neuronip.sh logs server
./run_neuronip.sh logs ui
```

**Port already in use?**
```bash
# Use different ports
NEURONIP_API_PORT=9090 NEURONIP_UI_PORT=4000 ./run_neuronip.sh
```

**Need to rebuild?**
```bash
./run_neuronip.sh build
./run_neuronip.sh run
```
