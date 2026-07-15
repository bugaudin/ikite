#!/bin/bash
# Process Windguru timer requests queued by the web app (run as root via cron).
set -euo pipefail

QUEUE_DIR="${WG_TIMER_QUEUE_DIR:-/var/lib/ikite-go/wg-timer-queue}"
SCRIPT="${WG_TIMER_SCRIPT:-/opt/ikite-go/deploy/add-wg-timer.sh}"

if [[ ! -d "$QUEUE_DIR" ]]; then
  exit 0
fi

shopt -s nullglob
for f in "$QUEUE_DIR"/*; do
  [[ -f "$f" ]] || continue
  id="$(basename "$f")"
  if [[ ! "$id" =~ ^[0-9]+$ ]]; then
    rm -f "$f"
    continue
  fi
  if "$SCRIPT" "$id"; then
    rm -f "$f"
  else
    echo "failed to enable timer for station $id" >&2
  fi
done
