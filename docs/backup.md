# Резервное копирование и восстановление

## Автоматические бэкапы

Скрипт `deploy/backup.sh` запускается ежедневно в 02:00 через systemd-таймер `chatix-backup.timer`.

Создаёт в `/var/backups/chatix/`:
- `db_YYYY-MM-DD_HHMMSS.sql.gz` — дамп PostgreSQL
- `files_YYYY-MM-DD_HHMMSS.tar.gz` — архив каталога `/var/lib/chatix/files`

Хранятся 7 дней, старые удаляются автоматически.

## Ручной бэкап

```bash
/opt/chatix/deploy/backup.sh