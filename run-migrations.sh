#!/bin/bash
# Run All Database Schema Scripts (Version 1.0.0)
# This script runs all SQL schema files in the correct order to set up the database
#
# Schema Version: 1.0.0
# =====================
# These SQL files represent the complete database schema for NeuronIP version 1.0.0.
# They are not migrations in the traditional sense - they are the initial schema
# definition that sets up all tables, indexes, functions, and other database objects
# needed for the application to run.
#
# The files are numbered (001_, 002_, etc.) to ensure they run in the correct order,
# as some tables depend on others being created first.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Schema version
SCHEMA_VERSION="1.0.0"

# Default database connection settings
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-neuronip}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-}"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCHEMA_DIR="${SCRIPT_DIR}/api/migrations"

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

# Parse command line arguments
show_help() {
    cat << EOF
Run All Database Schema Scripts (Version ${SCHEMA_VERSION})

Usage: $0 [OPTIONS]

This script runs all SQL schema files from api/migrations/ in sequential order
to set up the NeuronIP database schema version ${SCHEMA_VERSION}.

Options:
  -h, --host HOST          Database host (default: localhost)
  -p, --port PORT          Database port (default: 5432)
  -d, --database NAME      Database name (default: neuronip)
  -U, --user USER          Database user (default: postgres)
  -W, --password           Prompt for password
  --schema-dir DIR         Schema directory (default: api/migrations)
  --dry-run               Show what would be executed without running
  -v, --verbose           Show detailed output
  --help                  Show this help message

Environment Variables:
  DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD

Examples:
  $0                                    # Run all schema scripts with defaults
  $0 -h localhost -d mydb -U myuser    # Run with custom connection
  $0 --dry-run                          # Preview what would run
  $0 -v                                 # Verbose output

EOF
}

DRY_RUN=false
VERBOSE=false

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
            shift
            ;;
        --schema-dir)
            SCHEMA_DIR="$2"
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
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Check if schema directory exists
if [ ! -d "$SCHEMA_DIR" ]; then
    log_error "Schema directory not found: $SCHEMA_DIR"
    exit 1
fi

# Check if psql is available
if ! command -v psql &> /dev/null; then
    log_error "psql command not found. Please install PostgreSQL client tools."
    exit 1
fi

# Export password for psql
export PGPASSWORD

# Build connection string
CONN_STRING="host=$DB_HOST port=$DB_PORT dbname=$DB_NAME user=$DB_USER"

# Test database connection
log_step "Testing Database Connection"
log_info "Connecting to: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"

if ! psql "$CONN_STRING" -c "SELECT 1;" > /dev/null 2>&1; then
    log_error "Cannot connect to database"
    echo "  Host: $DB_HOST"
    echo "  Port: $DB_PORT"
    echo "  Database: $DB_NAME"
    echo "  User: $DB_USER"
    echo ""
    echo "Please check:"
    echo "  1. Database is running"
    echo "  2. Connection settings are correct"
    echo "  3. User has proper permissions"
    exit 1
fi
log_success "Database connection successful"
echo ""

# Find all SQL files and sort them by number
log_step "Scanning Schema Files (Version ${SCHEMA_VERSION})"
SQL_FILES=$(find "$SCHEMA_DIR" -maxdepth 1 -type f -name "*.sql" | sort)

if [ -z "$SQL_FILES" ]; then
    log_error "No SQL files found in $SCHEMA_DIR"
    exit 1
fi

# Count files
FILE_COUNT=$(echo "$SQL_FILES" | wc -l | tr -d ' ')
log_success "Found $FILE_COUNT schema file(s)"
echo ""

# Check for duplicate sequence numbers (warn but continue)
DUPLICATES=$(echo "$SQL_FILES" | sed 's|.*/||' | sed 's/^\([0-9][0-9][0-9]\)_.*/\1/' | sort | uniq -d)
if [ -n "$DUPLICATES" ]; then
    log_warning "Duplicate sequence numbers found:"
    for dup in $DUPLICATES; do
        echo "$SQL_FILES" | grep "/${dup}_" | sed 's|.*/|  |'
    done
    echo ""
fi

if [ "$DRY_RUN" = true ]; then
    log_warning "DRY RUN MODE - No files will be executed"
    echo ""
fi

# Track execution
SUCCESS_COUNT=0
FAILED_COUNT=0
FAILED_FILES=()

# Execute each SQL file in sequence
log_step "Executing Schema Scripts (Version ${SCHEMA_VERSION})"

while IFS= read -r sql_file; do
    filename=$(basename "$sql_file")
    number=$(echo "$filename" | grep -oE '^[0-9]+' | head -1 || echo "???")
    
    echo -e "${BLUE}[$number]${NC} ${YELLOW}$filename${NC}"
    
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
        else
            # Show last few lines of error
            echo "   Error:"
            psql "$CONN_STRING" -f "$sql_file" 2>&1 | tail -3 | sed 's/^/   /'
        fi
        
        # Ask if should continue
        read -p "   Continue with next file? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_error "Execution stopped by user"
            break
        fi
    fi
    echo ""
done <<< "$SQL_FILES"

# Summary
echo ""
log_step "Execution Summary"
echo "Total files: $FILE_COUNT"
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
    log_success "All schema scripts executed successfully!"
    echo ""
    log_info "Database schema version ${SCHEMA_VERSION} has been set up."
    exit 0
fi
