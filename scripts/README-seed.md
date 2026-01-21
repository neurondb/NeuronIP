# Demo Data Seeding

This directory contains scripts to seed demo data into NeuronIP.

## Quick Start

### Option 1: Seed via API (Recommended)

This approach seeds data through the API endpoints and works as long as the API server is running:

```bash
# Make sure API is running first
# Then run:
./scripts/seed-via-api.sh

# Or with custom API URL:
API_URL=http://localhost:8082/api/v1 ./scripts/seed-via-api.sh

# With API key (if required):
API_URL=http://localhost:8082/api/v1 API_KEY=your-key ./scripts/seed-via-api.sh
```

This creates:
- ✅ 3 demo users (demo@example.com, john@example.com, jane@example.com)
- ✅ API keys
- ✅ Support tickets
- ✅ Saved searches
- ✅ Warehouse schemas (from demo JSON files)

**Default credentials:**
- Email: `demo@example.com`
- Password: `demo123`

### Option 2: Seed via Database (Direct)

This approach seeds data directly into the database. Requires database connection:

```bash
# Set database environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=neuronip
export DB_PASSWORD=neuronip
export DB_NAME=neuronip

# Run migrations first
cd api
go run cmd/migrate/main.go -command up

# Then seed
go run cmd/seed/main.go -type demo -clear
```

## What Gets Seeded

### Users
- `demo@example.com` - Admin user
- `john@example.com` - Regular user
- `jane@example.com` - Regular user

### Support Tickets
- Sample tickets with conversations
- Customer support scenarios

### Knowledge Base
- Documentation articles
- FAQ entries
- Best practices guides

### Warehouse Schemas
- Sales data schema
- Sample tables (products, customers, orders, order_items)
- Relationships and metadata

### Saved Searches
- "Top Products by Revenue"
- "Active Customers"
- "Monthly Sales Summary"

### API Keys
- Demo API keys for testing

### Workflows
- Sample workflow definitions

### Metrics
- Pre-defined business metrics
- Revenue, customer count, AOV

## Troubleshooting

### API not accessible
Make sure the API server is running:
```bash
# Check API health
curl http://localhost:8082/health
```

### Database connection errors
For direct database seeding, ensure:
- Database is running
- Environment variables are set correctly
- Database user has proper permissions

### Authentication errors
If API requires authentication:
- Get an API key from the UI
- Use it with the `API_KEY` environment variable

## Demo Data Files

Demo data is loaded from:
- `examples/demos/support-tickets-demo.json`
- `examples/demos/knowledge-base-demo.json`
- `examples/demos/warehouse-sales-demo.json`

You can customize these files to add more demo data.
