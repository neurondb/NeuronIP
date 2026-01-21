#!/bin/bash
# Build and Run NeuronIP Services
# This script builds and runs the NeuronIP API and Frontend services

set -e

echo "🚀 Building and Running NeuronIP Services"
echo "=========================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker daemon is not running!"
    echo "Please start Docker Desktop or Docker daemon and try again."
    exit 1
fi

# Navigate to project root
cd "$(dirname "$0")/.."

# Create network if it doesn't exist
echo "📡 Creating network..."
docker network create neurondb-network 2>/dev/null || echo "Network already exists"

# Build API image
echo ""
echo "🔨 Building API image..."
cd api
docker build -t neuronip-api:local .
cd ..

# Build Frontend image
echo ""
echo "🔨 Building Frontend image..."
cd frontend
docker build -t neuronip-frontend:local .
cd ..

# Check if external services are needed
echo ""
echo "⚠️  Note: This script builds the NeuronIP services only."
echo "External services (PostgreSQL, NeuronDB, NeuronMCP, NeuronAgent) must be running separately."
echo ""

# Ask if user wants to run the services
read -p "Do you want to run the services now? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "🚀 Starting services..."
    
    # Run API (assuming external services are configured)
    echo "Starting API on port 8082..."
    docker run -d \
        --name neuronip-api \
        --network neurondb-network \
        -p 8082:8082 \
        -e DB_HOST=${DB_HOST:-postgres} \
        -e DB_PORT=${DB_PORT:-5432} \
        -e DB_USER=${DB_USER:-neurondb} \
        -e DB_PASSWORD=${DB_PASSWORD:-neurondb} \
        -e DB_NAME=${DB_NAME:-neuronip} \
        -e SERVER_PORT=8082 \
        -e NEURONDB_HOST=${NEURONDB_HOST:-neurondb} \
        -e NEURONDB_PORT=${NEURONDB_PORT:-5432} \
        -e NEURONDB_DATABASE=${NEURONDB_DATABASE:-neurondb} \
        -e NEURONAGENT_ENDPOINT=${NEURONAGENT_ENDPOINT:-http://neuronagent:8080} \
        neuronip-api:local
    
    # Run Frontend
    echo "Starting Frontend on port 3001..."
    docker run -d \
        --name neuronip-frontend \
        --network neurondb-network \
        -p 3001:3000 \
        -e NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL:-http://localhost:8082/api/v1} \
        neuronip-frontend:local
    
    echo ""
    echo "✅ Services started!"
    echo ""
    echo "📊 Service Status:"
    docker ps --filter "name=neuronip" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    echo ""
    echo "🌐 Access points:"
    echo "  - Frontend: http://localhost:3001"
    echo "  - API: http://localhost:8082"
    echo "  - API Health: http://localhost:8082/health"
    echo ""
    echo "📝 View logs:"
    echo "  - API: docker logs -f neuronip-api"
    echo "  - Frontend: docker logs -f neuronip-frontend"
    echo ""
    echo "🛑 Stop services:"
    echo "  docker stop neuronip-api neuronip-frontend"
    echo "  docker rm neuronip-api neuronip-frontend"
else
    echo ""
    echo "✅ Images built successfully!"
    echo ""
    echo "📦 Built images:"
    docker images | grep neuronip
    echo ""
    echo "To run services manually:"
    echo "  docker run -d --name neuronip-api -p 8082:8082 neuronip-api:local"
    echo "  docker run -d --name neuronip-frontend -p 3001:3000 neuronip-frontend:local"
fi
