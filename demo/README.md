# Demo Data for NeuronIP

This directory contains comprehensive demo data that populates the NeuronIP database with realistic information, making all pages show meaningful content.

## Files

- **`demo-data.sql`** - Complete demo dataset with realistic data for all features

## What Gets Populated

The demo data includes:

### Users & Profiles
- 8 demo users with different roles (admin, analyst, developer, support)
- User profiles with company, job titles, and locations
- Realistic user activity and login history

### Semantic Search
- 32 search history entries (last 30 days)
- 5 saved searches with different configurations
- Search queries covering various business topics

### Data Warehouse
- 3 warehouse schemas (analytics, customer_data, marketing)
- 10 warehouse queries with natural language questions
- Query results with chart configurations
- Query explanations for AI-generated insights

### Workflows
- 5 workflow definitions (Daily Sales Report, Customer Churn Analysis, etc.)
- 10 workflow executions with various statuses
- Workflow performance metrics

### Agents
- 5 AI agents (Sales Analytics, Customer Support, Data Quality, etc.)
- Agent performance metrics (last 7 days, hourly)
- Realistic performance data

### Support System
- 3 support agents
- 10 support tickets with various statuses
- Support conversations for ticket resolution

### Metrics & Analytics
- 6 business metrics (MRR, CAC, Churn Rate, AOV, etc.)
- 4 metric catalog entries
- Usage metrics (last 30 days, hourly)
- Cost tracking (last 3 months)

### Compliance & Governance
- 234 audit log entries (last 30 days)
- 3 compliance policies (GDPR, Rate Limiting, PII Filtering)
- System logs for observability

### Data Sources & Ingestion
- 5 data sources (PostgreSQL, MySQL, APIs)
- 168 ingestion jobs (last 7 days)
- Various job statuses and progress

### Knowledge Base
- 50 knowledge documents
- Various document types (articles, documentation, tutorials, FAQs)
- Realistic content for search and discovery

## Usage

### Prerequisites

1. All migrations must be run first:
   ```bash
   ./scripts/run-migrations.sh
   ```

2. Database must be accessible and configured

### Running Demo Data

**Option 1: Using psql directly**
```bash
psql -h localhost -U postgres -d neuronip -f demo/demo-data.sql
```

**Option 2: Using environment variables**
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=neuronip
export DB_USER=postgres
export PGPASSWORD=your_password

psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f demo/demo-data.sql
```

**Option 3: Using the migration script (if you add it)**
```bash
# The demo data can be run after migrations
./scripts/run-migrations.sh
psql -h localhost -U postgres -d neuronip -f demo/demo-data.sql
```

### Verification

After running the demo data, verify it was inserted:

```sql
-- Check user count
SELECT COUNT(*) FROM neuronip.users;

-- Check search history
SELECT COUNT(*) FROM neuronip.search_history;

-- Check warehouse queries
SELECT COUNT(*) FROM neuronip.warehouse_queries;

-- Check workflows
SELECT COUNT(*) FROM neuronip.workflows;

-- Check agents
SELECT COUNT(*) FROM neuronip.agents;
```

## Data Characteristics

### Realistic Data
- **User names**: Real-sounding names with professional emails
- **Company**: Acme Corporation (consistent across all users)
- **Timestamps**: Spread over last 30-90 days for realistic activity patterns
- **Metrics**: Realistic numeric values (revenue, counts, percentages)
- **Statuses**: Mix of completed, in-progress, and pending states

### Coverage
- **Dashboard**: Populated with metrics, activity, and system status
- **Semantic Search**: Search history and saved searches
- **Warehouse Q&A**: Queries, results, and explanations
- **Workflows**: Workflow definitions and execution history
- **Compliance**: Audit logs and policy enforcement
- **Support**: Tickets, conversations, and agent activity
- **Observability**: Usage metrics, costs, and system logs

## Resetting Demo Data

To reset and reload demo data:

```sql
-- WARNING: This will delete all data!
-- Only run this if you want to start fresh

TRUNCATE TABLE neuronip.audit_logs CASCADE;
TRUNCATE TABLE neuronip.usage_metrics CASCADE;
TRUNCATE TABLE neuronip.workflow_executions CASCADE;
TRUNCATE TABLE neuronip.warehouse_queries CASCADE;
TRUNCATE TABLE neuronip.search_history CASCADE;
-- ... (truncate other tables as needed)

-- Then re-run demo-data.sql
```

Or use a fresh database:

```bash
# Drop and recreate database
dropdb -h localhost -U postgres neuronip
createdb -h localhost -U postgres neuronip

# Run migrations
./scripts/run-migrations.sh

# Load demo data
psql -h localhost -U postgres -d neuronip -f demo/demo-data.sql
```

## Notes

- Demo data uses fixed UUIDs for relationships to ensure referential integrity
- Timestamps are spread over realistic time periods (last 30-90 days)
- Some data uses `ON CONFLICT DO NOTHING` to allow safe re-runs
- Vector embeddings are not included (would require actual embedding generation)
- The data is designed to make all UI pages show meaningful content

## Troubleshooting

**Error: relation does not exist**
- Make sure all migrations have been run first
- Check that the schema `neuronip` exists

**Error: duplicate key violation**
- Some tables use `ON CONFLICT DO NOTHING` to handle duplicates
- If you get errors, you may need to truncate tables first

**Error: foreign key violation**
- Make sure data is inserted in the correct order
- The script handles dependencies, but if you modify it, maintain the order

## Next Steps

After loading demo data:
1. Start the backend API
2. Start the frontend UI
3. Log in with any of the demo users
4. Explore all pages - they should now show realistic data!
