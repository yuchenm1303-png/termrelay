#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

if ! command -v openssl >/dev/null 2>&1; then
  echo "openssl is required to generate deployment secrets." >&2
  exit 1
fi

if [[ -f .env ]]; then
  echo ".env already exists; refusing to overwrite it."
  exit 0
fi

cp .env.example .env

replace_value() {
  local key="$1"
  local value="$2"
  sed -i "s|^${key}=.*|${key}=${value}|" .env
}

postgres_password="$(openssl rand -hex 24)"
redis_password="$(openssl rand -hex 24)"
admin_password="$(openssl rand -hex 12)"
jwt_secret="$(openssl rand -hex 32)"
totp_key="$(openssl rand -hex 32)"

replace_value POSTGRES_PASSWORD "$postgres_password"
replace_value REDIS_PASSWORD "$redis_password"
replace_value ADMIN_PASSWORD "$admin_password"
replace_value JWT_SECRET "$jwt_secret"
replace_value TOTP_ENCRYPTION_KEY "$totp_key"

chmod 600 .env
mkdir -p data/app data/postgres data/redis data/caddy/data data/caddy/config

cat <<EOF
TermRelay deployment files are prepared.

Generated administrator password:
  ${admin_password}

Before starting:
  1. Edit .env and set TERMRELAY_DOMAIN.
  2. Edit ADMIN_EMAIL if needed.
  3. Store the administrator password in a password manager.

Start with:
  docker compose pull
  docker compose up -d
EOF
