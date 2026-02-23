#!/bin/bash
# ============================================================================
# PostgreSQL Automated Backup Script
# Backs up database, compresses with gzip, and uploads to Google Cloud Storage
# ============================================================================

set -euo pipefail

# ── Configuration ───────────────────────────────────────────────
BACKUP_DIR="${BACKUP_DIR:-/tmp/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
GCS_BUCKET="${GCS_BUCKET:-gs://your-backup-bucket}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="taskmanager_${TIMESTAMP}.sql.gz"

# Database config
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_USER="${POSTGRES_USER:-hamfa}"
DB_NAME="${POSTGRES_DB:-taskmanager}"

# Logging
LOG_FILE="${BACKUP_DIR}/backup.log"

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

error_exit() {
    log "ERROR: $1"
    # Send alert (optional: configure webhook URL)
    if [ -n "${SLACK_WEBHOOK_URL:-}" ]; then
        curl -s -X POST "$SLACK_WEBHOOK_URL" \
            -H 'Content-Type: application/json' \
            -d "{\"text\":\"🚨 Backup Failed: $1\"}" > /dev/null
    fi
    exit 1
}

# ── Pre-flight Checks ──────────────────────────────────────────
mkdir -p "$BACKUP_DIR"

command -v pg_dump >/dev/null 2>&1 || error_exit "pg_dump not found"
command -v gzip >/dev/null 2>&1 || error_exit "gzip not found"

log "Starting backup of database '$DB_NAME'..."

# ── Perform Backup ─────────────────────────────────────────────
PGPASSWORD="${POSTGRES_PASSWORD}" pg_dump \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --format=custom \
    --compress=9 \
    --verbose \
    2>> "$LOG_FILE" | gzip > "${BACKUP_DIR}/${BACKUP_FILE}"

if [ $? -ne 0 ]; then
    error_exit "pg_dump failed"
fi

BACKUP_SIZE=$(du -h "${BACKUP_DIR}/${BACKUP_FILE}" | cut -f1)
log "Backup created: ${BACKUP_FILE} (${BACKUP_SIZE})"

# ── Upload to GCS ──────────────────────────────────────────────
if command -v gsutil >/dev/null 2>&1; then
    log "Uploading to ${GCS_BUCKET}..."
    gsutil cp "${BACKUP_DIR}/${BACKUP_FILE}" "${GCS_BUCKET}/backups/${BACKUP_FILE}"
    
    if [ $? -eq 0 ]; then
        log "Upload successful: ${GCS_BUCKET}/backups/${BACKUP_FILE}"
    else
        error_exit "Upload to GCS failed"
    fi
else
    log "WARNING: gsutil not found, skipping GCS upload"
fi

# ── Cleanup Old Backups ────────────────────────────────────────
log "Cleaning up backups older than ${RETENTION_DAYS} days..."
find "$BACKUP_DIR" -name "taskmanager_*.sql.gz" -mtime +"$RETENTION_DAYS" -delete
REMAINING=$(find "$BACKUP_DIR" -name "taskmanager_*.sql.gz" | wc -l)
log "Remaining local backups: ${REMAINING}"

# ── Verify Backup ──────────────────────────────────────────────
if [ -f "${BACKUP_DIR}/${BACKUP_FILE}" ] && [ -s "${BACKUP_DIR}/${BACKUP_FILE}" ]; then
    log "✅ Backup completed successfully!"
    
    # Send success notification
    if [ -n "${SLACK_WEBHOOK_URL:-}" ]; then
        curl -s -X POST "$SLACK_WEBHOOK_URL" \
            -H 'Content-Type: application/json' \
            -d "{\"text\":\"✅ Backup Successful: ${BACKUP_FILE} (${BACKUP_SIZE})\"}" > /dev/null
    fi
else
    error_exit "Backup file is empty or missing!"
fi
