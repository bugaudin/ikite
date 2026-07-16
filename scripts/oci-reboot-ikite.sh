#!/usr/bin/env bash
# Soft-reboot the ikite Oracle instance via OCI API.
set -euo pipefail

ENV_FILE="${HOME}/.oci/ikite-instance.env"
if [[ -f "$ENV_FILE" ]]; then
  # shellcheck source=/dev/null
  source "$ENV_FILE"
fi

INSTANCE_ID="${INSTANCE_OCID:-${instance_ocid:-}}"
if [[ -z "$INSTANCE_ID" ]]; then
  echo "No instance OCID. Run scripts/oci-finish-setup.sh first, or set INSTANCE_OCID." >&2
  exit 1
fi

echo "Rebooting instance ${INSTANCE_ID}..."
oci compute instance action --action SOFTRESET --instance-id "$INSTANCE_ID"
echo "Reboot requested. Wait ~2 min, then: curl https://ikite.fyi/healthz"
