#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="/opt/chatix/bin"
SERVICE_NAME="chatix.service"

echo "Сборка бинарника..."
cd "$ROOT/server"
CGO_ENABLED=1 GOOS=linux go build -o "$BIN_DIR/chatix-server" ./cmd/server
CGO_ENABLED=1 GOOS=linux go build -o "$BIN_DIR/chatix-cli" ./cmd/cli

echo "Применение миграций..."
# Используем goose (нужно будет установить: go install github.com/pressly/goose/v3/cmd/goose@latest)
# Пока просто SQL через psql для простоты, либо добавим goose в install-server.sh
sudo -u chatix psql -d chatix -f "$ROOT/server/migrations/001_init.sql" || echo "Миграции уже применены или ошибка."

echo "Перезапуск службы..."
sudo systemctl restart "$SERVICE_NAME"
sudo systemctl status "$SERVICE_NAME" --no-pager