# Authentication

The miner handles authentication automatically. On every startup it walks through a priority chain and stops at the first method that succeeds.

## Priority chain

| Priority | Method | When it's used |
|----------|--------|----------------|
| 1 | **Cookie file** | A token was saved from a previous successful login. Reused automatically; a refresh is attempted if the cookie has expired. |
| 2 | **`auth_token` in YAML** | Token set directly in the config file under `auth_token:`. Not recommended for production — prefer env vars. |
| 3 | **`TWITCH_AUTH_TOKEN_<USERNAME>`** | Environment variable. Recommended for headless and multi-account setups. |
| 4 | **`TWITCH_PASSWORD_<USERNAME>`** | Environment variable. Last resort — may require 2FA, which makes it unreliable in headless environments. |
| 5 | **Device code flow** | Interactive — the miner prints a code and a URL; you activate it on `https://www.twitch.tv/activate`. Token is saved as a cookie for future runs. |

After authentication, the token is validated against the Twitch OAuth2 endpoint to confirm it belongs to the expected username (derived from the config filename). A mismatch produces a clear error:

```
authenticated as "wrong_user" but config expects "your_username" — please log in with the correct account
```

## Recommended approach per environment

| Environment | Recommended method |
|-------------|--------------------|
| Local development | Device code flow (one-time, then reuses cookie) |
| Docker / Fly.io / cloud VM | `TWITCH_AUTH_TOKEN_<USERNAME>` env var |
| Multi-account | `TWITCH_AUTH_TOKEN_<USERNAME>` per account |
| CI/CD | `TWITCH_AUTH_TOKEN_<USERNAME>` via secrets |

## Variable naming convention

The variable suffix is the username in **uppercase**, with hyphens replaced by underscores.

| Username | Variable |
|----------|----------|
| `guliveer_` | `TWITCH_AUTH_TOKEN_GULIVEER_` |
| `my-user` | `TWITCH_AUTH_TOKEN_MY_USER` |
| `streamer123` | `TWITCH_AUTH_TOKEN_STREAMER123` |

The same convention applies to `TWITCH_PASSWORD_<USERNAME>`.

## Obtaining an OAuth token

The easiest way to get a token to use with `TWITCH_AUTH_TOKEN_<USERNAME>` is to complete the device code flow once locally, then copy the token from the saved cookie file:

```bash
# Cookie files are saved to the DATA_DIR directory (default: current directory)
cat cookies/your_username.json
# The token is in the "access_token" field
```

Alternatively, use [Twitch Token Generator](https://twitchtokengenerator.com/) — generate a token with the `chat:read` and `channel:read:predictions` scopes at minimum.

## Token scopes required

The miner uses the OAuth token to make authenticated Twitch API calls. The device code flow and password login request the necessary scopes automatically. If you supply a token manually via env var, ensure it includes at minimum:

- `channel:read:predictions`
- `channel:manage:predictions`
- `channel:read:redemptions`
- `drops:claim`
- `chat:read` (for IRC-based features)

## Cookie persistence

Validated tokens are saved to `{DATA_DIR}/cookies/{username}.json`. On subsequent runs, the cookie is loaded and reused — no re-authentication needed unless the token expires.

Cookie files are written atomically (temp file + rename), so a crash mid-write never leaves a truncated token file behind.

If `COOKIE_ENCRYPTION_KEY` is set but cannot be parsed as a Base64-encoded 32-byte AES-256 key, the miner refuses to start instead of silently falling back to a plaintext cookie jar — generate a key with `./tools/gen-cookie-key.sh`. See README section 1.6.5 for details.

In Docker, mount a persistent volume to `DATA_DIR` to preserve cookies across container restarts:

```bash
docker run -v miner_data:/data -e DATA_DIR=/data ...
```
