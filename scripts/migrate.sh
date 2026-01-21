#!/bin/bash
# Migration Helper Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_DIR="${SCRIPT_DIR}/../api"

usage() {
  echo "Usage: $0 [COMMAND] [OPTIONS]"
  echo ""
  echo "Commands:"
  echo "  up        Apply all pending migrations"
  echo "  down      Roll back migrations (use -steps N for N steps)"
  echo "  status    Show migration status"
  echo "  create    Create a new migration (use -name NAME)"
  echo ""
  echo "Examples:"
  echo "  $0 up                    # Apply all migrations"
  echo "  $0 down -steps 1         # Roll back 1 migration"
  echo "  $0 status                # Show status"
  echo "  $0 create -name 'add_users_table'  # Create migration"
}

cd "$API_DIR"

case "${1:-}" in
  up)
    go run cmd/migrate/main.go -command up
    ;;
  down)
    STEPS="${2:-0}"
    go run cmd/migrate/main.go -command down -steps "$STEPS"
    ;;
  status)
    go run cmd/migrate/main.go -command status
    ;;
  create)
    if [ -z "$2" ]; then
      echo "Error: Migration name required"
      usage
      exit 1
    fi
    go run cmd/migrate/main.go -command create -name "$2"
    ;;
  *)
    usage
    exit 1
    ;;
esac
