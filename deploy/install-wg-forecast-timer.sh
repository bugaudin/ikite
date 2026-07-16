#!/bin/bash
# Install systemd timer for daily Windguru forecast fetch at 07:00 Israel time.
set -euo pipefail

INSTALL="${1:-/opt/ikite-go}"

sudo tee /etc/systemd/system/ikite-wgforecast.service >/dev/null <<UNIT
[Unit]
Description=ikite-go Windguru forecast collector
After=network.target mysql.service
Requires=mysql.service

[Service]
Type=oneshot
User=ikite
Group=ikite
WorkingDirectory=${INSTALL}
EnvironmentFile=/etc/ikite-go/env
ExecStart=${INSTALL}/bin/wgforecast
UNIT

sudo tee /etc/systemd/system/ikite-wgforecast.timer >/dev/null <<'UNIT'
[Unit]
Description=Daily Windguru forecast at 07:00 Asia/Jerusalem

[Timer]
OnCalendar=*-*-* 07:00:00 Asia/Jerusalem
Persistent=true
Unit=ikite-wgforecast.service

[Install]
WantedBy=timers.target
UNIT

sudo systemctl daemon-reload
sudo systemctl enable --now ikite-wgforecast.timer
echo "Enabled ikite-wgforecast.timer (daily 07:00 Asia/Jerusalem)"
