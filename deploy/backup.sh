#!/usr/bin/env bash
set -euo pipefail

BACKUP_DIR="/var/backups/chatix"
DATE=$(date +%Y-%m-%d_%H%M%S)
DB_BACKUP="$BACKUP_DIR/db_$DATE.sql.gz"
FILES_BACKUP="$BACKUP_DIR/files_$DATE.tar.gz"

mkdir -p "$BACKUP_DIR"

echo "Бэкап базы данных..."
sudo -u postgres pg_dump chatix | gzip > "$DB_BACKUP"

echo "Бэкап файлов..."
tar -czf "$FILES_BACKUP" -C /var/lib/chatix files

echo "Ротация (храним 7 дней)..."
find "$BACKUP_DIR" -type f -name "*.gz" -mtime +7 -delete

echo "Готово: $DB_BACKUP, $FILES_BACKUP"