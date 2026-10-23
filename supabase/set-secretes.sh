#!/usr/bin/env bash
set -euo pipefail

# Copy to set-secrets.sh (gitignored by you), fill in real values, run once.
# SUPABASE_URL / SUPABASE_ANON_KEY / SUPABASE_SERVICE_ROLE_KEY are injected by Supabase automatically.
PROJECT_REF="weunvxrpkfeznuodhhye"

echo "Setting secrets for project ${PROJECT_REF}..."
supabase secrets set \
  --project-ref "${PROJECT_REF}" \
  KIOSK_SECRET="08fa77eeaecad495aa71b5f7d6fa4cdffff876ae7880cef09777cbdf83570377" \
  VERIFY_API_KEY="vchk_eLAxyDcEVkXXJqYOOEWHOUbvvoIrbIyRtUqPtXiFPqDnAebxttQfmEEnkgVprnRL" \
  VERIFY_WEBHOOK_SECRET="whsec_faBkv38EdvjbbpyQJoLrqDJwS-SR4zz0Cf6JSWWkf0I" \
  VERIFY_API_BASE_URL="${VERIFY_API_BASE_URL:-https://checkoutapi.verify.et}" \
  WALLET_WEB_URL="${WALLET_WEB_URL:-https://wallet.autosuq.com}" \
  CHAPA_SECRET_KEY="${CHAPA_SECRET_KEY:?}" \
  CHAPA_WEBHOOK_SECRET="${CHAPA_WEBHOOK_SECRET:?}" \
  MQTT_BROKER_URL="${MQTT_BROKER_URL:-mqtts://0b8c3235cf1145668f5d9841d803ed68.s1.eu.hivemq.cloud:8883}" \
  MQTT_USERNAME="autosuq" \
  MQTT_PASSWORD="Jannatul4dows"
