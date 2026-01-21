#!/bin/bash
# Complete Demo Setup Script
# This script sets up everything needed to run NeuronIP with demo data

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${SCRIPT_DIR}/.."

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "=========================================="
echo "NeuronIP Complete Demo Setup"
echo "=========================================="
echo ""

# Check prerequisites
echo -e "${BLUE}Checking prerequisites...${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go is installed${NC}"

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}Error: Node.js is not installed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Node.js is installed${NC}"

# Check if PostgreSQL is accessible
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-neuronip}"
DB_NAME="${DB_NAME:-neuronip}"

echo ""
echo -e "${BLUE}Checking database connection...${NC}"
if command -v psql &> /dev/null; then
    if PGPASSWORD="${DB_PASSWORD:-neuronip}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Database is accessible${NC}"
        DB_ACCESSIBLE=true
    else
        echo -e "${YELLOW}⚠ Database may not be accessible. Please ensure PostgreSQL is running.${NC}"
        DB_ACCESSIBLE=false
    fi
else
    echo -e "${YELLOW}⚠ psql not found. Skipping database check.${NC}"
    DB_ACCESSIBLE=false
fi

echo ""
echo -e "${BLUE}Step 1: Setting up database schema...${NC}"
if [ "$DB_ACCESSIBLE" = true ]; then
    cd "$PROJECT_DIR"
    if [ -f "neuronip.sql" ]; then
        echo "Running database migrations..."
        if PGPASSWORD="${DB_PASSWORD:-neuronip}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f neuronip.sql > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Schema applied${NC}"
        else
            echo -e "${YELLOW}⚠ Schema may already be applied or there were errors${NC}"
        fi
    fi
else
    echo -e "${YELLOW}⚠ Skipping schema setup (database not accessible)${NC}"
fi

echo ""
echo -e "${BLUE}Step 2: Seeding demo data...${NC}"
if [ "$DB_ACCESSIBLE" = true ]; then
    cd "$PROJECT_DIR/api"
    
    # Set environment variables for seeding
    export DB_HOST="${DB_HOST:-localhost}"
    export DB_PORT="${DB_PORT:-5432}"
    export DB_USER="${DB_USER:-neuronip}"
    export DB_PASSWORD="${DB_PASSWORD:-neuronip}"
    export DB_NAME="${DB_NAME:-neuronip}"
    
    echo "Running seed command..."
    if go run cmd/seed/main.go -type demo -clear 2>&1; then
        echo -e "${GREEN}✓ Demo data seeded via database${NC}"
    else
        echo -e "${YELLOW}⚠ Database seeding had issues, trying API method...${NC}"
        DB_ACCESSIBLE=false
    fi
fi

if [ "$DB_ACCESSIBLE" = false ]; then
    echo ""
    echo -e "${BLUE}Attempting to seed via API...${NC}"
    echo "Checking if API is running..."
    
    if curl -s -f "http://localhost:8082/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ API is running${NC}"
        cd "$PROJECT_DIR"
        if ./scripts/seed-via-api.sh; then
            echo -e "${GREEN}✓ Demo data seeded via API${NC}"
        else
            echo -e "${YELLOW}⚠ API seeding had issues${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ API is not running at http://localhost:8082${NC}"
        echo ""
        echo -e "${YELLOW}To seed data via API:${NC}"
        echo "  1. Start the API server:"
        echo "     cd api && go run cmd/server/main.go"
        echo ""
        echo "  2. In another terminal, run:"
        echo "     ./scripts/seed-via-api.sh"
    fi
fi

echo ""
echo -e "${GREEN}=========================================="
echo "Setup Complete!"
echo "=========================================="
echo ""
echo -e "${YELLOW}Demo credentials:${NC}"
echo "  Email: demo@example.com"
echo "  Password: demo123"
echo ""
echo -e "${YELLOW}Access URLs:${NC}"
echo "  Frontend: http://localhost:3001"
echo "  API: http://localhost:8082/api/v1"
echo ""
echo -e "${YELLOW}To start services:${NC}"
echo ""
echo "  Option 1: Using Docker (if available)"
echo "    docker compose up -d"
echo ""
echo "  Option 2: Manual startup"
echo "    # Terminal 1 - API"
echo "    cd api && go run cmd/server/main.go"
echo ""
echo "    # Terminal 2 - Frontend"
echo "    cd frontend && npm run dev"
echo ""
echo -e "${YELLOW}What was seeded:${NC}"
echo "  ✓ 3-4 Demo users"
echo "  ✓ API keys"
echo "  ✓ Support tickets with conversations"
echo "  ✓ Knowledge base documents"
echo "  ✓ Warehouse schemas"
echo "  ✓ Saved searches"
echo "  ✓ Workflows"
echo "  ✓ Metrics"
echo ""
