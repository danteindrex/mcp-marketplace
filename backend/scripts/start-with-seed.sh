#!/usr/bin/env bash
set -euo pipefail

SUPER_ADMIN_EMAIL="${SUPER_ADMIN_EMAIL:-admin@platform.local}"
SUPER_ADMIN_PASSWORD="${SUPER_ADMIN_PASSWORD:-}"
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_SECRET="${JWT_SECRET:-}"
MONGO_URI="${MONGO_URI:-mongodb://localhost:27017}"
MONGO_DB_NAME="${MONGO_DB_NAME:-mcp_marketplace}"

if [[ -z "${SUPER_ADMIN_PASSWORD}" ]]; then
  SUPER_ADMIN_PASSWORD="$(openssl rand -hex 16)Aa1!"
fi

if [[ -z "${JWT_SECRET}" ]]; then
  JWT_SECRET="$(openssl rand -hex 32)"
fi

export SUPER_ADMIN_EMAIL
export SUPER_ADMIN_PASSWORD
export BASE_URL
export JWT_SECRET
export MONGO_URI
export MONGO_DB_NAME
export ALLOW_INSECURE_DEFAULTS=false

echo "Seeding super admin..."
go run ./cmd/seed-super-admin

echo
echo "Super admin credentials:"
echo "  Email:    ${SUPER_ADMIN_EMAIL}"
echo "  Password: ${SUPER_ADMIN_PASSWORD}"
echo
echo "Starting backend server..."
go run ./cmd/server
