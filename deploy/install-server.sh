#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

if [[ $EUID -ne 0 ]]; then
  echo "Запустите скрипт с root или через sudo." >&2
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

apt-get update -y

apt-get install -y \
  ca-certificates \
  curl \
  gnupg \
  lsb-release \
  sudo \
  git \
  build-essential \
  pkg-config \
  jq \
  postgresql \
  postgresql-contrib \
  libvips-dev \
  libvips-tools

# Valkey: пробуем пакет из APT.
VALKEY_SERVICE=""
if apt-cache show valkey-server >/dev/null 2>&1; then
  apt-get install -y valkey-server
  VALKEY_SERVICE="valkey-server"
elif apt-cache show valkey >/dev/null 2>&1; then
  apt-get install -y valkey
  VALKEY_SERVICE="valkey"
else
  echo "ВНИМАНИЕ: в APT не найден valkey-server или valkey." >&2
  echo "Нужно будет отдельно решить, откуда ставить Valkey." >&2
fi

# Caddy: официальный репозиторий.
curl -fsSL 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor --yes -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/caddy-stable-archive-keyring.gpg] https://dl.cloudsmith.io/public/caddy/stable/deb/debian any-version main" > /etc/apt/sources.list.d/caddy-stable.list
apt-get update -y
apt-get install -y caddy

# Go: последняя стабильная версия.
GO_LATEST="$(curl -fsSL 'https://go.dev/VERSION?m=text' | head -n1)"
CURRENT_GO="$(/usr/local/go/bin/go version 2>/dev/null | awk '{print $3}' || true)"

if [[ "$GO_LATEST" != "$CURRENT_GO" ]]; then
  ARCH="$(dpkg --print-architecture)"
  case "$ARCH" in
    amd64)
      GO_ARCH="amd64"
      ;;
    arm64)
      GO_ARCH="arm64"
      ;;
    *)
      echo "Неподдерживаемая архитектура для Go: $ARCH" >&2
      exit 1
      ;;
  esac

  curl -fsSL "https://go.dev/dl/${GO_LATEST}.linux-${GO_ARCH}.tar.gz" -o /tmp/go.tar.gz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf /tmp/go.tar.gz
  rm -f /tmp/go.tar.gz
fi

ln -sf /usr/local/go/bin/go /usr/local/bin/go
ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt

# Node.js LTS через NodeSource.
curl -fsSL https://deb.nodesource.com/setup_lts.x | bash -
apt-get install -y nodejs

# Пользователь и каталоги.
id -u chatix >/dev/null 2>&1 || useradd --system --home-dir /opt/chatix --shell /usr/sbin/nologin chatix

install -d -o chatix -g chatix \
  /etc/chatix \
  /var/lib/chatix/files \
  /var/lib/chatix/updates \
  /var/lib/chatix/tmp \
  /var/backups/chatix \
  /opt/chatix/bin

# PostgreSQL.
systemctl enable --now postgresql

if ! sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='chatix'" | grep -q 1; then
  sudo -u postgres psql -c "CREATE ROLE chatix LOGIN"
fi

if ! sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='chatix'" | grep -q 1; then
  sudo -u postgres createdb -O chatix chatix
fi

sudo -u postgres psql -d chatix -c "CREATE EXTENSION IF NOT EXISTS pg_trgm; CREATE EXTENSION IF NOT EXISTS unaccent;"

# Valkey service.
if [[ -n "$VALKEY_SERVICE" ]]; then
  systemctl enable --now "$VALKEY_SERVICE"
fi

# Caddy пока только включаем, конфигурацию добавим позже.
systemctl enable caddy || true

# Конфиг по умолчанию.
if [[ ! -f /etc/chatix/config.yaml ]]; then
  if [[ -f "$ROOT/deploy/config.example.yaml" ]]; then
    install -m 640 -o root -g chatix "$ROOT/deploy/config.example.yaml" /etc/chatix/config.yaml
  fi
fi

# Сбор фактических версий.
bash "$ROOT/deploy/collect-latest.sh"

echo
echo "Установка окружения завершена."
echo "Проверьте: /opt/chatix/deploy/versions.env"