#!/bin/bash
# Run all SQL migration files in sequence
# Files are ordered by their numeric prefix (001_, 002_, etc.)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
SQL_DIR="${SQL_DIR:-./sql}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-neuronip}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-}"
PGPASSWORD="${DB_PASSWORD}"

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
        --sql-dir)
            SQL_DIR="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Run all SQL migration files in sequence from the sql/ directory"
            echo ""
            echo "Options:"
            echo "  -h, --host HOST          Database host (default: localhost)"
            echo "  -p, --port PORT          Database port (default: 5432)"
            echo "  -d, --database NAME      Database name (default: neuronip)"
            echo "  -U, --user USER          Database user (default: postgres)"
            echo "  -W, --password           Prompt for password"
            echo "  --sql-dir DIR            SQL files directory (default: ./sql)"
            echo "  --dry-run                Show what would be executed without running"
            echo "  -v, --verbose            Show detailed output"
            echo "  --help                   Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD, SQL_DIR"
            echo ""
            echo "Examples:"
            echo "  $0"
            echo "  $0 -h localhost -d mydb -U myuser"
            echo "  $0 --sql-dir ./migrations --dry-run"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Check if sql directory exists
if [ ! -d "$SQL_DIR" ]; then
    echo -e "${RED}Error: SQL directory '$SQL_DIR' does not exist${NC}"
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

# Find all SQL files with numbered prefixes and sort them
echo -e "${BLUE}Scanning SQL files in '$SQL_DIR'...${NC}"
SQL_FILES=$(find "$SQL_DIR" -maxdepth 1 -type f -name "[0-9][0-9][0-9]_*.sql" | sort)

# Check for duplicate sequence numbers
DUPLICATES=$(echo "$SQL_FILES" | sed 's|.*/||' | sed 's/^\([0-9][0-9][0-9]\)_.*/\1/' | sort | uniq -d)
if [ -n "$DUPLICATES" ]; then
    echo -e "${YELLOW}Warning: Duplicate sequence numbers found:${NC}"
    for dup in $DUPLICATES; do
        echo -e "  ${YELLOW}$dup${NC}: $(echo "$SQL_FILES" | grep "/${dup}_")"
    done
    echo ""
fi

if [ -z "$SQL_FILES" ]; then
    echo -e "${YELLOW}Warning: No SQL files found matching pattern '###_*.sql' in '$SQL_DIR'${NC}"
    exit 1
fi

# Count files
FILE_COUNT=$(echo "$SQL_FILES" | wc -l | tr -d ' ')
echo -e "${GREEN}Found $FILE_COUNT SQL file(s) to execute${NC}"
echo ""

if [ "$DRY_RUN" = true ]; then
    echo -e "${YELLOW}DRY RUN MODE - No files will be executed${NC}"
    echo ""
fi

# Track execution
SUCCESS_COUNT=0
FAILED_COUNT=0
FAILED_FILES=()

# Execute each SQL file in sequence
echo -e "${BLUE}Executing SQL files in sequence...${NC}"
echo ""

while IFS= read -r sql_file; do
    # Extract filename and number
    filename=$(basename "$sql_file")
    number=$(echo "$filename" | grep -oE '^[0-9]+' | head -1)
    
    echo -e "${BLUE}[$number]${NC} Executing: ${YELLOW}$filename${NC}"
    
    if [ "$VERBOSE" = true ] || [ "$DRY_RUN" = true ]; then
        echo "   File: $sql_file"
    fi
    
    if [ "$DRY_RUN" = true ]; then
        echo -e "   ${YELLOW}[DRY RUN] Would execute: psql $CONN_STRING -f \"$sql_file\"${NC}"
        echo ""
        continue
    fi
    
    # Execute SQL file
    if psql "$CONN_STRING" -f "$sql_file" > /dev/null 2>&1; then
        echo -e "   ${GREEN}✓ Success${NC}"
        ((SUCCESS_COUNT++))
    else
        echo -e "   ${RED}✗ Failed${NC}"
        ((FAILED_COUNT++))
        FAILED_FILES+=("$filename")
        
        # Show error details
        if [ "$VERBOSE" = true ]; then
            echo "   Error details:"
            psql "$CONN_STRING" -f "$sql_file" 2>&1 | sed 's/^/   /'
        fi
        
        # Ask if should continue
        read -p "   Continue with next file? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${RED}Execution stopped by user${NC}"
            break
        fi
    fi
    echo ""
done <<< "$SQL_FILES"

# Summary
echo "=========================================="
echo -e "${BLUE}Execution Summary${NC}"
echo "=========================================="
echo -e "Total files: $FILE_COUNT"
echo -e "${GREEN}Successful: $SUCCESS_COUNT${NC}"
if [ $FAILED_COUNT -gt 0 ]; then
    echo -e "${RED}Failed: $FAILED_COUNT${NC}"
    echo ""
    echo "Failed files:"
    for file in "${FAILED_FILES[@]}"; do
        echo -e "  ${RED}✗${NC} $file"
    done
    exit 1
else
    echo -e "${GREEN}All migrations executed successfully!${NC}"
    exit 0
fi
