#!/bin/bash
# Load Demo Data Script
# Runs the demo-data.sql file to populate the database with realistic data

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-neuronip}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-}"
PGPASSWORD="${DB_PASSWORD}"
DEMO_FILE="${DEMO_FILE:-./demo/demo-data.sql}"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--host)
            DB_HOST="$2"
            shift 2
            ;;
        -p|--port)
            DB_PORT="$2"
            shift 2
            ;;
        -d|--database)
            DB_NAME="$2"
            shift 2
            ;;
        -U|--user)
            DB_USER="$2"
            shift 2
            ;;
        -W|--password)
            read -s -p "Password: " DB_PASSWORD
            echo
            PGPASSWORD="$DB_PASSWORD"
            shift
            ;;
        -f|--file)
            DEMO_FILE="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Load demo data into NeuronIP database"
            echo ""
            echo "Options:"
            echo "  -h, --host HOST          Database host (default: localhost)"
            echo "  -p, --port PORT          Database port (default: 5432)"
            echo "  -d, --database NAME      Database name (default: neuronip)"
            echo "  -U, --user USER          Database user (default: postgres)"
            echo "  -W, --password           Prompt for password"
            echo "  -f, --file FILE          Demo data SQL file (default: ./demo/demo-data.sql)"
            echo "  --help                   Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD"
            echo ""
            echo "Examples:"
            echo "  $0"
            echo "  $0 -h localhost -d neuronip -U postgres"
            echo "  $0 --file ./demo/custom-demo-data.sql"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Check if demo file exists
if [ ! -f "$DEMO_FILE" ]; then
    echo -e "${RED}Error: Demo data file '$DEMO_FILE' does not exist${NC}"
    exit 1
fi

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql command not found. Please install PostgreSQL client tools.${NC}"
    exit 1
fi

# Export password for psql
export PGPASSWORD

# Build connection string
CONN_STRING="host=$DB_HOST port=$DB_PORT dbname=$DB_NAME user=$DB_USER"

# Test database connection
echo -e "${BLUE}Testing database connection...${NC}"
if ! psql "$CONN_STRING" -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}Error: Cannot connect to database${NC}"
    echo "  Host: $DB_HOST"
    echo "  Port: $DB_PORT"
    echo "  Database: $DB_NAME"
    echo "  User: $DB_USER"
    exit 1
fi
echo -e "${GREEN}✓ Database connection successful${NC}"
echo ""

# Check if migrations have been run
echo -e "${BLUE}Checking if migrations have been run...${NC}"
TABLE_COUNT=$(psql "$CONN_STRING" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'neuronip';" 2>/dev/null | tr -d ' ' || echo "0")

if [ "$TABLE_COUNT" -eq "0" ]; then
    echo -e "${YELLOW}Warning: No tables found in 'neuronip' schema${NC}"
    echo "  Please run migrations first: ./scripts/run-migrations.sh"
    read -p "  Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo -e "${GREEN}✓ Found $TABLE_COUNT tables in neuronip schema${NC}"
fi
echo ""

# Load demo data
echo -e "${BLUE}Loading demo data from '$DEMO_FILE'...${NC}"
echo ""

if psql "$CONN_STRING" -f "$DEMO_FILE" 2>&1; then
    echo ""
    echo -e "${GREEN}✓ Demo data loaded successfully!${NC}"
    echo ""
    echo -e "${BLUE}Verification:${NC}"
    
    # Show some counts
    echo -e "  Users: $(psql "$CONN_STRING" -t -c "SELECT COUNT(*) FROM neuronip.users;" 2>/dev/null | tr -d ' ')"
    echo -e "  Search History: $(psql "$CONN_STRING" -t -c "SELECT COUNT(*) FROM neuronip.search_history;" 2>/dev/null | tr -d ' ')"
    echo -e "  Warehouse Queries: $(psql "$CONN_STRING" -t -c "SELECT COUNT(*) FROM neuronip.warehouse_queries;" 2>/dev/null | tr -d ' ')"
    echo -e "  Workflows: $(psql "$CONN_STRING" -t -c "SELECT COUNT(*) FROM neuronip.workflows;" 2>/dev/null | tr -d ' ')"
    echo -e "  Agents: $(psql "$CONN_STRING" -t -c "SELECT COUNT(*) FROM neuronip.agents;" 2>/dev/null | tr -d ' ')"
    echo -e "  Support Tickets: $(psql "$CONN_STRING" -t -c "SELECT COUNT(*) FROM neuronip.support_tickets;" 2>/dev/null | tr -d ' ')"
    echo -e "  Audit Logs: $(psql "$CONN_STRING" -t -c "SELECT COUNT(*) FROM neuronip.audit_logs;" 2>/dev/null | tr -d ' ')"
    echo ""
    echo -e "${GREEN}Demo data is ready! You can now start the application and see populated pages.${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}✗ Error loading demo data${NC}"
    echo "  Check the error messages above for details"
    exit 1
fi
