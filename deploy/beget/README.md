# Beget / shared-hosting proxies

Upload these PHP scripts to your hosting account and set the URLs in environment:

| File | Env variable |
|------|----------------|
| `wg_station.php` | `BEGET_WG_STATION_URL` — e.g. `https://your-host.example/wg_station.php?id_station=%d` |
| `wind_history.php` (if used) | `KY_HISTORY_URL` |
| `proxy.php` (if used) | `SURFO_LIVE_URL` |

Do not commit hosting credentials or account-specific URLs to git.
