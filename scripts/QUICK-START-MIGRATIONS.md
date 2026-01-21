# Quick Start: Running SQL Migrations

## Setup Complete ✅

- ✅ Created `sql/` directory with all 33 migration files
- ✅ Created `scripts/run-migrations.sh` script
- ✅ Files are numbered and will run in sequence

## Quick Usage

### 1. Basic Run (with defaults)
```bash
./scripts/run-migrations.sh
```

### 2. With Database Connection Parameters
```bash
./scripts/run-migrations.sh \
  -h your-db-host \
  -p 5432 \
  -d neuronip \
  -U your-username \
  -W
```

### 3. Using Environment Variables
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=neuronip
export DB_USER=postgres
export DB_PASSWORD=your_password

./scripts/run-migrations.sh
```

### 4. Dry Run (see what would execute)
```bash
./scripts/run-migrations.sh --dry-run
```

### 5. Verbose Output
```bash
./scripts/run-migrations.sh --verbose
```

## What It Does

1. **Scans** the `sql/` directory for files matching `###_*.sql` pattern
2. **Sorts** them by sequence number (001, 002, 003, etc.)
3. **Tests** database connection
4. **Executes** each file in order
5. **Reports** success/failure for each file
6. **Stops** on error (with option to continue)

## Example Output

```
Testing database connection...
✓ Database connection successful

Scanning SQL files in './sql'...
Found 33 SQL file(s) to execute

Executing SQL files in sequence...

[001] Executing: 001_users_schema.sql
   ✓ Success

[002] Executing: 002_ingestion_schema.sql
   ✓ Success

...

==========================================
Execution Summary
==========================================
Total files: 33
Successful: 33
All migrations executed successfully!
```

## File Structure

```
NeuronIP/
├── sql/                          # SQL migration files
│   ├── 001_users_schema.sql
│   ├── 002_ingestion_schema.sql
│   ├── 003_metadata_schema.sql
│   └── ... (30 more files)
└── scripts/
    └── run-migrations.sh         # Migration runner script
```

## Troubleshooting

**Connection Error?**
- Make sure PostgreSQL is running
- Check connection parameters
- Verify database exists: `CREATE DATABASE neuronip;`

**File Not Found?**
- Check that `sql/` directory exists
- Verify SQL files are in `sql/` directory
- Files must match pattern: `###_*.sql`

**Permission Denied?**
- Make script executable: `chmod +x scripts/run-migrations.sh`
- Check database user has CREATE privileges

## Next Steps

After running migrations:
1. Verify tables: `psql -d neuronip -c "\dt"`
2. Check schema: `psql -d neuronip -c "\dn"`
3. Start the API: The backend will now be able to connect to the database
