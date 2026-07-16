# Beget / shared-hosting proxy

Upload **one** script to your hosting account:

| File | Env variable |
|------|----------------|
| `proxy_post.php` | `BEGET_PROXY_URL` — e.g. `https://your-host.example/proxy_post.php` |

`proxy_post.php` is **not** in git (contains a hardcoded secret). Copy from `proxy_post.php.example`, set `PROXY_SECRET` (`openssl rand -hex 32`), upload to Beget, and set the same value as `BEGET_PROXY_SECRET` in ikite env.

The Go app POSTs JSON to `proxy_post.php` with header `X-Proxy-Secret`, plus the real target URL, method, headers, and body.

Upstream URLs (fetched via the proxy, not called directly):

| Env variable | Example upstream |
|--------------|------------------|
| `KY_HISTORY_URL` | `https://surfo.co.il/wp-content/themes/vibes-child/inc/weather/data/api_wind.php` |
| `SURFO_LIVE_URL` | `https://surfo.co.il/.../surfo_live.json` |

Windguru station and forecast targets are built in Go (`internal/sources/windguru/`).

Do not commit hosting credentials, proxy secrets, or account-specific URLs to git.
