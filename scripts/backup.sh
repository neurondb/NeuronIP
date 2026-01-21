#!/bin/bash
# Backup script for NeuronIP database and files

set -e

BACKUP_DIR="${BACKUP_DIR:-/var/backups/neuronip}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-neuronip}"
DB_USER="${DB_USER:-postgres}"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/full_${TIMESTAMP}.sql"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Perform database backup
echo "Starting database backup..."
PGPASSWORD="$DB_PASSWORD" pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -F c -f "${BACKUP_FILE}.dump"

# Compress backup
echo "Compressing backup..."
gzip "${BACKUP_FILE}.dump"

# Backup configuration files if they exist
if [ -d "/etc/neuronip" ]; then
    echo "Backing up configuration files..."
    tar -czf "$BACKUP_DIR/config_${TIMESTAMP}.tar.gz" /etc/neuronip
fi

# Cleanup old backups
echo "Cleaning up backups older than $RETENTION_DAYS days..."
find "$BACKUP_DIR" -name "*.sql.gz" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup completed: ${BACKUP_FILE}.dump.gz"
