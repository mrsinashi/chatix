#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_ENV="$ROOT/deploy/versions.env"
OUT_MD="$ROOT/deploy/versions.md"

SUDO=""
if [[ $EUID -ne 0 ]]; then
  SUDO="sudo"
fi

$SUDO apt-get update -y >/dev/null
$SUDO apt-get install -y curl ca-certificates gnupg jq >/dev/null

apt_candidate() {
  local pkg="$1"
  apt-cache policy "$pkg" 2>/dev/null | awk '/Candidate:/ {print $2}' || true
}

npm_latest() {
  local pkg="$1"
  local encoded="${pkg//\//%2F}"
  (curl -fsSL "https://registry.npmjs.org/${encoded}/latest" | jq -r '.version // "unknown"') 2>/dev/null || echo "unknown"
}

github_latest() {
  local repo="$1"
  (curl -fsSL -H "Accept: application/vnd.github+json" "https://api.github.com/repos/${repo}/releases/latest" | jq -r '.tag_name // "unknown"') 2>/dev/null || echo "unknown"
}

crates_latest() {
  local crate="$1"
  (curl -fsSL -H "User-Agent: chatix-bootstrap" "https://crates.io/api/v1/crates/${crate}" | jq -r '.crate.max_stable_version // .crate.newest_version // "unknown"') 2>/dev/null || echo "unknown"
}

OS_VERSION="$(. /etc/os-release && echo "$PRETTY_NAME")"
KERNEL="$(uname -r)"
ARCH="$(dpkg --print-architecture)"

GO_LATEST="$(curl -fsSL 'https://go.dev/VERSION?m=text' 2>/dev/null | head -n1 || true)"
[[ -z "${GO_LATEST:-}" ]] && GO_LATEST="unknown"

NODE_LATEST="$(curl -fsSL https://nodejs.org/dist/index.json 2>/dev/null | jq -r 'map(select(.lts)) | .[0].version // empty' | sed 's/^v//' || true)"
[[ -z "${NODE_LATEST:-}" ]] && NODE_LATEST="unknown"

APT_POSTGRESQL="$(apt_candidate postgresql)"
APT_POSTGRESQL_CONTRIB="$(apt_candidate postgresql-contrib)"
APT_VALKEY_SERVER="$(apt_candidate valkey-server)"
APT_VALKEY="$(apt_candidate valkey)"
APT_LIBVIPS_DEV="$(apt_candidate libvips-dev)"
APT_LIBVIPS_TOOLS="$(apt_candidate libvips-tools)"
APT_CADDY="$(apt_candidate caddy)"

GH_CADDY="$(github_latest caddyserver/caddy)"
GH_VALKEY="$(github_latest valkey-io/valkey)"
GH_TAURI="$(github_latest tauri-apps/tauri)"

CRATE_TAURI="$(crates_latest tauri)"
CRATE_TAURI_BUILD="$(crates_latest tauri-build)"

NPM_REACT="$(npm_latest react)"
NPM_REACT_DOM="$(npm_latest react-dom)"
NPM_VITE="$(npm_latest vite)"
NPM_VITE_PLUGIN_REACT="$(npm_latest @vitejs/plugin-react)"
NPM_TYPESCRIPT="$(npm_latest typescript)"
NPM_TYPES_REACT="$(npm_latest @types/react)"
NPM_TYPES_REACT_DOM="$(npm_latest @types/react-dom)"
NPM_ZUSTAND="$(npm_latest zustand)"
NPM_TANSTACK_QUERY="$(npm_latest @tanstack/react-query)"
NPM_DEXIE="$(npm_latest dexie)"
NPM_MOTION="$(npm_latest motion)"
NPM_LUCIDE_REACT="$(npm_latest lucide-react)"
NPM_TAURI_API="$(npm_latest @tauri-apps/api)"
NPM_TAURI_CLI="$(npm_latest @tauri-apps/cli)"

COLLECTED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

cat > "$OUT_ENV" <<EOF
COLLECTED_AT="$COLLECTED_AT"
OS_VERSION="$OS_VERSION"
KERNEL="$KERNEL"
ARCH="$ARCH"

GO_LATEST="$GO_LATEST"
NODE_LATEST="$NODE_LATEST"

APT_POSTGRESQL="$APT_POSTGRESQL"
APT_POSTGRESQL_CONTRIB="$APT_POSTGRESQL_CONTRIB"
APT_VALKEY_SERVER="$APT_VALKEY_SERVER"
APT_VALKEY="$APT_VALKEY"
APT_LIBVIPS_DEV="$APT_LIBVIPS_DEV"
APT_LIBVIPS_TOOLS="$APT_LIBVIPS_TOOLS"
APT_CADDY="$APT_CADDY"

GH_CADDY="$GH_CADDY"
GH_VALKEY="$GH_VALKEY"
GH_TAURI="$GH_TAURI"

CRATE_TAURI="$CRATE_TAURI"
CRATE_TAURI_BUILD="$CRATE_TAURI_BUILD"

NPM_REACT="$NPM_REACT"
NPM_REACT_DOM="$NPM_REACT_DOM"
NPM_VITE="$NPM_VITE"
NPM_VITE_PLUGIN_REACT="$NPM_VITE_PLUGIN_REACT"
NPM_TYPESCRIPT="$NPM_TYPESCRIPT"
NPM_TYPES_REACT="$NPM_TYPES_REACT"
NPM_TYPES_REACT_DOM="$NPM_TYPES_REACT_DOM"
NPM_ZUSTAND="$NPM_ZUSTAND"
NPM_TANSTACK_QUERY="$NPM_TANSTACK_QUERY"
NPM_DEXIE="$NPM_DEXIE"
NPM_MOTION="$NPM_MOTION"
NPM_LUCIDE_REACT="$NPM_LUCIDE_REACT"
NPM_TAURI_API="$NPM_TAURI_API"
NPM_TAURI_CLI="$NPM_TAURI_CLI"
EOF

cat > "$OUT_MD" <<EOF
# Версии компонентов

Собрано: $COLLECTED_AT

## Система

- ОС: $OS_VERSION
- Ядро: $KERNEL
- Архитектура: $ARCH

## Runtime

- Go latest: $GO_LATEST
- Node LTS latest: $NODE_LATEST

## APT

- postgresql: $APT_POSTGRESQL
- postgresql-contrib: $APT_POSTGRESQL_CONTRIB
- valkey-server: $APT_VALKEY_SERVER
- valkey: $APT_VALKEY
- libvips-dev: $APT_LIBVIPS_DEV
- libvips-tools: $APT_LIBVIPS_TOOLS
- caddy: $APT_CADDY

## GitHub releases

- Caddy: $GH_CADDY
- Valkey: $GH_VALKEY
- Tauri: $GH_TAURI

## crates.io

- tauri: $CRATE_TAURI
- tauri-build: $CRATE_TAURI_BUILD

## npm

- react: $NPM_REACT
- react-dom: $NPM_REACT_DOM
- vite: $NPM_VITE
- @vitejs/plugin-react: $NPM_VITE_PLUGIN_REACT
- typescript: $NPM_TYPESCRIPT
- @types/react: $NPM_TYPES_REACT
- @types/react-dom: $NPM_TYPES_REACT_DOM
- zustand: $NPM_ZUSTAND
- @tanstack/react-query: $NPM_TANSTACK_QUERY
- dexie: $NPM_DEXIE
- motion: $NPM_MOTION
- lucide-react: $NPM_LUCIDE_REACT
- @tauri-apps/api: $NPM_TAURI_API
- @tauri-apps/cli: $NPM_TAURI_CLI
EOF

chmod 644 "$OUT_ENV" "$OUT_MD"

echo "Собрано:"
echo "$OUT_ENV"
echo "$OUT_MD"
echo
cat "$OUT_ENV"