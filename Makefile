# NeuronIP Makefile
# Optional targets for running with a local NeuronDB build (see docs/development/neurondb-local-dev.md).

NEURONDB_REPO_PATH ?= ../neurondb

.PHONY: neurondb-network neurondb-up neurondb-down run-api run-with-neurondb help

help:
	@echo "NeuronIP targets for local NeuronDB development:"
	@echo "  make neurondb-network   Create docker network neurondb-network"
	@echo "  make neurondb-up        Start NeuronDB Postgres from NEURONDB_REPO_PATH"
	@echo "  make neurondb-down     Stop NeuronDB container"
	@echo "  make run-api            Run API server (go run ./api/cmd/server)"
	@echo "  make run-with-neurondb  Start NeuronDB then run API (requires NEURONDB_REPO_PATH)"
	@echo ""
	@echo "Set NEURONDB_REPO_PATH to your NeuronDB repo (default: ../neurondb)"

neurondb-network:
	docker network create neurondb-network 2>/dev/null || true

neurondb-up: neurondb-network
	cd $(NEURONDB_REPO_PATH) && docker compose up -d neurondb

neurondb-down:
	cd $(NEURONDB_REPO_PATH) && docker compose stop neurondb

run-api:
	cd api && go run ./cmd/server

run-with-neurondb: neurondb-up
	@echo "Waiting for NeuronDB Postgres..."
	@sleep 5
	$(MAKE) run-api
