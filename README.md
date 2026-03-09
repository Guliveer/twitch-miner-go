# Twitch Channel Points Miner — Go Edition

[![GitHub Stars](https://img.shields.io/github/stars/Guliveer/twitch-miner-go?style=for-the-badge&logo=github&color=gold)](https://github.com/Guliveer/twitch-miner-go/stargazers)
[![CI](https://img.shields.io/github/actions/workflow/status/Guliveer/twitch-miner-go/ci.yml?style=for-the-badge&logo=github&label=CI)](https://github.com/Guliveer/twitch-miner-go/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/v/release/Guliveer/twitch-miner-go?style=for-the-badge&logo=semanticrelease&label=latest)](https://github.com/Guliveer/twitch-miner-go/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Guliveer/twitch-miner-go?style=for-the-badge&logo=go)](https://go.dev/)
[![License](https://img.shields.io/github/license/Guliveer/twitch-miner-go?style=for-the-badge)](https://github.com/Guliveer/twitch-miner-go/blob/main/LICENSE.txt)
[![Go Report Card](https://goreportcard.com/badge/github.com/Guliveer/twitch-miner-go?style=for-the-badge)](https://goreportcard.com/report/github.com/Guliveer/twitch-miner-go)

A high-performance Go rewrite of the [Twitch Channel Points Miner v2](https://github.com/rdavydov/Twitch-Channel-Points-Miner-v2). Mines channel points, claims bonuses, places predictions, joins raids, claims drops, and more — all with a fraction of the resource usage.

## Features

- **Multi-account support** — run multiple Twitch accounts from a single binary
- **Channel points mining** — automatic minute-watched events, bonus claims, watch streaks
- **Predictions** — configurable betting strategies (SMART, HIGH_ODDS, MOST_VOTED, etc.)
- **Drops** — automatic campaign sync and drop claiming
- **Raids** — automatic raid joining
- **Community moments** — automatic moment claiming
- **Community goals** — automatic goal contributions
- **Gift sub detection** — notifies when your account receives a gifted subscription
- **Category watcher** — auto-discover streamers by game category
- **Notifications** — Telegram, Discord, Webhook, Matrix, Pushover, Gotify
- **Analytics dashboard** — built-in web UI for monitoring
- **Fly.io ready** — deploy with a single command

## Resource Comparison

| Resource     | Python (per account) | Go (per account)  |
| ------------ | -------------------- | ----------------- |
| Memory       | ~80–120 MB           | ~5–15 MB          |
| Docker image | ~800 MB              | ~10–15 MB         |
| Startup time | ~5–10 s              | < 100 ms          |
| Threads      | 60+                  | ~10–20 goroutines |

## Running Locally

**Prerequisites:** [Go 1.24+](https://go.dev/dl/)

**Unix (macOS/Linux):**

```bash
./run.sh
```

**Windows:**

```batch
run.bat
```

**With custom flags:**

```bash
./run.sh -config configs -port 9090 -log-level debug
```

> The scripts build the binary and run it in one step. You can also build manually with `go build -o twitch-miner-go ./cmd/twitch-miner-go`.

### Flags

| Flag         | Default   | Description                                                     |
| ------------ | --------- | --------------------------------------------------------------- |
| `-config`    | `configs` | Path to the configuration directory                             |
| `-port`      | `8080`    | Port for the health/analytics server                            |
| `-log-level` | `INFO`    | Log level: DEBUG, INFO, WARN, ERROR (effective default: `INFO`) |
| `-version`   | `false`   | Print version and exit                                          |

## Configuration

Create one YAML file per account in the `configs/` directory. **The filename (without extension) becomes the Twitch username** — no `username` field is needed in the YAML.

For example, to add an account for Twitch user `guliveer_`, create `configs/guliveer_.yaml`.

```bash
# Copy the example and customize for your account
cp configs/example.yaml.example configs/your_twitch_username.yaml
```

See [`configs/example.yaml.example`](configs/example.yaml.example) for the full schema. Files with a `.yaml.example` extension are not loaded as configs — only `.yaml` and `.yml` files are loaded.

### Quick Start

```yaml
# configs/your_twitch_username.yaml
# The filename IS the username — no username field needed.

# Set to false to disable this account without deleting the config (default: true)
enabled: true

features:
  claim_drops_startup: false
  enable_analytics: true

priority:
  - STREAK
  - DROPS
  - ORDER

streamer_defaults:
  make_predictions: true
  follow_raid: true
  claim_drops: true
  claim_moments: true
  watch_streak: true
  community_goals: false
  chat: "ONLINE"
  bet:
    strategy: "SMART"
    percentage: 5
    max_points: 50000
    delay: 6
    delay_mode: "FROM_END"

streamers:
  - username: "streamer1"
  - username: "streamer2"
    settings:
      make_predictions: false

# Blacklisted streamers excluded even if followed
blacklist:
  - "unwanted_streamer"

# Follow mode — also watch all followed channels
followers:
  enabled: false
  order: "ASC"
```

### Environment Variables

Secrets and auth tokens are injected via environment variables. Per-account variables **require** the `_<USERNAME>` suffix (uppercase) to scope them to the correct account.

| Variable                         | Description                                         |
| -------------------------------- | --------------------------------------------------- |
| `TWITCH_AUTH_TOKEN_<USERNAME>`   | OAuth token (fallback for headless auth)            |
| `TWITCH_PASSWORD_<USERNAME>`     | Twitch password (last-resort auth, may require 2FA) |
| `TELEGRAM_TOKEN_<USERNAME>`      | Telegram bot token                                  |
| `TELEGRAM_CHAT_ID_<USERNAME>`    | Telegram chat ID                                    |
| `DISCORD_WEBHOOK_<USERNAME>`     | Discord webhook URL                                 |
| `WEBHOOK_URL_<USERNAME>`         | Generic webhook URL                                 |
| `MATRIX_HOMESERVER_<USERNAME>`   | Matrix homeserver URL                               |
| `MATRIX_ROOM_ID_<USERNAME>`      | Matrix room ID                                      |
| `MATRIX_ACCESS_TOKEN_<USERNAME>` | Matrix access token                                 |
| `PUSHOVER_TOKEN_<USERNAME>`      | Pushover API token                                  |
| `PUSHOVER_USER_KEY_<USERNAME>`   | Pushover user key                                   |
| `GOTIFY_URL_<USERNAME>`          | Gotify server URL                                   |
| `GOTIFY_TOKEN_<USERNAME>`        | Gotify app token                                    |
| `PORT`                           | HTTP server port                                    |
| `LOG_LEVEL`                      | Log level override                                  |

For example, for user `guliveer_` the Telegram token variable is `TELEGRAM_TOKEN_GULIVEER_` and the auth token variable is `TWITCH_AUTH_TOKEN_GULIVEER_`.

### `.env` File Support

The project supports loading environment variables from a `.env` file at startup using [`joho/godotenv`](https://github.com/joho/godotenv). This is **optional** — if no `.env` file is present, the app runs normally using YAML configs and/or standard environment variables.

Environment variables (whether from `.env` or the system) **override** the corresponding values from YAML config files for notification secrets only. This allows you to keep sensitive tokens out of version-controlled YAML files.

**Global variables:**

| Variable    | Description                                        | Default |
| ----------- | -------------------------------------------------- | ------- |
| `LOG_LEVEL` | Log level (`DEBUG`, `INFO`, `WARN`, `ERROR`)       | `INFO`  |
| `PORT`      | HTTP server port for the health/analytics endpoint | `8080`  |

**Per-account notification secret overrides** (pattern: `VARIABLE_<UPPERCASE_USERNAME>`):

| Variable Pattern                 | Description           |
| -------------------------------- | --------------------- |
| `TELEGRAM_TOKEN_<USERNAME>`      | Telegram bot token    |
| `TELEGRAM_CHAT_ID_<USERNAME>`    | Telegram chat ID      |
| `DISCORD_WEBHOOK_<USERNAME>`     | Discord webhook URL   |
| `WEBHOOK_URL_<USERNAME>`         | Generic webhook URL   |
| `MATRIX_HOMESERVER_<USERNAME>`   | Matrix homeserver URL |
| `MATRIX_ROOM_ID_<USERNAME>`      | Matrix room ID        |
| `MATRIX_ACCESS_TOKEN_<USERNAME>` | Matrix access token   |
| `PUSHOVER_TOKEN_<USERNAME>`      | Pushover API token    |
| `PUSHOVER_USER_KEY_<USERNAME>`   | Pushover user key     |
| `GOTIFY_URL_<USERNAME>`          | Gotify server URL     |
| `GOTIFY_TOKEN_<USERNAME>`        | Gotify app token      |

**Example `.env` file:**

```dotenv
# Global
LOG_LEVEL=DEBUG
PORT=9090

# Notification secrets for user "guliveer_"
TELEGRAM_TOKEN_GULIVEER_=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
TELEGRAM_CHAT_ID_GULIVEER_=987654321
DISCORD_WEBHOOK_GULIVEER_=https://discord.com/api/webhooks/...
```

> See [`.env.example`](.env.example) for the starter template.

## Notifications

The miner supports multiple notification providers. Configure them in your account YAML file under the `notifications` key. Sensitive credentials (tokens, API keys) are injected via environment variables — see [Environment Variables](#environment-variables) above.

### Supported Providers

| Provider | Config key | Required env vars                                                                             |
| -------- | ---------- | --------------------------------------------------------------------------------------------- |
| Telegram | `telegram` | `TELEGRAM_TOKEN_<USERNAME>`, `TELEGRAM_CHAT_ID_<USERNAME>`                                    |
| Discord  | `discord`  | `DISCORD_WEBHOOK_<USERNAME>`                                                                  |
| Gotify   | `gotify`   | `GOTIFY_URL_<USERNAME>`, `GOTIFY_TOKEN_<USERNAME>`                                            |
| Pushover | `pushover` | `PUSHOVER_TOKEN_<USERNAME>`, `PUSHOVER_USER_KEY_<USERNAME>`                                   |
| Matrix   | `matrix`   | `MATRIX_HOMESERVER_<USERNAME>`, `MATRIX_ROOM_ID_<USERNAME>`, `MATRIX_ACCESS_TOKEN_<USERNAME>` |
| Webhook  | `webhook`  | `WEBHOOK_URL_<USERNAME>`                                                                      |

> Replace `<USERNAME>` with the Twitch username in **UPPERCASE**. For example, user `guliveer_` → `TELEGRAM_TOKEN_GULIVEER_`.

### Example: Telegram

```yaml
notifications:
  telegram:
    enabled: true
    token: "YOUR_BOT_TOKEN"
    chat_id: "YOUR_CHAT_ID"
    events:
      - "DROP_CLAIM"
      - "DROP_STATUS"
      - "STREAMER_ONLINE"
      - "STREAMER_OFFLINE"
      - "BET_WIN"
      - "BET_LOSE"
    disable_notification: false
```

> **Tip:** The `token` and `chat_id` fields in YAML are optional — if omitted, the miner reads them from `TELEGRAM_TOKEN_<USERNAME>` and `TELEGRAM_CHAT_ID_<USERNAME>` environment variables instead. This is the recommended approach for production/headless deployments.

### Event Filtering

The `events` list controls which events trigger a notification for a given provider. Events are configured per notification provider in the YAML config under `notifications > {provider} > events`.

- If the `events` list is **empty or omitted**, **all** events are sent to that provider.
- If specific events are listed, **only those events** trigger notifications for that provider.

**Available events:**

| Event                   | Emoji | Description                       |
| ----------------------- | ----- | --------------------------------- |
| `STREAMER_ONLINE`       | 🟢    | Streamer goes online              |
| `STREAMER_OFFLINE`      | ⚫    | Streamer goes offline             |
| `GAIN_FOR_RAID`         | 💵    | Points gained from a raid         |
| `GAIN_FOR_CLAIM`        | 💵    | Points gained from claiming bonus |
| `GAIN_FOR_WATCH`        | 💵    | Points gained from watching       |
| `GAIN_FOR_WATCH_STREAK` | 💵    | Points gained from watch streak   |
| `BET_WIN`               | 🏆    | Prediction bet won                |
| `BET_LOSE`              | 💸    | Prediction bet lost               |
| `BET_REFUND`            | ↩️    | Prediction bet refunded           |
| `BET_FILTERS`           | 🎰    | Prediction filtered by settings   |
| `BET_GENERAL`           | 🎰    | General prediction info           |
| `BET_FAILED`            | 🎰    | Prediction bet failed             |
| `BET_START`             | 🎰    | Prediction started                |
| `BONUS_CLAIM`           | 💵    | Bonus claimed                     |
| `MOMENT_CLAIM`          | 🎉    | Community moment claimed          |
| `JOIN_RAID`             | ⚔️    | Joined a raid                     |
| `DROP_CLAIM`            | 📦    | Drop claimed                      |
| `DROP_STATUS`           | 📦    | Drop progress status              |
| `CHAT_MENTION`          | 💬    | Mentioned in chat                 |
| `GIFTED_SUB`            | 🎁    | Received a gifted sub (via IRC)   |
| `TEST`                  | —     | Test notification (see below)     |

> **Note:** Emojis are prepended to log messages and notifications automatically. The emoji mappings are defined in [`eventEmoji`](internal/logger/logger.go:19). The event type constants themselves are defined in [`internal/model/settings.go`](internal/model/settings.go:7).

**Example — send only specific events to Telegram:**

```yaml
notifications:
  telegram:
    enabled: true
    token: "..."
    chat_id: "..."
    events:
      - "GIFTED_SUB"
      - "BET_WIN"
      - "BET_LOSE"
      - "DROP_CLAIM"
```

### Notification Batching

Instead of receiving a separate notification for every single event, you can enable **batching** to group events by streamer or category and deliver them as a single message at a configurable interval.

**Global defaults with per-provider overrides:**

```yaml
notifications:
  batch:
    enabled: true
    interval: 30m          # flush buffered events every 30 minutes
    max_entries: 15         # max lines per message; splits into multiple if exceeded
    immediate_events:       # these bypass batching (empty list = batch everything)
      - "BET_WIN"
      - "BET_LOSE"
      - "DROP_CLAIM"

  discord:
    enabled: true
    batch:
      interval: 15m         # override: Discord flushes every 15 minutes
  telegram:
    enabled: true
    batch:
      enabled: false         # override: Telegram sends instantly
```

**How it works:**

- Events are grouped by their notification title (e.g. `oliwer | xQc`, `oliwer | Valorant`).
- At each interval, all buffered events for a group are joined as newline-separated lines and sent as one message.
- If a single event arrives in a batch window, it is sent as a regular (unbatched) message.
- If the buffer exceeds `max_entries`, the message is split into multiple sends.
- On graceful shutdown, all pending batched events are flushed before the process exits.
- Events listed in `immediate_events` are always sent instantly, bypassing the buffer.
- If `immediate_events` is empty or omitted, **all** events are batched.
- Per-provider `batch` config overrides the global defaults. Omitted fields inherit from global.

### Testing Notifications

The miner exposes a `POST /api/test-notification` endpoint on the analytics server to verify your notification setup. It sends a test message to **all** enabled notification providers, **bypassing event filters**.

```bash
curl -X POST http://localhost:8080/api/test-notification
```

A successful response looks like:

```json
{
  "status": "ok",
  "message": "Test notification sent to all enabled notifiers"
}
```

If some providers fail, you'll get a partial status with error details:

```json
{
  "status": "partial",
  "errors": ["telegram: 401 Unauthorized"]
}
```

> **Note:** Replace `8080` with your configured port (the `-port` flag or `PORT` env var). This endpoint is useful for verifying that tokens, chat IDs, and webhook URLs are correctly configured before relying on notifications in production.

## Authentication

Authentication is automatic — on first run the miner walks through a priority chain until one method succeeds:

| Priority | Method                                     | Description                                                                                                                             |
| -------- | ------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| 1        | **Cookie file**                            | Saved from a previous successful login. Reused automatically. If the token is expired, a **refresh token** flow is attempted first.     |
| 2        | **Auth token from config** (`auth_token`)  | Token set directly in the YAML config file — checked in [`Login()`](internal/auth/auth.go:79).                                          |
| 3        | **`TWITCH_AUTH_TOKEN_<USERNAME>` env var** | Fallback — checked directly in [`Login()`](internal/auth/auth.go:79) via `os.Getenv()`.                                                 |
| 4        | **`TWITCH_PASSWORD_<USERNAME>` env var**   | Last resort — password login, checked via `os.Getenv()` in [`Login()`](internal/auth/auth.go:79). May require 2FA.                      |
| 5        | **Device code flow (RECOMMENDED)**         | Interactive — displays a code in the terminal and waits for you to activate it at [twitch.tv/activate](https://www.twitch.tv/activate). |

Once authenticated by **any** method, the token is validated against the Twitch OAuth2 endpoint to confirm it belongs to the expected user (derived from the config filename). If there's a mismatch — for example, you completed the device code flow with the wrong Twitch account — the system will show a clear error like:

> `authenticated as "wrong_user" but config expects "your_username" — please log in with the correct account`

The validated token is then saved to a cookie file and reused on subsequent starts — so the device code flow is typically a one-time step.

> **⚠️ Warning:** Password login (step 4) may trigger Twitch's two-factor authentication (2FA) prompt, making it less reliable in fully headless environments. **Prefer `TWITCH_AUTH_TOKEN_<USERNAME>` over `TWITCH_PASSWORD_<USERNAME>` whenever possible.** Use the password method only as a last resort when you cannot obtain an OAuth token.

### When to use the env vars

The `TWITCH_AUTH_TOKEN_<USERNAME>` env var is the **recommended** fallback when the interactive device code flow is impractical:

- **Headless deployments** — servers or containers without interactive terminal access (e.g., Fly.io, cloud VMs)
- **Multi-account setups** — pre-seed tokens for several accounts without running the device flow for each
- **CI/CD environments** — automated pipelines where no human is present to complete the device flow

The variable name is `TWITCH_AUTH_TOKEN_` followed by the **uppercase** username with hyphens replaced by underscores. Examples:

| Username    | Env var                       |
| ----------- | ----------------------------- |
| `guliveer_` | `TWITCH_AUTH_TOKEN_GULIVEER_` |
| `my-user`   | `TWITCH_AUTH_TOKEN_MY_USER`   |

> **Note:** Both `TWITCH_AUTH_TOKEN_<USERNAME>` and `TWITCH_PASSWORD_<USERNAME>` are read directly in the auth flow (`os.Getenv()`), not through the config layer or `applyEnvOverrides()`.

## Docker

```bash
# Build
docker build -t twitch-miner-go .

# Run (DATA_DIR is required to persist cookies across container restarts)
docker run -d \
  -p 8080:8080 \
  -v miner_data:/data \
  -e DATA_DIR=/data \
  twitch-miner-go

# Run with auth token (recommended — for headless environments)
docker run -d \
  -p 8080:8080 \
  -v miner_data:/data \
  -e DATA_DIR=/data \
  -e TWITCH_AUTH_TOKEN_YOUR_USERNAME=your_oauth_token \
  twitch-miner-go

# Run with password (optional — last resort, may require 2FA)
docker run -d \
  -p 8080:8080 \
  -v miner_data:/data \
  -e DATA_DIR=/data \
  -e TWITCH_PASSWORD_YOUR_USERNAME=your_twitch_password \
  twitch-miner-go
```

## Deploy to Fly.io

The repo includes [`fly.toml`](fly.toml) — the Fly.io deployment config. You can customize the app name, region, and VM settings directly in this file.

### Setup

```bash
# 1. Copy the example account config and customize (filename = your Twitch username)
cp configs/example.yaml.example configs/your_twitch_username.yaml

# 2. Install flyctl
curl -L https://fly.io/install.sh | sh

# 3. Login
fly auth login

# 4. Create the app (first time only)
fly launch --no-deploy

# 5. Create a volume for persistent data
fly volumes create miner_data --region fra --size 1

# 6. (Optional) Set auth token for headless login — skips the interactive device code flow (recommended)
fly secrets set TWITCH_AUTH_TOKEN_YOUR_USERNAME=your_oauth_token

# 7. (Optional) Set password for last-resort login — less reliable than auth token, may require 2FA
fly secrets set TWITCH_PASSWORD_YOUR_USERNAME=your_twitch_password

# 8. Set notification secrets (replace YOUR_USERNAME with your Twitch username in uppercase)
fly secrets set TELEGRAM_TOKEN_YOUR_USERNAME=your_bot_token
fly secrets set TELEGRAM_CHAT_ID_YOUR_USERNAME=your_chat_id
```

### CI/CD Auto-Deploy

Pushes to `main` are automatically deployed via the [CI workflow](.github/workflows/ci.yml) after build and version bump succeed. This requires a `FLY_API_TOKEN` GitHub secret:

```bash
# 1. Generate a deploy token scoped to your app
flyctl tokens create deploy -a twitch-miner-go

# 2. Set it as a GitHub repo secret
gh secret set FLY_API_TOKEN --repo <owner>/<repo>
# (paste the token when prompted)
```

> If `FLY_API_TOKEN` is not set, the deploy step is **skipped gracefully** — build and version bump still run normally.

### Manual Deploy

> **Note:**
> _[Fly.io](https://fly.io)_ is a personal preference — therefore this project is prepared for it out of the box with a `fly.toml` and volume configuration for cookie persistence.
> However, the miner is designed to be portable and can run on any platform that supports Go. You are not limited to Fly.io — feel free to deploy on AWS, GCP, Azure, Heroku, DigitalOcean, or any other hosting provider of your choice.

```bash
fly deploy

# View logs
fly logs

# Check health
curl https://your-app-name.fly.dev/health
```

## Development

This project uses [Conventional Commits](https://www.conventionalcommits.org/) and automated versioning. See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the full commit convention, git hooks setup, and versioning workflow.

## Auto-Update Checker

On startup, the miner automatically checks for new releases in the background via [`updater.CheckForUpdate()`](internal/updater/updater.go). If a newer version is available, a notification is printed to the terminal. This check is non-blocking and does not affect startup time.

## License

This project is licensed under the GNU GPL v3.0 License. See the [LICENSE](LICENSE.txt) file for details.
