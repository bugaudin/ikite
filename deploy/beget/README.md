# Beget / shared-hosting proxy

Upload **one** script to your hosting account:

| File | Env variable |
|------|----------------|
| `proxy_post.php` | `BEGET_PROXY_URL` — e.g. `https://your-host.example/proxy_post.php` |

The Go app POSTs JSON to `proxy_post.php` with the real target URL, method, headers, and body. All upstream URLs and Windguru API details live in this repo — you never edit PHP on Beget when APIs change.

Upstream URLs (fetched via the proxy, not called directly):

| Env variable | Example upstream |
|--------------|------------------|
| `KY_HISTORY_URL` | `https://surfo.co.il/wp-content/themes/vibes-child/inc/weather/data/api_wind.php` |
| `SURFO_LIVE_URL` | `https://surfo.co.il/.../surfo_live.json` |

Windguru station and forecast targets are built in Go (`internal/sources/windguru/`).

Do not commit hosting credentials or account-specific URLs to git.
