# ikite-go

Israeli Mediterranean wind monitoring for kitesurfing/windsurfing.

## Features

- **Live wind table** (`/`) — multi-station readings with color-coded speeds and direction arrows
- **Graph** (`/graph`) — ApexCharts time series
- **Settings** (`/settings?pass=<uuid>`) — hidden admin page (threshold, spots, collectors)
- **Beach cameras** (`/camera`)
- **Home sensor ingest** (`/home?w=<knots>`)
- **Collector** — polls Windguru stations, Kiryat Yam history, Kiryat Haim (windometer), sends Telegram alerts
- **Forecast job** — fetches Hebrew AI report from surfo, translates to English, stores + Telegram

## Quick start

```bash
cp .env.example .env
# Edit .env — set DB_PASSWORD, proxy URLs, SETTINGS_PASS, Telegram tokens

docker compose up -d          # MySQL (dev credentials in compose file only)
make build
MIGRATE=1 ./bin/server        # http://localhost:8080
./bin/collector               # run every ~5 min via cron
./bin/forecast                # run every ~15–30 min via cron
```

## Cron examples

```cron
*/5 * * * *  cd /path/to/ikite-go && ./bin/collector >> /var/log/ikite-collector.log 2>&1
*/20 * * * * cd /path/to/ikite-go && ./bin/forecast  >> /var/log/ikite-forecast.log 2>&1
```

Per-station Windguru timers (production): see `deploy/setup-wg-timers.sh` and `deploy/add-wg-timer.sh`.

## Config

**All secrets and deployment-specific URLs must come from environment.** See `.env.example` (local) and `deploy/env.example` (production server).

| Variable | Purpose |
|----------|---------|
| `DB_*` / `DB_DSN` | MySQL connection |
| `SETTINGS_PASS` | UUID for `/settings?pass=…` (required to access admin) |
| `TELEGRAM_ALERT_TOKEN` / `TELEGRAM_ALERT_CHAT_ID` | Wind alert bot |
| `TELEGRAM_AI_TOKEN` / `TELEGRAM_AI_CHAT_ID` | Forecast bot |
| `KY_HISTORY_URL` | Kiryat Yam history HTML proxy |
| `SURFO_LIVE_URL` | AI forecast JSON proxy |
| `BEGET_WG_STATION_URL` | Windguru station proxy (`%d` = station id) |
| `WG_TIMER_QUEUE_DIR` | Directory for pending timer requests (web writes, root cron processes) |
| `WG_TIMER_SCRIPT` | Path to `deploy/add-wg-timer.sh` (used by queue processor) |

## Layout

```
cmd/server      HTTP dashboard
cmd/collector   wind poll + alerts
cmd/forecast    AI forecast job
internal/       packages (store, sources, notify, web)
migrations/     MySQL schema
deploy/         systemd timers, PHP proxies, env template
```

Spots (names, Windguru IDs, visibility) are stored in the `spots` database table.

## Publishing / security

- `.env`, `server-keys/`, `*.pem`, and `bin/` are gitignored
- Never commit Telegram tokens, DB passwords, `SETTINGS_PASS`, or SSH keys
- Upload `deploy/beget/*.php` to your own hosting; configure URLs via env

## Not ported (yet)

Windguru global spot-rating crawler (`wg_*.php`, `spots_*.php`) — secondary subsystem from the PHP site.
