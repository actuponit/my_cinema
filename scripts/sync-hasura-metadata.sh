#!/usr/bin/env bash
set -euo pipefail

# Pulls live Hasura metadata from remote instance into ./hasura/metadata
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

if [ -f "${ROOT_DIR}/.env" ]; then
  # shellcheck disable=SC1091
  source "${ROOT_DIR}/.env"
fi

ENDPOINT="${REMOTE_HASURA_ENDPOINT:-https://pleasing-viper-86.hasura.app}"
ADMIN_SECRET="${HASURA_GRAPHQL_ADMIN_SECRET:-myadminsecretkey}"

echo "Exporting Hasura metadata from ${ENDPOINT}..."
hasura metadata export \
  --project "${ROOT_DIR}/hasura" \
  --endpoint "${ENDPOINT}" \
  --admin-secret "${ADMIN_SECRET}"

echo "Hasura metadata exported successfully to ${ROOT_DIR}/hasura/metadata"
