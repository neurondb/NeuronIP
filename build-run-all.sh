#!/bin/bash
# Build and Run All - Complete Enterprise Implementation
# This script builds the application, runs all migrations, and starts the server

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

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

log_step() {
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}▶${NC} $1"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Check prerequisites
check_prerequisites() {
    log_step "Checking Prerequisites"
    
    # Check Go
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21+"
        exit 1
    fi
    GO_VERSION=$(go version | awk '{print $3}')
    log_success "Go found: $GO_VERSION"
    
    # Check PostgreSQL connection (optional, will warn if not available)
    if command -v psql &> /dev/null; then
        log_success "PostgreSQL client found"
    else
        log_warning "PostgreSQL client not found (migrations may need manual execution)"
    fi
    
    # Check if we're in the right directory
    if [ ! -f "api/go.mod" ]; then
        log_error "Not in NeuronIP root directory. Please run from project root."
        exit 1
    fi
    log_success "Project structure verified"
}

# Build the application
build_application() {
    log_step "Building Application"
    
    cd "$SCRIPT_DIR/api"
    
    log_info "Updating and downloading dependencies..."
    if go mod tidy && go mod download; then
        log_success "Dependencies updated and downloaded"
    else
        log_error "Failed to update/download dependencies"
        exit 1
    fi
    
    log_info "Building server binary..."
    if go build -o bin/server ./cmd/server; then
        log_success "Server built successfully"
    else
        log_error "Failed to build server"
        exit 1
    fi
    
    log_info "Building migration tool..."
    if go build -o bin/migrate ./cmd/migrate; then
        log_success "Migration tool built successfully"
    else
        log_error "Failed to build migration tool"
        exit 1
    fi
    
    cd "$SCRIPT_DIR"
}

# Run all migrations
run_migrations() {
    log_step "Running Database Migrations"
    
    cd "$SCRIPT_DIR/api"
    
    # Check if migration tool exists
    if [ ! -f "bin/migrate" ]; then
        log_warning "Migration tool not found, building..."
        build_application
    fi
    
    log_info "Running all migrations..."
    if ./bin/migrate -command=up; then
        log_success "All migrations applied successfully"
    else
        log_error "Migration failed"
        log_info "You may need to check your database connection"
        log_info "Database config: Check api/.env or environment variables"
        exit 1
    fi
    
    log_info "Checking migration status..."
    ./bin/migrate -command=status
    
    cd "$SCRIPT_DIR"
}

# Verify new migrations exist
verify_new_migrations() {
    log_step "Verifying New Enterprise Migrations"
    
    NEW_MIGRATIONS=(
        "051_clustering.sql"
        "052_observability_enhancements.sql"
        "053_connector_registry.sql"
        "054_streaming_schema.sql"
        "055_ml_lifecycle.sql"
        "056_model_governance.sql"
        "057_budgets.sql"
    )
    
    MISSING=0
    for migration in "${NEW_MIGRATIONS[@]}"; do
        if [ -f "api/migrations/$migration" ]; then
            log_success "Found: $migration"
        else
            log_error "Missing: $migration"
            MISSING=1
        fi
    done
    
    if [ $MISSING -eq 1 ]; then
        log_error "Some migrations are missing!"
        exit 1
    fi
    
    log_success "All enterprise migrations verified"
}

# Start the server
start_server() {
    log_step "Starting Server"
    
    cd "$SCRIPT_DIR/api"
    
    # Check if server binary exists
    if [ ! -f "bin/server" ]; then
        log_warning "Server binary not found, building..."
        build_application
    fi
    
    log_info "Starting NeuronIP server..."
    log_info "Server will run in foreground. Press Ctrl+C to stop."
    echo ""
    
    # Run server
    ./bin/server
}

# Main execution
main() {
    echo ""
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                              ║"
    echo "║                    NeuronIP Enterprise - Build & Run All                    ║"
    echo "║                                                                              ║"
    echo "╚══════════════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo ""
    
    # Parse arguments
    SKIP_BUILD=false
    SKIP_MIGRATIONS=false
    SKIP_VERIFY=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            --skip-migrations)
                SKIP_MIGRATIONS=true
                shift
                ;;
            --skip-verify)
                SKIP_VERIFY=true
                shift
                ;;
            --migrations-only)
                SKIP_BUILD=true
                SKIP_VERIFY=false
                SKIP_MIGRATIONS=false
                shift
                ;;
            --build-only)
                SKIP_MIGRATIONS=true
                SKIP_VERIFY=false
                shift
                ;;
            -h|--help)
                cat << EOF
NeuronIP Build & Run All Script

Usage: ./build-run-all.sh [options]

Options:
  --skip-build          Skip building the application
  --skip-migrations     Skip running migrations
  --skip-verify         Skip verifying new migrations exist
  --migrations-only     Only run migrations (skip build and start)
  --build-only          Only build (skip migrations and start)
  -h, --help            Show this help message

This script will:
  1. Check prerequisites (Go, PostgreSQL client)
  2. Verify new enterprise migrations exist
  3. Build the application (server + migration tool)
  4. Run all database migrations (including new enterprise migrations)
  5. Start the server

Environment Variables:
  The script uses standard database environment variables:
  - DB_HOST (default: localhost)
  - DB_PORT (default: 5432)
  - DB_USER (default: postgres)
  - DB_PASSWORD
  - DB_NAME (default: neuronip)

Examples:
  ./build-run-all.sh                    # Full build, migrate, and run
  ./build-run-all.sh --migrations-only   # Only run migrations
  ./build-run-all.sh --build-only        # Only build
  ./build-run-all.sh --skip-build        # Skip build, just migrate and run

EOF
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Run './build-run-all.sh --help' for usage"
                exit 1
                ;;
        esac
    done
    
    # Execute steps
    check_prerequisites
    
    if [ "$SKIP_VERIFY" = false ]; then
        verify_new_migrations
    fi
    
    if [ "$SKIP_BUILD" = false ]; then
        build_application
    fi
    
    if [ "$SKIP_MIGRATIONS" = false ]; then
        run_migrations
    fi
    
    if [ "$SKIP_BUILD" = false ] && [ "$SKIP_MIGRATIONS" = false ]; then
        echo ""
        log_step "Summary"
        log_success "Build completed"
        log_success "Migrations completed"
        echo ""
        log_info "Enterprise features are now available:"
        echo "  • Clustering & Distributed Systems"
        echo "  • Enterprise Observability"
        echo "  • Enhanced Data Ingestion (50+ connectors)"
        echo "  • Streaming & CDC Pipelines"
        echo "  • ML Lifecycle & Governance"
        echo "  • Advanced Workflows"
        echo "  • Cost Tracking & Budgets"
        echo "  • Integration Marketplace"
        echo "  • Partner Management"
        echo "  • Enterprise Onboarding"
        echo ""
        read -p "Start the server now? (y/n) " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            start_server
        else
            log_info "Server not started. Run './api/bin/server' to start manually."
        fi
    elif [ "$SKIP_MIGRATIONS" = false ]; then
        log_success "Migrations completed"
    elif [ "$SKIP_BUILD" = false ]; then
        log_success "Build completed"
    fi
}

# Run main
main "$@"
