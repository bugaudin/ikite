#!/bin/bash
# Enable a 5-minute Windguru collector timer for one station (run as root).
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
if [[ -f "$TIMER" ]]; then
  systemctl enable --now "ikite-wg-collector@${STATION_ID}.timer"
  echo "Timer already exists for station ${STATION_ID}"
  exit 0
fi

offset=$(( (STATION_ID % 15) * 20 ))
tee "$TIMER" >/dev/null <<UNIT
[Unit]
Description=Windguru station ${STATION_ID} every 5 minutes

[Timer]
OnBootSec=$((3 * 60 + offset))s
OnUnitActiveSec=5min
Persistent=true

[Install]
WantedBy=timers.target
UNIT

systemctl daemon-reload
systemctl enable --now "ikite-wg-collector@${STATION_ID}.timer"
echo "Enabled timer for Windguru station ${STATION_ID}"
