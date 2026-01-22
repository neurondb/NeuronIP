#!/bin/bash
# NeuronIP Run Script
# Modular script to build and run NeuronIP services
# Usage: ./run_neuronip.sh [command] [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
DEFAULT_API_PORT=8082
DEFAULT_UI_PORT=3001
DEFAULT_DB_HOST=host.docker.internal
DEFAULT_DB_PORT=5432
DEFAULT_DB_USER=${USER:-postgres}
DEFAULT_DB_PASSWORD=""
DEFAULT_DB_NAME=neuronip
DEFAULT_NETWORK=neurondb-network

# Service names
SERVER_NAME=neuronip-server
UI_NAME=neuronip-ui
SERVER_IMAGE=neuronip-server:local
UI_IMAGE=neuronip-ui:local

# Configuration (can be overridden by environment variables)
API_PORT=${NEURONIP_API_PORT:-$DEFAULT_API_PORT}
UI_PORT=${NEURONIP_UI_PORT:-$DEFAULT_UI_PORT}
DB_HOST=${DB_HOST:-$DEFAULT_DB_HOST}
DB_PORT=${DB_PORT:-$DEFAULT_DB_PORT}
DB_USER=${DB_USER:-$DEFAULT_DB_USER}
DB_PASSWORD=${DB_PASSWORD:-$DEFAULT_DB_PASSWORD}
DB_NAME=${DB_NAME:-$DEFAULT_DB_NAME}
NETWORK=${NEURONIP_NETWORK:-$DEFAULT_NETWORK}

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✅${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}❌${NC} $1"
}

# Check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker daemon is not running!"
        echo "Please start Docker Desktop or Docker daemon and try again."
        exit 1
    fi
}

# Create network if it doesn't exist
setup_network() {
    if ! docker network inspect "$NETWORK" > /dev/null 2>&1; then
        log_info "Creating network: $NETWORK"
        docker network create "$NETWORK" > /dev/null 2>&1
        log_success "Network created"
    else
        log_info "Network $NETWORK already exists"
    fi
}

# Build server image
build_server() {
    log_info "Building server image..."
    cd "$SCRIPT_DIR/api"
    if docker build -t "$SERVER_IMAGE" . > /dev/null 2>&1; then
        log_success "Server image built successfully"
    else
        log_error "Failed to build server image"
        return 1
    fi
    cd "$SCRIPT_DIR"
}

# Build UI image
build_ui() {
    log_info "Building UI image..."
    cd "$SCRIPT_DIR/frontend"
    if docker build -t "$UI_IMAGE" . > /dev/null 2>&1; then
        log_success "UI image built successfully"
    else
        log_error "Failed to build UI image"
        return 1
    fi
    cd "$SCRIPT_DIR"
}

# Stop and remove containers
stop_containers() {
    log_info "Stopping containers..."
    docker stop "$SERVER_NAME" "$UI_NAME" 2>/dev/null || true
    docker rm "$SERVER_NAME" "$UI_NAME" 2>/dev/null || true
    log_success "Containers stopped"
}

# Start server container
start_server() {
    log_info "Starting server..."
    docker run -d \
        --name "$SERVER_NAME" \
        -p "$API_PORT:8082" \
        -e DB_HOST="$DB_HOST" \
        -e DB_PORT="$DB_PORT" \
        -e DB_USER="$DB_USER" \
        -e DB_PASSWORD="$DB_PASSWORD" \
        -e DB_NAME="$DB_NAME" \
        -e SERVER_PORT=8082 \
        --network "$NETWORK" \
        "$SERVER_IMAGE" > /dev/null 2>&1
    
    # Wait for server to start
    sleep 3
    
    if docker ps --format "{{.Names}}" | grep -q "^${SERVER_NAME}$"; then
        log_success "Server started on port $API_PORT"
    else
        log_error "Failed to start server"
        docker logs "$SERVER_NAME" --tail 10
        return 1
    fi
}

# Start UI container
start_ui() {
    log_info "Starting UI..."
    docker run -d \
        --name "$UI_NAME" \
        -p "$UI_PORT:3000" \
        -e NEXT_PUBLIC_API_URL="http://localhost:$API_PORT/api/v1" \
        --network "$NETWORK" \
        "$UI_IMAGE" > /dev/null 2>&1
    
    # Wait for UI to start
    sleep 3
    
    if docker ps --format "{{.Names}}" | grep -q "^${UI_NAME}$"; then
        log_success "UI started on port $UI_PORT"
    else
        log_error "Failed to start UI"
        docker logs "$UI_NAME" --tail 10
        return 1
    fi
}

# Check service health
check_health() {
    log_info "Checking service health..."
    
    # Check server
    if curl -s "http://localhost:$API_PORT/health" > /dev/null 2>&1; then
        log_success "Server is healthy"
    else
        log_warning "Server health check failed (may still be starting)"
    fi
    
    # Check UI
    if curl -s "http://localhost:$UI_PORT" > /dev/null 2>&1; then
        log_success "UI is accessible"
    else
        log_warning "UI health check failed (may still be starting)"
    fi
}

# Show status
show_status() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📊 NeuronIP Services Status"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    docker ps --filter "name=$SERVER_NAME" --filter "name=$UI_NAME" \
        --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "No containers running"
    
    echo ""
    echo "🌐 Access URLs:"
    echo "   Server API: http://localhost:$API_PORT"
    echo "   UI:         http://localhost:$UI_PORT"
    echo ""
    echo "📋 Configuration:"
    echo "   Database: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
    echo "   Network:  $NETWORK"
    echo ""
}

# Show logs
show_logs() {
    local service=${1:-all}
    
    case $service in
        server|s)
            docker logs -f "$SERVER_NAME"
            ;;
        ui|u)
            docker logs -f "$UI_NAME"
            ;;
        all|*)
            docker logs -f "$SERVER_NAME" "$UI_NAME"
            ;;
    esac
}

# Main commands
case "${1:-run}" in
    build)
        check_docker
        log_info "Building NeuronIP services..."
        build_server
        build_ui
        log_success "All images built successfully"
        ;;
    
    start|run)
        check_docker
        log_info "Starting NeuronIP services..."
        
        # Build if images don't exist
        if ! docker image inspect "$SERVER_IMAGE" > /dev/null 2>&1; then
            log_warning "Server image not found, building..."
            build_server
        fi
        
        if ! docker image inspect "$UI_IMAGE" > /dev/null 2>&1; then
            log_warning "UI image not found, building..."
            build_ui
        fi
        
        setup_network
        stop_containers
        start_server
        start_ui
        check_health
        show_status
        
        echo ""
        log_success "NeuronIP is running!"
        echo ""
        echo "🔑 To view demo data, set API token in browser console:"
        echo "   localStorage.setItem('api_token', 'test-key-82f13cedd19abec5bdd9ffad70f3f774')"
        echo ""
        ;;
    
    stop)
        check_docker
        log_info "Stopping NeuronIP services..."
        stop_containers
        log_success "Services stopped"
        ;;
    
    restart)
        check_docker
        log_info "Restarting NeuronIP services..."
        stop_containers
        start_server
        start_ui
        check_health
        show_status
        ;;
    
    status|ps)
        show_status
        ;;
    
    logs)
        show_logs "${2:-all}"
        ;;
    
    build-server)
        check_docker
        build_server
        ;;
    
    build-ui)
        check_docker
        build_ui
        ;;
    
    start-server)
        check_docker
        setup_network
        stop_containers
        start_server
        ;;
    
    start-ui)
        check_docker
        setup_network
        start_ui
        ;;
    
    clean)
        check_docker
        log_info "Cleaning up containers and images..."
        stop_containers
        docker rmi "$SERVER_IMAGE" "$UI_IMAGE" 2>/dev/null || true
        log_success "Cleanup complete"
        ;;
    
    help|--help|-h)
        cat << EOF
NeuronIP Run Script

Usage: ./run_neuronip.sh [command] [options]

Commands:
  run, start          Build (if needed) and start all services (default)
  build               Build all Docker images
  stop                Stop all services
  restart             Restart all services
  status, ps          Show service status
  logs [service]      Show logs (server|ui|all)
  build-server        Build only the server image
  build-ui            Build only the UI image
  start-server        Start only the server
  start-ui            Start only the UI
  clean               Remove containers and images
  help                Show this help message

Options (Environment Variables):
  NEURONIP_API_PORT      Server API port (default: $DEFAULT_API_PORT)
  NEURONIP_UI_PORT       UI port (default: $DEFAULT_UI_PORT)
  DB_HOST                Database host (default: $DEFAULT_DB_HOST)
  DB_PORT                Database port (default: $DEFAULT_DB_PORT)
  DB_USER                Database user (default: current user)
  DB_PASSWORD            Database password (default: empty)
  DB_NAME                Database name (default: $DEFAULT_DB_NAME)
  NEURONIP_NETWORK       Docker network (default: $DEFAULT_NETWORK)

Examples:
  ./run_neuronip.sh                    # Start everything
  ./run_neuronip.sh build              # Build images only
  ./run_neuronip.sh logs server        # View server logs
  DB_HOST=myhost ./run_neuronip.sh     # Use custom database host

EOF
        ;;
    
    *)
        log_error "Unknown command: $1"
        echo "Run './run_neuronip.sh help' for usage information"
        exit 1
        ;;
esac
