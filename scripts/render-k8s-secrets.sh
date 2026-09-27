#!/usr/bin/env bash
# .env -> k8s/01-config-secrets.yaml (gitignore'da). Hardcode yok.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$ROOT/.env"
OUT="$ROOT/k8s/01-config-secrets.yaml"

if [ ! -f "$ENV_FILE" ]; then
  echo "Yok: $ENV_FILE  →  cp .env.example .env && doldur" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

: "${DB_USER:?DB_USER boş}"
: "${DB_PASS:?DB_PASS boş}"
: "${MINIO_ROOT_USER:?MINIO_ROOT_USER boş}"
: "${MINIO_ROOT_PASSWORD:?MINIO_ROOT_PASSWORD boş}"
: "${ADMIN_USER:?ADMIN_USER boş}"
: "${ADMIN_PASS:?ADMIN_PASS boş}"
: "${GRAFANA_USER:?GRAFANA_USER boş}"
: "${GRAFANA_PASS:?GRAFANA_PASS boş}"

umask 077
cat >"$OUT" <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: sadecekedi-config
data:
  POSTGRES_DB: sadecekedi
---
apiVersion: v1
kind: Secret
metadata:
  name: sadecekedi-secrets
type: Opaque
stringData:
  DB_USER: "${DB_USER}"
  DB_PASS: "${DB_PASS}"
  DATABASE_URL: "postgres://${DB_USER}:${DB_PASS}@postgres-service:5432/sadecekedi?sslmode=disable"
  MINIO_ROOT_USER: "${MINIO_ROOT_USER}"
  MINIO_ROOT_PASSWORD: "${MINIO_ROOT_PASSWORD}"
  MINIO_ACCESS_KEY: "${MINIO_ROOT_USER}"
  MINIO_SECRET_KEY: "${MINIO_ROOT_PASSWORD}"
  ADMIN_USER: "${ADMIN_USER}"
  ADMIN_PASS: "${ADMIN_PASS}"
  GRAFANA_USER: "${GRAFANA_USER}"
  GRAFANA_PASS: "${GRAFANA_PASS}"
  CLOUDFLARE_TOKEN: "${CLOUDFLARE_TOKEN:-}"
EOF

echo "yazıldı: $OUT (git dışı)"
