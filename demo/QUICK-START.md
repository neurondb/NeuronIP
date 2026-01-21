# Quick Start: Loading Demo Data

## Problem
The Warehouse page (and other pages) show "No schemas available" or empty data because the demo data hasn't been loaded into the database yet.

## Solution

### Step 1: Load Demo Data

You need to load the demo data SQL file into your PostgreSQL database. Since your database is on a different machine, use one of these methods:

#### Option A: Using the Script (Recommended)
```bash
# Set your database connection details
export DB_HOST=your_database_host
export DB_PORT=5432
export DB_USER=your_database_user
export DB_PASSWORD=your_database_password
export DB_NAME=neuronip

# Run the script
./scripts/load-demo-data.sh
```

#### Option B: Using psql Directly
```bash
psql -h YOUR_DB_HOST -U YOUR_DB_USER -d neuronip -f demo/demo-data.sql
```

#### Option C: With Connection String
```bash
psql "postgresql://USER:PASSWORD@HOST:PORT/neuronip" -f demo/demo-data.sql
```

### Step 2: Verify Data Was Loaded

After loading, verify the data:

```sql
-- Check warehouse schemas
SELECT COUNT(*) FROM neuronip.warehouse_schemas;
-- Should return 3

-- Check users
SELECT COUNT(*) FROM neuronip.users;
-- Should return 8

-- Check warehouse queries
SELECT COUNT(*) FROM neuronip.warehouse_queries;
-- Should return 10

-- Check workflows
SELECT COUNT(*) FROM neuronip.workflows;
-- Should return 5
```

### Step 3: Restart Backend API

After loading demo data, restart your backend API so it can fetch the new data:

```bash
# If running in Docker
docker restart neuronip-api

# If running directly
# Stop and restart your Go server
```

### Step 4: Refresh Frontend

Refresh your browser to see the populated data.

## What Gets Populated

After loading demo data, you'll see:

### Warehouse Page
- ✅ 3 database schemas (analytics, customer_data, marketing)
- ✅ 10 warehouse queries with results
- ✅ Query history
- ✅ Schema explorer with tables

### Dashboard
- ✅ Activity metrics (searches, queries, workflows, compliance checks)
- ✅ Recent activity
- ✅ System status

### Semantic Search
- ✅ 32 search history entries
- ✅ 5 saved searches

### Workflows
- ✅ 5 workflow definitions
- ✅ 10 workflow executions

### Agent Hub
- ✅ 5 AI agents
- ✅ Performance metrics

### Support
- ✅ 10 support tickets
- ✅ Support conversations

### Compliance
- ✅ 720 audit log entries
- ✅ 3 compliance policies

### And More...
- Users, metrics, data sources, ingestion jobs, knowledge documents, etc.

## Troubleshooting

### "Cannot connect to database"
- Check your database connection details
- Verify the database is accessible from your machine
- Check firewall/network settings

### "No tables found in neuronip schema"
- Run migrations first: `./scripts/run-migrations.sh`
- Or apply the schema: `psql -d neuronip -f neuronip.sql`

### "Relation does not exist"
- Make sure all migrations have been run
- Check that the `neuronip` schema exists

### Data Still Not Showing
1. Check backend API logs for errors
2. Verify database connection in backend config
3. Check browser console for API errors
4. Ensure backend can reach the database

## Next Steps

Once demo data is loaded:
1. All pages will show realistic data
2. You can test all features with sample data
3. The application will look production-ready

For more details, see `demo/README.md`.
