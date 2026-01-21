#!/bin/bash
# Seed Demo Data Script
# This script seeds the database with demo data for all features

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_DIR="${SCRIPT_DIR}/../api"

echo "=========================================="
echo "NeuronIP Demo Data Seeder"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

cd "$API_DIR"

echo -e "${BLUE}Step 1: Running database migrations...${NC}"
if ! go run cmd/migrate/main.go -command up; then
    echo -e "${YELLOW}Warning: Some migrations may have already been applied${NC}"
fi
echo ""

echo -e "${BLUE}Step 2: Seeding demo data...${NC}"
go run cmd/seed/main.go -type demo -clear
echo ""

echo -e "${GREEN}✓ Demo data seeded successfully!${NC}"
echo ""
echo -e "${YELLOW}You can now access the application with demo data:${NC}"
echo "  - Frontend: http://localhost:3001"
echo "  - API: http://localhost:8082/api/v1"
echo ""
echo -e "${YELLOW}Demo data includes:${NC}"
echo "  ✓ Users"
echo "  ✓ Support tickets with conversations"
echo "  ✓ Knowledge base documents"
echo "  ✓ Warehouse schemas and sample data"
echo "  ✓ Saved searches"
echo "  ✓ API keys"
echo "  ✓ Workflows"
echo "  ✓ Metrics"
echo ""
