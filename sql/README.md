# SQL Migrations Directory

This directory contains all SQL migration files for the NeuronIP database schema.

## File Naming Convention

All SQL files follow the pattern: `###_description.sql`

- `###` - Three-digit sequence number (001, 002, 003, etc.)
- `description` - Descriptive name of the migration
- `.sql` - SQL file extension

Examples:
- `001_users_schema.sql`
- `002_ingestion_schema.sql`
- `003_metadata_schema.sql`

## Running Migrations

### Using the Migration Script

The `scripts/run-migrations.sh` script automatically runs all SQL files in sequence:

```bash
# Basic usage (uses defaults)
./scripts/run-migrations.sh

# With custom database connection
./scripts/run-migrations.sh \
  -h localhost \
  -p 5432 \
  -d neuronip \
  -U postgres \
  -W

# Using environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=neuronip
export DB_USER=postgres
export DB_PASSWORD=your_password
./scripts/run-migrations.sh

# Dry run (see what would be executed)
./scripts/run-migrations.sh --dry-run

# Verbose output
./scripts/run-migrations.sh --verbose
```

### Manual Execution

You can also run files manually in order:

```bash
psql -h localhost -U postgres -d neuronip -f 001_users_schema.sql
psql -h localhost -U postgres -d neuronip -f 002_ingestion_schema.sql
# ... and so on
```

### Using psql with Connection String

```bash
PGPASSWORD=your_password psql \
  -h localhost \
  -p 5432 \
  -U postgres \
  -d neuronip \
  -f 001_users_schema.sql
```

## Migration Order

Migrations are executed in numerical order:
1. `001_users_schema.sql` - User management and authentication
2. `002_ingestion_schema.sql` - Data ingestion tables
3. `003_metadata_schema.sql` - Metadata management
4. `004_governance_schema.sql` - Data governance
5. ... and so on

**Important**: Always run migrations in sequence. Do not skip files or run them out of order.

## Adding New Migrations

When adding a new migration:

1. Use the next available number (e.g., if last is `030_`, use `031_`)
2. Use descriptive names (e.g., `031_new_feature.sql`)
3. Place the file in this `sql/` directory
4. The migration script will automatically pick it up and run it in sequence

## Troubleshooting

### Connection Errors

If you get connection errors, check:
- PostgreSQL is running
- Database exists: `CREATE DATABASE neuronip;`
- User has proper permissions
- Connection parameters are correct

### Migration Failures

If a migration fails:
- Check the error message
- Verify the previous migrations completed successfully
- Review the SQL file for syntax errors
- Check database state: `\dt` in psql

### Duplicate Execution

The script will attempt to run all files. If migrations are idempotent (use `IF NOT EXISTS`), they can be run multiple times safely.

## File List

Current migration files:

```bash
ls -1 sql/*.sql | sort
```
