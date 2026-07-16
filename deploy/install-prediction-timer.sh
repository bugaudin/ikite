#!/bin/bash
# Install systemd timer for daily KY thermal prediction at 11:00 in July+August (Asia/Jerusalem).
set -euo pipefail

INSTALL="${1:-/opt/ikite-go}"

sudo tee /etc/systemd/system/ikite-prediction.service >/dev/null <<UNIT
[Unit]
Description=ikite-go KY thermal prediction Telegram notify
After=network.target mysql.service
Requires=mysql.service

[Service]
Type=oneshot
User=ikite
Group=ikite
WorkingDirectory=${INSTALL}
EnvironmentFile=/etc/ikite-go/env
ExecStart=${INSTALL}/bin/prediction
UNIT

sudo tee /etc/systemd/system/ikite-prediction.timer >/dev/null <<'UNIT'
[Unit]
Description=Daily KY prediction at 11:00 in July and August (Asia/Jerusalem)

[Timer]
OnCalendar=*-07-* 11:00:00 Asia/Jerusalem
OnCalendar=*-08-* 11:00:00 Asia/Jerusalem
Persistent=true
Unit=ikite-prediction.service

[Install]
WantedBy=timers.target
UNIT

sudo systemctl daemon-reload
sudo systemctl enable --now ikite-prediction.timer
echo "Enabled ikite-prediction.timer (daily 11:00 in July+August, Asia/Jerusalem)"
