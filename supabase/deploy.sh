#!/usr/bin/env bash
set -euo pipefail

# Deploys all AutoSuq Supabase Edge Functions to project weunvxrpkfeznuodhhye
PROJECT_REF="weunvxrpkfeznuodhhye"

echo "=== Deploying Supabase Edge Functions to project ${PROJECT_REF} ==="

FUNCTIONS=(
  "wallet-pay"
  "wallet-topup"
  "wallet-topup-status"
  "wallet-credentials"
  "verify-webhook"
  "chapa-webhook"
  "chapa-initialize"
)

for fn in "${FUNCTIONS[@]}"; do
  echo "--> Deploying function: ${fn}"
  supabase functions deploy "${fn}" --project-ref "${PROJECT_REF}" --no-verify-jwt --use-api
done

echo "=== All functions deployed successfully! ==="
echo ""
echo "Webhook URLs:"
echo "  Verify.et: https://${PROJECT_REF}.supabase.co/functions/v1/verify-webhook"
echo "  Chapa:     https://${PROJECT_REF}.supabase.co/functions/v1/chapa-webhook"
