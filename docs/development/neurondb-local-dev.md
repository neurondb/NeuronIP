# Running NeuronIP with a Local NeuronDB Build

This guide describes how to run NeuronIP against a Postgres instance that has the NeuronDB extension built and installed from the local `neurondb` repository (e.g. `../neurondb` or a path you specify). This lets you use the latest NeuronDB features (hybrid fusion, reranking, drift, sparse vectors, etc.) during development.

## Prerequisites

- Docker and Docker Compose (for the recommended flow)
- Or: a local Postgres 16+ with the NeuronDB extension installed from source
- NeuronIP repo and (optionally) the NeuronDB repo at a path you control

## Option 1: Docker – NeuronDB + NeuronIP on the same network

1. **Create the external network** (if not already present):

   ```bash
   docker network create neurondb-network
   ```

2. **Start NeuronDB from the NeuronDB repo** (builds and runs Postgres with the NeuronDB extension):

   ```bash
   cd /path/to/neurondb   # e.g. /Users/ibrarahmed/pgelephant/pge/neurondb
   docker compose up -d neurondb
   ```

   NeuronDB’s Postgres listens on port `5433` by default (mapped from container 5432).

3. **Create the `neuronip` database and apply schema** (one-time):

   Either use NeuronIP’s docker-compose init (which expects the same network), or run manually:

   ```bash
   # From NeuronIP repo
   export PGHOST=localhost PGPORT=5433 PGUSER=neurondb PGPASSWORD=neurondb PGDATABASE=postgres
   psql -c "CREATE DATABASE neuronip;"
   psql -d neuronip -f neuronip.sql
   ```

4. **Point NeuronIP at NeuronDB**:

   Set env (or `.env`) so NeuronIP uses the same DB for app and NeuronDB (recommended):

   ```bash
   export DB_HOST=localhost
   export DB_PORT=5433
   export DB_USER=neurondb
   export DB_PASSWORD=neurondb
   export DB_NAME=neuronip
   # NEURONDB_* default to DB_* when unset, so one DB is used
   ```

5. **Run NeuronIP API and UI**:

   ```bash
   cd /path/to/NeuronIP
   # If using NeuronIP docker-compose (expects neurondb-network and external neurondb):
   docker compose up -d neuronip-server neuronip-ui
   ```

   Or run the API binary locally (no Docker):

   ```bash
   cd api && go run ./cmd/server
   ```

   The API will log `NeuronDB extension ready` with a version if the extension is installed.

## Option 2: NeuronIP Makefile target (optional)

If you set `NEURONDB_REPO_PATH` to the NeuronDB repo root, you can add a Make target that starts NeuronDB and then runs NeuronIP. Example (add to NeuronIP’s `Makefile` if you have one):

```makefile
NEURONDB_REPO_PATH ?= ../neurondb

neurondb-network:
	docker network create neurondb-network 2>/dev/null || true

neurondb-up: neurondb-network
	cd $(NEURONDB_REPO_PATH) && docker compose up -d neurondb

neurondb-down:
	cd $(NEURONDB_REPO_PATH) && docker compose stop neurondb

run-with-neurondb: neurondb-up
	@echo "Waiting for NeuronDB Postgres..."
	@sleep 5
	$(MAKE) run-api  # or: go run ./api/cmd/server
```

Then:

```bash
export NEURONDB_REPO_PATH=/Users/ibrarahmed/pgelephant/pge/neurondb
make run-with-neurondb
```

## Option 3: Local Postgres with NeuronDB extension from source

1. Build and install the NeuronDB extension from the NeuronDB repo (see NeuronDB’s `Makefile` / `build.sh` and docs).
2. Create a database and run NeuronDB’s SQL migration (e.g. `neurondb--*.sql` or the repo’s install instructions).
3. Run `neuronip.sql` on that database to create the NeuronIP schema (and `CREATE EXTENSION neurondb` if not already present).
4. Set `DB_*` and optionally `NEURONDB_*` to that instance and run the NeuronIP API as above.

## Verifying NeuronDB capabilities

- On startup, the API logs either `NeuronDB extension ready` with a version or a warning if the extension is missing.
- NeuronIP uses capability checks (e.g. `HasFunction("neurondb", "hybrid_search_fusion")`) to enable hybrid fusion, reranking, and drift when the extension provides them; otherwise it falls back to built-in SQL or skips those features.

## Troubleshooting

- **Connection refused**: Ensure NeuronDB’s Postgres is listening on the host/port you set in `DB_HOST`/`DB_PORT` (e.g. `localhost:5433` for Docker).
- **Extension not found**: Run `CREATE EXTENSION neurondb;` in the same database where you applied `neuronip.sql`.
- **Function does not exist**: Your NeuronDB build may not include that function (e.g. older version). NeuronIP will fall back when a capability is missing; see `docs/neurondb-capability-audit.md` for which functions are used.
