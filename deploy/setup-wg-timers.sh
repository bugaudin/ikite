#!/bin/bash
set -euo pipefail

STATIONS=(2763 2049 3377 2256 1909 2752 3379 5730 5731 5732 1091 5500 14905 2667 3708 15233)

sudo tee /etc/systemd/system/ikite-wg-collector@.service >/dev/null <<'UNIT'
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

offset=0
for id in "${STATIONS[@]}"; do
  sudo tee "/etc/systemd/system/ikite-wg-collector@${id}.timer" >/dev/null <<UNIT
[Unit]
Description=Windguru station ${id} every minute

[Timer]
OnBootSec=$((3 * 60 + offset))s
OnUnitActiveSec=1min
Persistent=true

[Install]
WantedBy=timers.target
UNIT
  sudo systemctl enable --now "ikite-wg-collector@${id}.timer"
  offset=$((offset + 4))
done

sudo systemctl daemon-reload
echo "Enabled ${#STATIONS[@]} windguru station timers (4s apart, poll every 1 min)"
