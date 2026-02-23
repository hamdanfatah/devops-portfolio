# Database Backup Scheduler

Automated PostgreSQL backup script with GCS upload, retention policies, and alerting.

## Features

- **Automated backups** with `pg_dump` + gzip compression
- **Cloud storage upload** to Google Cloud Storage
- **Retention policies** — configurable cleanup of old backups
- **Slack notifications** for success/failure alerts
- **Verification** — ensures backup file integrity

## Usage

```bash
# Manual run
./backup.sh

# With custom config
POSTGRES_HOST=db.example.com \
BACKUP_DIR=/data/backups \
GCS_BUCKET=gs://my-backups \
RETENTION_DAYS=14 \
./backup.sh
```

## Scheduled via Cron

```bash
# Add to crontab
crontab -e

# Daily at 2 AM
0 2 * * * /opt/automation/backup-scheduler/backup.sh
```

## Environment Variables

| Variable            | Default      | Description              |
| ------------------- | ------------ | ------------------------ |
| `POSTGRES_HOST`     | localhost    | Database host            |
| `POSTGRES_PORT`     | 5432         | Database port            |
| `POSTGRES_USER`     | hamfa        | Database user            |
| `POSTGRES_PASSWORD` | —            | Database password        |
| `POSTGRES_DB`       | taskmanager  | Database name            |
| `BACKUP_DIR`        | /tmp/backups | Local backup directory   |
| `RETENTION_DAYS`    | 7            | Keep backups for N days  |
| `GCS_BUCKET`        | —            | GCS bucket URI           |
| `SLACK_WEBHOOK_URL` | —            | Slack webhook for alerts |
