#!/usr/bin/env bash
# Finish OCI CLI setup after uploading the public key in Oracle Console.
# Usage:
#   TENANCY_OCID=ocid1.tenancy... USER_OCID=ocid1.user... ./scripts/oci-finish-setup.sh
# Optional: INSTANCE_OCID=... or INSTANCE_NAME=ikite (default) to save ~/.oci/ikite-instance.env

set -euo pipefail

REGION="${OCI_REGION:-il-jerusalem-1}"
TENANCY_OCID="${TENANCY_OCID:-}"
USER_OCID="${USER_OCID:-}"
KEY_FILE="${HOME}/.oci/oci_api_key.pem"

if [[ -z "$TENANCY_OCID" || -z "$USER_OCID" ]]; then
  echo "Set TENANCY_OCID and USER_OCID from Oracle Console." >&2
  echo "  Tenancy: Administration → Tenancy details → OCID" >&2
  echo "  User:    Profile → User settings → OCID" >&2
  exit 1
fi

if [[ ! -f "$KEY_FILE" ]]; then
  echo "Missing $KEY_FILE — run: openssl genrsa -out $KEY_FILE 2048" >&2
  exit 1
fi

FINGERPRINT=$(openssl rsa -pubout -outform DER -in "$KEY_FILE" 2>/dev/null | openssl md5 -c | awk '{print $2}')

mkdir -p "${HOME}/.oci"
chmod 700 "${HOME}/.oci"

cat > "${HOME}/.oci/config" <<EOF
[DEFAULT]
user=${USER_OCID}
fingerprint=${FINGERPRINT}
tenancy=${TENANCY_OCID}
region=${REGION}
key_file=${KEY_FILE}
EOF
chmod 600 "${HOME}/.oci/config"

echo "Wrote ~/.oci/config (region=${REGION})"
echo "Testing API access..."
oci iam region list --query 'data[0].name' --raw-output >/dev/null
echo "OK — OCI CLI is configured."

if [[ -n "${INSTANCE_OCID:-}" ]]; then
  INSTANCE_ID="$INSTANCE_OCID"
else
  INSTANCE_PATTERN="${INSTANCE_NAME:-ikite}"
  echo ""
  echo "Finding OCI instance (display name contains \"${INSTANCE_PATTERN}\")..."
  INSTANCE_ID=$(oci compute instance list --all \
    --query "data[?\"lifecycle-state\"=='RUNNING'].{id:id,name:\"display-name\"}" \
    --output json 2>/dev/null | INSTANCE_PATTERN="$INSTANCE_PATTERN" python3 -c "
import json, os, sys
pattern = os.environ.get('INSTANCE_PATTERN', 'ikite').lower()
for i in json.load(sys.stdin):
    if pattern in (i.get('name') or '').lower():
        print(i['id'])
        break
" 2>/dev/null || true)
fi

if [[ -n "$INSTANCE_ID" ]]; then
  echo "instance_ocid=${INSTANCE_ID}" > "${HOME}/.oci/ikite-instance.env"
  chmod 600 "${HOME}/.oci/ikite-instance.env"
  echo "Saved instance OCID to ~/.oci/ikite-instance.env"
else
  echo "Could not auto-detect instance — set INSTANCE_OCID or INSTANCE_NAME, or run: oci compute instance list --all"
fi
