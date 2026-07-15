#!/bin/bash
# Enable a 1-minute Windguru collector poll for one station (run as root).
# Per-spot interval and hours are enforced in the collector.
# Usage: add-wg-timer.sh STATION_ID
set -euo pipefail

STATION_ID="${1:-}"
if [[ ! "$STATION_ID" =~ ^[0-9]+$ ]]; then
  echo "usage: $0 STATION_ID" >&2
  exit 1
fi

SERVICE=/etc/systemd/system/ikite-wg-collector@.service
if [[ ! -f "$SERVICE" ]]; then
  tee "$SERVICE" >/dev/null <<'UNIT'
[Unit]
Description=ikite-go Windguru station %i
After=mysql.service

[Service]
Type=oneshot
User=ikite
Group=ikite
WorkingDirectory=/opt/ikite-go
EnvironmentFile=/etc/ikite-go/env
Environment=MIGRATE=0
ExecStart=/opt/ikite-go/bin/collector -wg-station=%i
UNIT
fi

TIMER="/etc/systemd/system/ikite-wg-collector@${STATION_ID}.timer"

offset=$(( (STATION_ID % 15) * 4 ))
tee "$TIMER" >/dev/null <<UNIT
[Unit]
Description=Windguru station ${STATION_ID} every minute

[Timer]
OnBootSec=$((3 * 60 + offset))s
OnUnitActiveSec=1min
Persistent=true

[Install]
WantedBy=timers.target
UNIT

systemctl daemon-reload
systemctl enable --now "ikite-wg-collector@${STATION_ID}.timer"
echo "Enabled timer for Windguru station ${STATION_ID}"
