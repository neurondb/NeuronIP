#!/bin/bash
# Seed Demo Data via API Script
# This script creates demo data by making API calls

set -e

API_URL="${API_URL:-http://localhost:8082/api/v1}"
API_KEY="${API_KEY:-}"

echo "=========================================="
echo "NeuronIP Demo Data Seeder (via API)"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}API URL: ${API_URL}${NC}"
echo ""

# Check if API is accessible
echo -e "${BLUE}Checking API health...${NC}"
if ! curl -s -f "${API_URL}/../health" > /dev/null 2>&1; then
    echo -e "${RED}Error: API is not accessible at ${API_URL}${NC}"
    echo -e "${YELLOW}Please make sure the API server is running${NC}"
    exit 1
fi
echo -e "${GREEN}✓ API is accessible${NC}"
echo ""

# Create auth header if API key is provided
AUTH_HEADER=""
if [ -n "$API_KEY" ]; then
    AUTH_HEADER="-H \"Authorization: Bearer ${API_KEY}\""
fi

echo -e "${BLUE}Creating demo users...${NC}"
curl -s -X POST "${API_URL}/users" \
    -H "Content-Type: application/json" \
    $AUTH_HEADER \
    -d '{
        "email": "demo@example.com",
        "name": "Demo User",
        "password": "demo123",
        "role": "admin"
    }' || echo "User may already exist"

curl -s -X POST "${API_URL}/users" \
    -H "Content-Type: application/json" \
    $AUTH_HEADER \
    -d '{
        "email": "john@example.com",
        "name": "John Doe",
        "password": "demo123",
        "role": "user"
    }' || echo "User may already exist"

curl -s -X POST "${API_URL}/users" \
    -H "Content-Type: application/json" \
    $AUTH_HEADER \
    -d '{
        "email": "jane@example.com",
        "name": "Jane Smith",
        "password": "demo123",
        "role": "user"
    }' || echo "User may already exist"

echo -e "${GREEN}✓ Users created${NC}"
echo ""

echo -e "${BLUE}Creating demo API keys...${NC}"
curl -s -X POST "${API_URL}/api-keys" \
    -H "Content-Type: application/json" \
    $AUTH_HEADER \
    -d '{
        "name": "Demo API Key",
        "rate_limit": 1000
    }' || echo "API key may already exist"

echo -e "${GREEN}✓ API keys created${NC}"
echo ""

echo -e "${BLUE}Creating demo support tickets...${NC}"
TICKET1=$(curl -s -X POST "${API_URL}/support/tickets" \
    -H "Content-Type: application/json" \
    $AUTH_HEADER \
    -d '{
        "customer_id": "customer-123",
        "subject": "Password reset not working",
        "description": "I tried to reset my password but did not receive the email.",
        "priority": "high"
    }' || echo "")

echo -e "${GREEN}✓ Support tickets created${NC}"
echo ""

echo -e "${BLUE}Creating demo saved searches...${NC}"
curl -s -X POST "${API_URL}/warehouse/saved-searches" \
    -H "Content-Type: application/json" \
    $AUTH_HEADER \
    -d '{
        "name": "Top Products by Revenue",
        "description": "Find products with highest revenue",
        "query": "SELECT name, SUM(total_amount) as revenue FROM orders JOIN order_items ON orders.order_id = order_items.order_id GROUP BY name ORDER BY revenue DESC LIMIT 10",
        "is_public": true
    }' || echo "Search may already exist"

curl -s -X POST "${API_URL}/warehouse/saved-searches" \
    -H "Content-Type: application/json" \
    $AUTH_HEADER \
    -d '{
        "name": "Active Customers",
        "description": "List all active customers",
        "query": "SELECT * FROM customers WHERE registration_date > NOW() - INTERVAL \"30 days\"",
        "is_public": false
    }' || echo "Search may already exist"

echo -e "${GREEN}✓ Saved searches created${NC}"
echo ""

echo -e "${BLUE}Uploading knowledge base documents...${NC}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEMO_DIR="${SCRIPT_DIR}/../examples/demos"

if [ -f "${DEMO_DIR}/knowledge-base-demo.json" ]; then
    echo "Knowledge base file found - documents can be uploaded via UI"
fi

echo -e "${GREEN}✓ Knowledge base ready${NC}"
echo ""

echo -e "${BLUE}Creating warehouse schemas...${NC}"
if [ -f "${DEMO_DIR}/warehouse-sales-demo.json" ]; then
    WAREHOUSE_DATA=$(cat "${DEMO_DIR}/warehouse-sales-demo.json")
    curl -s -X POST "${API_URL}/warehouse/schemas" \
        -H "Content-Type: application/json" \
        $AUTH_HEADER \
        -d "$WAREHOUSE_DATA" || echo "Schema may already exist"
fi

echo -e "${GREEN}✓ Warehouse schemas created${NC}"
echo ""

echo -e "${GREEN}=========================================="
echo "Demo data seeding complete!"
echo "=========================================="
echo ""
echo -e "${YELLOW}You can now access the application with demo data:${NC}"
echo "  - Frontend: http://localhost:3001"
echo "  - API: ${API_URL}"
echo ""
echo -e "${YELLOW}Demo credentials:${NC}"
echo "  - Email: demo@example.com"
echo "  - Password: demo123"
echo ""
echo -e "${YELLOW}Demo data includes:${NC}"
echo "  ✓ 3 Demo users"
echo "  ✓ API keys"
echo "  ✓ Support tickets"
echo "  ✓ Saved searches"
echo "  ✓ Knowledge base (upload via UI)"
echo "  ✓ Warehouse schemas"
echo ""
