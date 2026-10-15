#!/usr/bin/env bash
set -euo pipefail

# Applies local metadata from ./hasura/metadata to local Hasura engine at http://localhost:8000
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

ENDPOINT="${LOCAL_HASURA_ENDPOINT:-http://localhost:8000}"
ADMIN_SECRET="${HASURA_GRAPHQL_ADMIN_SECRET:-myadminsecretkey}"

echo "Applying metadata to local Hasura engine at ${ENDPOINT}..."
hasura metadata apply \
  --project "${ROOT_DIR}/hasura" \
  --endpoint "${ENDPOINT}" \
  --admin-secret "${ADMIN_SECRET}"

echo "Metadata applied successfully to ${ENDPOINT}"
