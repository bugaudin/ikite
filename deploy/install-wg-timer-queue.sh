#!/bin/bash
# One-time server setup: queue directory + systemd timer for new Windguru timers.
set -euo pipefail

QUEUE_DIR=/var/lib/ikite-go/wg-timer-queue
INSTALL=/opt/ikite-go

sudo mkdir -p "$QUEUE_DIR"
sudo chown ikite:ikite "$QUEUE_DIR"
sudo chmod 755 "$QUEUE_DIR"

sudo chmod +x "$INSTALL/deploy/process-wg-timer-queue.sh"
sudo chmod +x "$INSTALL/deploy/add-wg-timer.sh"

sudo tee /etc/systemd/system/ikite-wg-timer-queue.service >/dev/null <<UNIT
[Unit]
Description=Process pending Windguru collector timer requests

[Service]
Type=oneshot
Environment=WG_TIMER_QUEUE_DIR=$QUEUE_DIR
Environment=WG_TIMER_SCRIPT=$INSTALL/deploy/add-wg-timer.sh
ExecStart=$INSTALL/deploy/process-wg-timer-queue.sh
UNIT

sudo tee /etc/systemd/system/ikite-wg-timer-queue.timer >/dev/null <<'UNIT'
[Unit]
Description=Process Windguru timer queue every minute

[Timer]
OnBootSec=1min
OnUnitActiveSec=1min
Persistent=true

[Install]
WantedBy=timers.target
UNIT

sudo systemctl daemon-reload
sudo systemctl enable --now ikite-wg-timer-queue.timer

echo "Installed wg-timer queue at $QUEUE_DIR (processed every minute via systemd)"
