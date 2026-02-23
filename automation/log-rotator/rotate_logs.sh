#!/bin/bash
# ============================================================================
# Log Rotation Script
# Rotates, compresses, and cleans up application log files
# ============================================================================

set -euo pipefail

# ── Configuration ───────────────────────────────────────────────
LOG_DIR="${LOG_DIR:-/var/log/task-manager}"
ARCHIVE_DIR="${LOG_DIR}/archive"
MAX_SIZE_MB="${MAX_SIZE_MB:-50}"          # Rotate when log exceeds this size
RETENTION_DAYS="${RETENTION_DAYS:-30}"     # Delete archives older than this
COMPRESS="${COMPRESS:-true}"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] [LOG-ROTATOR] $1"
}

# ── Pre-flight ──────────────────────────────────────────────────
mkdir -p "$ARCHIVE_DIR"

# ── Rotate Logs ─────────────────────────────────────────────────
rotate_file() {
    local file="$1"
    local basename=$(basename "$file")
    local rotated="${ARCHIVE_DIR}/${basename}.${TIMESTAMP}"

    # Check file size
    local size_mb=$(du -m "$file" 2>/dev/null | cut -f1)
    if [ "${size_mb:-0}" -lt "$MAX_SIZE_MB" ]; then
        log "SKIP: $basename (${size_mb}MB < ${MAX_SIZE_MB}MB threshold)"
        return
    fi

    log "Rotating: $basename (${size_mb}MB)"

    # Copy and truncate (zero-downtime rotation)
    cp "$file" "$rotated"
    truncate -s 0 "$file"

    # Compress if enabled
    if [ "$COMPRESS" = "true" ]; then
        gzip "$rotated"
        log "Compressed: ${rotated}.gz"
    fi
}

# ── Process All Log Files ──────────────────────────────────────
log "Starting log rotation in: $LOG_DIR"

file_count=0
for logfile in "$LOG_DIR"/*.log; do
    [ -f "$logfile" ] || continue
    rotate_file "$logfile"
    file_count=$((file_count + 1))
done

log "Processed $file_count log file(s)"

# ── Cleanup Old Archives ──────────────────────────────────────
log "Cleaning archives older than ${RETENTION_DAYS} days..."
deleted=$(find "$ARCHIVE_DIR" -name "*.gz" -mtime +"$RETENTION_DAYS" -delete -print | wc -l)
log "Deleted $deleted old archive(s)"

# ── Summary ────────────────────────────────────────────────────
archive_count=$(find "$ARCHIVE_DIR" -name "*.gz" 2>/dev/null | wc -l)
archive_size=$(du -sh "$ARCHIVE_DIR" 2>/dev/null | cut -f1)
log "✅ Rotation complete. Archives: $archive_count files ($archive_size)"
