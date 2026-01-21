#!/bin/bash
# Quick script to load demo data - just update the connection details below

# ============================================================================
# UPDATE THESE WITH YOUR DATABASE CONNECTION DETAILS
# ============================================================================
DB_HOST="${DB_HOST:-your_database_host}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-your_database_user}"
DB_PASSWORD="${DB_PASSWORD:-your_database_password}"
DB_NAME="${DB_NAME:-neuronip}"

# ============================================================================
# NO NEED TO EDIT BELOW THIS LINE
# ============================================================================

export PGPASSWORD="$DB_PASSWORD"

echo "🚀 Loading NeuronIP Demo Data..."
echo ""
echo "Connection details:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo ""

# Test connection
echo "Testing connection..."
if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✅ Connection successful!"
    echo ""
    echo "Loading demo data..."
    
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$(dirname "$0")/demo-data.sql"; then
        echo ""
        echo "✅ Demo data loaded successfully!"
        echo ""
        echo "Verification:"
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
            SELECT 
                'Users' as table_name, COUNT(*)::text as count FROM neuronip.users
            UNION ALL
            SELECT 'Warehouse Schemas', COUNT(*)::text FROM neuronip.warehouse_schemas
            UNION ALL
            SELECT 'Warehouse Queries', COUNT(*)::text FROM neuronip.warehouse_queries
            UNION ALL
            SELECT 'Workflows', COUNT(*)::text FROM neuronip.workflows
            UNION ALL
            SELECT 'Agents', COUNT(*)::text FROM neuronip.agents
            UNION ALL
            SELECT 'Support Tickets', COUNT(*)::text FROM neuronip.support_tickets
            UNION ALL
            SELECT 'Audit Logs', COUNT(*)::text FROM neuronip.audit_logs;
        "
        echo ""
        echo "🎉 Done! Refresh your frontend to see the data."
    else
        echo ""
        echo "❌ Error loading demo data. Check the error messages above."
        exit 1
    fi
else
    echo "❌ Cannot connect to database!"
    echo ""
    echo "Please update the connection details at the top of this script:"
    echo "  - DB_HOST"
    echo "  - DB_PORT"
    echo "  - DB_USER"
    echo "  - DB_PASSWORD"
    echo "  - DB_NAME"
    echo ""
    echo "Or set them as environment variables:"
    echo "  export DB_HOST=your_host"
    echo "  export DB_USER=your_user"
    echo "  export DB_PASSWORD=your_password"
    echo "  ./demo/load-now.sh"
    exit 1
fi
