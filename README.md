<p align="center">
  <picture>
    <img src="docs/social_preview_1.png" alt="twitch-miner-go - Efficient Auto Drops &amp; Points Claim for Twitch" width="100%">
  </picture>
</p>

# twitch-miner-go - Efficient Auto Drops & Points Claim for Twitch

[![GitHub Stars](https://img.shields.io/github/stars/Guliveer/twitch-miner-go?style=for-the-badge&logo=github&color=gold)](https://github.com/Guliveer/twitch-miner-go/stargazers)
[![CI](https://img.shields.io/github/actions/workflow/status/Guliveer/twitch-miner-go/ci.yml?style=for-the-badge&logo=github&label=CI)](https://github.com/Guliveer/twitch-miner-go/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/v/release/Guliveer/twitch-miner-go?style=for-the-badge&logo=semanticrelease&label=latest%20version)](https://github.com/Guliveer/twitch-miner-go/releases/latest)
[![Release Date](https://img.shields.io/github/release-date/Guliveer/twitch-miner-go?style=for-the-badge&logo=semanticrelease&label=latest%20release)](https://github.com/Guliveer/twitch-miner-go/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Guliveer/twitch-miner-go?style=for-the-badge&logo=go)](https://go.dev/)
[![Created At](https://img.shields.io/github/created-at/Guliveer/twitch-miner-go?style=for-the-badge&logo=github)](https://github.com/Guliveer/twitch-miner-go)
[![License](https://img.shields.io/github/license/Guliveer/twitch-miner-go?style=for-the-badge)](https://github.com/Guliveer/twitch-miner-go/blob/main/LICENSE.txt)

A high-performance Go rewrite of the [Twitch Channel Points Miner v2](https://github.com/rdavydov/Twitch-Channel-Points-Miner-v2). Mines channel points, claims bonuses, places predictions, joins raids, claims drops, and more — all with a fraction of the resource usage.

> ⭐ **[Star this repo](https://github.com/Guliveer/twitch-miner-go/stargazers)** to bookmark it and get notified about new releases — the project is young, actively developed, and the best time to jump in is now.

## 1.1. Table of Contents

- [twitch-miner-go - Efficient Auto Drops & Points Claim for Twitch](#twitch-miner-go---efficient-auto-drops--points-claim-for-twitch)
    - [1.1. Table of Contents](#11-table-of-contents)
    - [1.2. Features](#12-features)
    - [1.3. Resource Comparison](#13-resource-comparison)
    - [1.4. Running Locally](#14-running-locally)
        - [1.4.1. Flags](#141-flags)
    - [1.5. Configuration](#15-configuration)
        - [1.5.1. Quick Start](#151-quick-start)
        - [1.5.2. Config Editor](#152-config-editor)
        - [1.5.3. Database Mode (optional)](#153-database-mode-optional)
    - [1.6. Environment Variables](#16-environment-variables)
        - [1.6.1. Global](#161-global)
        - [1.6.2. Per-Account Authentication](#162-per-account-authentication)
        - [1.6.3. Notification Secrets](#163-notification-secrets)
        - [1.6.4. `.env` File Support](#164-env-file-support)
        - [1.6.4.1. How To Obtain Twitch Runtime Identifiers](#1641-how-to-obtain-twitch-runtime-identifiers)
    - [1.7. Notifications](#17-notifications)
        - [1.7.1. Supported Providers](#171-supported-providers)
        - [1.7.2. Example: Telegram](#172-example-telegram)
        - [1.7.3. Event Filtering](#173-event-filtering)
        - [1.7.4. Notification Batching](#174-notification-batching)
        - [1.7.5. Testing Notifications](#175-testing-notifications)
    - [1.8. Authentication](#18-authentication)
        - [1.8.1. When to use the env vars](#181-when-to-use-the-env-vars)
    - [1.9. Docker](#19-docker)
        - [1.9.1. Docker Compose](#191-docker-compose)
        - [1.9.2. GitHub Container Registry](#192-github-container-registry)
    - [1.10. Linux Service (systemd / OpenRC)](#110-linux-service-systemd--openrc)
        - [1.10.1. Managing the Service](#1101-managing-the-service)
        - [1.10.2. Uninstalling](#1102-uninstalling)
        - [1.10.3. Default File Locations](#1103-default-file-locations)
    - [1.11. Windows Service](#111-windows-service)
        - [1.11.1. Managing the Service](#1111-managing-the-service)
        - [1.11.2. Uninstalling](#1112-uninstalling)
    - [1.12. Deploy to Fly.io](#112-deploy-to-flyio)
        - [1.12.1. Setup](#1121-setup)
        - [1.12.2. CI/CD Auto-Deploy](#1122-cicd-auto-deploy)
        - [1.12.3. Manual Deploy](#1123-manual-deploy)
        - [1.12.4. Alternative Deployment](#1124-alternative-deployment)
    - [1.13. Development](#113-development)
    - [1.14. Auto-Update](#114-auto-update)
        - [1.14.1. Automatic updates](#1141-automatic-updates)
    - [1.15. License](#115-license)

## 1.2. Features

- 👥 **Multi-account support** — run multiple Twitch accounts from a single binary; configs can be stored as YAML files or in a PostgreSQL database (hot-reloaded without restart)
- ⛏️ **Channel points mining** — automatic minute-watched events, bonus claims, watch streaks
- 🔮 **Predictions** — configurable betting strategies (SMART, HIGH_ODDS, MOST_VOTED, etc.)
- 📦 **Drops** — automatic campaign sync and drop claiming
- ⚔️ **Raids** — automatic raid joining
- 🎉 **Community moments** — automatic moment claiming
- 🎯 **Community goals** — automatic goal contributions
- 🎁 **Gift sub detection** — notifies when your account receives a gifted subscription
- 🏷️ **Category watcher** — auto-discover streamers by game category
- 🤝 **Team watcher** — auto-discover streamers by Twitch team membership
- ⭐ **Followers mode** — automatically watch all followed channels
- 🔔 **Notifications** — Telegram, Discord, Webhook, Matrix, Pushover, Gotify
- 🚨 **Lifecycle alerts** — start, stop, and crash notifications with version info
- 📊 **Analytics dashboard** — built-in web UI for monitoring; see also [twitch-miner-go-dashboard](https://github.com/Guliveer/twitch-miner-go-dashboard) for a full management UI (DB mode)
- 🚀 **Fly.io ready** — deploy with a single command; Docker Compose and systemd service also supported

## 1.3. Resource Comparison

| Resource     | Python (per account) | Go (per account)  |
|--------------|----------------------|-------------------|
| Memory       | ~80–120 MB           | ~5–15 MB          |
| Docker image | ~800 MB              | ~10–15 MB         |
| Startup time | ~5–10 s              | < 100 ms          |
| Threads      | 60+                  | ~10–20 goroutines |

> Impressed by the difference? A [⭐ star](https://github.com/Guliveer/twitch-miner-go/stargazers) helps the next person find this instead of running the 800 MB Python image. Already using the miner? That one click keeps you in the loop for what ships next.

## 1.4. Running Locally

**Prerequisites:** [Go 1.25+](https://go.dev/dl/)

**Unix (macOS/Linux):**

```bash
./_run.sh
```

**Windows:**

```batch
_run.bat
```

**With custom flags:**

```bash
./_run.sh -config configs -port 9090 -log-level debug
```

> The scripts build the binary and run it in one step. You can also build manually with `go build -o twitch-miner-go ./cmd/twitch-miner-go`.

### 1.4.1. Flags

| Flag                     | Default   | Description                                                     |
|--------------------------|-----------|-----------------------------------------------------------------|
| `-config`                | `configs` | Path to the configuration directory                             |
| `-port`                  | `8080`    | Port for the health/analytics server                            |
| `-log-level`             | `INFO`    | Log level: DEBUG, INFO, WARN, ERROR (effective default: `INFO`) |
| `-log-format`            | `text`    | Output format: `text` (colored logfmt) or `json` (structured, Stdout and files) |
| `-log-dir`               | _(none)_  | Enable file logging; write `.log` files named with startup timestamp to this directory |
| `-healthcheck-url`       | _(none)_  | Probe the given HTTP URL and exit 0 on HTTP 200                 |
| `-version`               | `false`   | Print version and exit                                          |
| `-auto-update`           | `false`   | Download and apply the latest release automatically on startup  |
| `-no-lifecycle-notify`   | `false`   | Suppress `MINER_STARTED`, `MINER_STOPPED`, `MINER_CRASHED` notifications for this run |
| `-log-no-time`           | `false`   | Omit timestamps in console logs (useful when the platform adds its own, e.g. Fly.io)   |

## 1.5. Configuration

Create one YAML file per account in the `configs/` directory. **The filename (without extension) becomes the Twitch username** — no `username` field is needed in the YAML.

For example, to add an account for Twitch user `guliveer_`, create `configs/guliveer_.yaml`.

```bash
# Copy the example and customize for your account
cp configs/example.yaml.example configs/your_twitch_username.yaml
```

See [`configs/example.yaml.example`](configs/example.yaml.example) for the full schema. Files with a `.yaml.example` extension are not loaded as configs — only `.yaml` and `.yml` files are loaded.

> **Prefer a GUI?** Run `_edit-config.bat` (Windows) or `./_edit-config.sh` (Linux/macOS) to open a visual config editor in your browser. No additional runtimes required — the editor is a self-contained Go binary. See [Config Editor](#152-config-editor) below for details.

> **After cloning:** The repository includes the maintainer's own account configs (e.g. `guliveer_.yaml`). These are skipped automatically unless `RUN_OWNER_ACCOUNTS=true` is set — so they will not run on your machine and you do not need to delete or disable them. Just create your own config from the example template.

### 1.5.1. Quick Start

```yaml
# configs/your_twitch_username.yaml
# The filename IS the username — no username field needed.

# Set 'false' to disable this account without deleting the config (default: true)
enabled: true

features:
  claim_drops_startup: false
  enable_analytics: true

max_watch_streams: 2

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

### 1.5.2. Config Editor

A self-contained Go binary for creating, editing, and deleting account configs. Supports two modes:

- **Web GUI** (default) — opens a browser-based editor at `http://localhost:3000`
- **TUI** — interactive terminal forms, no browser required

**Quick start:**

```bash
# Windows
_edit-config.bat

# Linux / macOS
./_edit-config.sh
```

The script builds the binary on first run (requires Go), then launches it. The web editor opens automatically in your browser and reads/writes YAML files directly in `configs/`.

**Options:**

| Flag | Default | Description |
| --- | --- | --- |
| `--config` | `configs/` | Path to the config directory |
| `--port` | `3000` | Port for the web server |
| `--tui` | _(off)_ | Launch TUI mode instead of the web server |
| `--no-browser` | _(off)_ | Web mode: don't auto-open the browser |

**TUI mode** provides an interactive terminal interface — useful in headless environments or when a browser isn't available:

```bash
./_edit-config.sh --tui
# or
config-editor --tui --config /path/to/configs
```

> **Note:** In file mode the miner hot-reloads YAML changes automatically — no restart needed. See [Database Mode](#153-database-mode-optional) for the PostgreSQL-backed alternative.

### 1.5.3. Database Mode (optional)

By default the miner loads account configs from YAML files. You can optionally store them in a PostgreSQL database instead — useful when managing accounts programmatically via REST API (e.g. from a dashboard) without file system access.

**Enable DB mode:**

```dotenv
DB_ENABLED=true
DB_DSN=postgresql://user:password@host:5432/dbname?sslmode=require
```

The schema and migrations run automatically on first connection (via [goose](https://github.com/pressly/goose)).

**Migrate existing YAML configs to DB:**

```bash
go run ./cmd/db-seed              # seed from configs/ (default)
go run ./cmd/db-seed --dry-run    # preview without writing
go run ./cmd/db-seed --config /path/to/configs
```

**How it works:** The `Poller` watches the `accounts` table using PostgreSQL `LISTEN/NOTIFY` (instant) and a periodic ticker (default 30s). When a row changes, the corresponding miner is started, restarted, or stopped automatically. Changes via the REST API (`POST/PUT/DELETE /api/accounts`) are picked up within milliseconds.

**REST API endpoints** (only available when `DB_ENABLED=true`):

| Method   | Endpoint                        | Description                              |
|----------|---------------------------------|------------------------------------------|
| `GET`    | `/api/accounts`                 | List all accounts (username, enabled, updated_at, last_started_at) |
| `POST`   | `/api/accounts`                 | Create account (body: full config JSON)  |
| `GET`    | `/api/accounts/{username}`      | Get single account with full config      |
| `PUT`    | `/api/accounts/{username}`      | Update account config                    |
| `DELETE` | `/api/accounts/{username}`      | Delete account                           |

All endpoints return `501 Not Implemented` when `DB_ENABLED=false`. Auth follows the same HTTP Basic Auth as the dashboard (`DASHBOARD_USER` / `DASHBOARD_PASSWORD_SHA256`).

> **[twitch-miner-go-dashboard](https://github.com/Guliveer/twitch-miner-go-dashboard)** — a dedicated management UI for DB mode: add, edit, and disable accounts without touching YAML files or the API directly.

## 1.6. Environment Variables

Secrets and auth tokens are injected via environment variables. Per-account variables use a `_<USERNAME>` suffix (uppercase) to scope them to the correct account. Notification secrets also support a **global fallback** — if the per-account variable is not set, the miner checks for the key without the suffix (e.g. `TELEGRAM_TOKEN`), allowing all accounts to share one notification channel.

Use the `configs/` directory for per-account behavior such as watched streamers, betting strategy, and feature toggles. Use environment variables or `.env` for secrets and global runtime values that should not be duplicated per account.

For example, for user `guliveer_` the Telegram token variable is `TELEGRAM_TOKEN_GULIVEER_` and the auth token variable is `TWITCH_AUTH_TOKEN_GULIVEER_`.

> 📡 **Telemetry notice:** This project collects anonymous usage data by default — instance ID, version, OS, architecture, running account count, and total config count — to track adoption and prioritize development. No personal data, channel names, IP addresses, or identifying information is sent. **To disable, set `TELEMETRY_AGREE=false` in your environment.**  

> The instance ID is a random UUID v4 generated on first run and persisted in `DATA_DIR/.instance_id` — the same instance keeps the same ID across restarts. When running in Docker or on Fly.io, ensure `DATA_DIR` is a mounted/persistent volume so the ID survives container rebuilds.  

> Source code for the telemetry server: [github.com/Guliveer/twitch-miner-go-telemetry](https://github.com/Guliveer/twitch-miner-go-telemetry)

### 1.6.1. Global

| Variable                    | Description                                                                                   | Default          |
|-----------------------------|-----------------------------------------------------------------------------------------------|------------------|
| `LOG_LEVEL`                 | Log level (`DEBUG`, `INFO`, `WARN`, `ERROR`)                                                  | `INFO`           |
| `LOG_FORMAT`                | Output format: `text` (colored logfmt) or `json` (structured, Stdout and files)               | `text`           |
| `LOG_NO_TIME`               | Set to `true` to omit timestamps in console logs (avoids duplication when the platform adds timestamps, e.g. Fly.io) | `false`          |
| `LOG_DIR`                   | Enable file logging; directory for `.log` files named with startup timestamp                  | _(disabled)_     |
| `PORT`                      | HTTP server port for health/analytics                                                         | `8080`           |
| `DATA_DIR`                  | Persistent data directory (cookies, state)                                                    | `.`              |
| `TWITCH_CLIENT_ID_TV`       | Twitch TV client ID (falls back to built-in default if unset; override recommended)           | built-in default |
| `TWITCH_CLIENT_ID_BROWSER`  | Twitch browser client ID (falls back to built-in default if unset; override recommended)      | built-in default |
| `TWITCH_CLIENT_VERSION`     | Twitch browser client version (falls back to built-in default if unset; override recommended) | built-in default |
| `TWITCH_CLIENT_ID_MOBILE`   | Twitch mobile web client ID (falls back to built-in default if unset)                         | built-in default |
| `TWITCH_CLIENT_ID_ANDROID`  | Twitch Android client ID (falls back to built-in default if unset)                            | built-in default |
| `TWITCH_CLIENT_ID_IOS`      | Twitch iOS client ID (falls back to built-in default if unset)                                | built-in default |
| `DASHBOARD_USER`            | Username for analytics dashboard HTTP basic auth                                              | _(disabled)_     |
| `DASHBOARD_PASSWORD_SHA256` | SHA-256 hash of the dashboard password                                                        | _(none)_         |
| `RUN_OWNER_ACCOUNTS`        | Set to `true` to also run the maintainer's own account configs included in the repository     | `false`          |
| `DB_ENABLED`                | Set to `true` to use PostgreSQL-backed account store instead of YAML files                    | `false`          |
| `DB_DSN`                    | PostgreSQL connection string (required when `DB_ENABLED=true`)                                | _(unset)_        |
| `DB_POLL_INTERVAL`          | How often the DB Poller syncs when no `NOTIFY` is received (e.g. `30s`, `2m`)                 | `30s`            |
| `FILE_POLL_INTERVAL`        | How often the file watcher checks `configs/` for YAML changes (file mode only)                | `5s`             |
| `TELEMETRY_AGREE`           | Set to `false` to disable anonymous telemetry heartbeats (enabled by default)                 | `true`           |
| `TELEMETRY_URL`             | Override the compiled-in telemetry server URL (forks running their own server)                | _(compiled-in)_  |
| `HEARTBEAT_API_KEY`         | API key for the telemetry server heartbeat endpoint (`X-API-Key` header)                      | _(none)_         |
| `TELEMETRY_INTERVAL`        | How often to send heartbeats (e.g. `30m`, `1h`, `2h`)                                         | `10m`            |

> **Note:** Twitch client IDs and versions have compiled-in defaults (from `internal/constants`) that are used when the corresponding environment variables are unset. These defaults may become stale as Twitch updates their clients, so it is recommended to set these environment variables explicitly.

### 1.6.2. Per-Account Authentication

| Variable                       | Description                                         |
|--------------------------------|-----------------------------------------------------|
| `TWITCH_AUTH_TOKEN_<USERNAME>` | OAuth token (fallback for headless auth)            |
| `TWITCH_PASSWORD_<USERNAME>`   | Twitch password (last-resort auth, may require 2FA) |

### 1.6.3. Notification Secrets

Notification credentials support two levels: **global** (`KEY`) and **per-account** (`KEY_<USERNAME>`). Per-account takes precedence — if set, the global value is ignored for that account. Use global variables to send all accounts' notifications to a single channel.

| Global variable       | Per-account override             | Description           |
|-----------------------|----------------------------------|-----------------------|
| `TELEGRAM_TOKEN`      | `TELEGRAM_TOKEN_<USERNAME>`      | Telegram bot token    |
| `TELEGRAM_CHAT_ID`    | `TELEGRAM_CHAT_ID_<USERNAME>`    | Telegram chat ID      |
| `DISCORD_WEBHOOK`     | `DISCORD_WEBHOOK_<USERNAME>`     | Discord webhook URL   |
| `WEBHOOK_URL`         | `WEBHOOK_URL_<USERNAME>`         | Generic webhook URL   |
| `MATRIX_HOMESERVER`   | `MATRIX_HOMESERVER_<USERNAME>`   | Matrix homeserver URL |
| `MATRIX_ROOM_ID`      | `MATRIX_ROOM_ID_<USERNAME>`      | Matrix room ID        |
| `MATRIX_ACCESS_TOKEN` | `MATRIX_ACCESS_TOKEN_<USERNAME>` | Matrix access token   |
| `PUSHOVER_TOKEN`      | `PUSHOVER_TOKEN_<USERNAME>`      | Pushover API token    |
| `PUSHOVER_USER_KEY`   | `PUSHOVER_USER_KEY_<USERNAME>`   | Pushover user key     |
| `GOTIFY_URL`          | `GOTIFY_URL_<USERNAME>`          | Gotify server URL     |
| `GOTIFY_TOKEN`        | `GOTIFY_TOKEN_<USERNAME>`        | Gotify app token      |

### 1.6.4. `.env` File Support

The project supports loading environment variables from a `.env` file at startup using [`joho/godotenv`](https://github.com/joho/godotenv). This is **optional** — if no `.env` file is present, the app runs normally using YAML configs and/or standard environment variables.

Environment variables (whether from `.env` or the system) **override** the corresponding values from YAML config files for notification secrets only. This allows you to keep sensitive tokens out of version-controlled YAML files.

**Example `.env` file:**

```dotenv
# Global
LOG_LEVEL=DEBUG
PORT=9090

# Required Twitch runtime identifiers
TWITCH_CLIENT_ID_TV=your_tv_client_id
TWITCH_CLIENT_ID_BROWSER=your_browser_client_id
TWITCH_CLIENT_VERSION=your_client_version

# Optional Twitch client identifiers for future compatibility
TWITCH_CLIENT_ID_MOBILE=your_mobile_client_id
TWITCH_CLIENT_ID_ANDROID=your_android_client_id
TWITCH_CLIENT_ID_IOS=your_ios_client_id

# Global notification secrets (shared by all accounts)
TELEGRAM_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
TELEGRAM_CHAT_ID=987654321

# Per-account override (this account uses a different chat)
# TELEGRAM_CHAT_ID_GULIVEER_=111222333
# DISCORD_WEBHOOK_GULIVEER_=https://discord.com/api/webhooks/...
```

> See [`.env.example`](.env.example) for the starter template.

#### 1.6.4.1. How To Obtain Twitch Runtime Identifiers

The safest way to obtain these values is from real Twitch client requests that you control. Do not assume older values stay valid forever.

**Browser values**

1. Open `https://www.twitch.tv` in your browser.
2. Open DevTools and go to the `Network` tab.
3. Filter for `gql`.
4. Open a request to `https://gql.twitch.tv/gql`.
5. Copy these request headers:
    - `Client-Id` -> `TWITCH_CLIENT_ID_BROWSER`
    - `Client-Version` -> `TWITCH_CLIENT_VERSION`

**TV client ID without owning a TV**

You do not need a physical television. Practical options:

1. Use an Android TV emulator and inspect Twitch app traffic with a local proxy.
2. Use an Android phone emulator and the Twitch TV/device-login flow if you already proxy mobile traffic.
3. Reuse the TV client ID from a previously working setup and only replace it when Twitch invalidates it.

For this project today, the most important runtime values are:

- `TWITCH_CLIENT_ID_TV`
- `TWITCH_CLIENT_ID_BROWSER`
- `TWITCH_CLIENT_VERSION`

The mobile and platform-specific IDs are kept for future compatibility, but the current runtime path depends primarily on the TV and browser values.

## 1.7. Notifications

The miner supports multiple notification providers. Configure them in your account YAML file under the `notifications` key. Sensitive credentials (tokens, API keys) are injected via environment variables — see [Notification Secrets](#163-notification-secrets) above.

### 1.7.1. Supported Providers

| Provider | Config key | Required env vars                                                  |
|----------|------------|--------------------------------------------------------------------|
| Telegram | `telegram` | `TELEGRAM_TOKEN_*`, `TELEGRAM_CHAT_ID_*`                           |
| Discord  | `discord`  | `DISCORD_WEBHOOK_*`                                                |
| Gotify   | `gotify`   | `GOTIFY_URL_*`, `GOTIFY_TOKEN_*`                                   |
| Pushover | `pushover` | `PUSHOVER_TOKEN_*`, `PUSHOVER_USER_KEY_*`                          |
| Matrix   | `matrix`   | `MATRIX_HOMESERVER_*`, `MATRIX_ROOM_ID_*`, `MATRIX_ACCESS_TOKEN_*` |
| Webhook  | `webhook`  | `WEBHOOK_URL_*`                                                    |

> `*` = `_<USERNAME>` suffix (uppercase) or global (without suffix). Per-account takes precedence. For example, `TELEGRAM_TOKEN_GULIVEER_` overrides `TELEGRAM_TOKEN`. See the full variable list in [Notification Secrets](#163-notification-secrets).

### 1.7.2. Example: Telegram

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

### 1.7.3. Event Filtering

The `events` list controls which events trigger a notification for a given provider. Events are configured per notification provider in the YAML config under `notifications > {provider} > events`.

- If the `events` list is **empty or omitted**, **all** events are sent to that provider.
- If specific events are listed, **only those events** trigger notifications for that provider.

**Available events:**

| Event                   | Emoji | Description                                                                        |
|-------------------------|-------|------------------------------------------------------------------------------------|
| `STREAMER_ONLINE`       | 🟢    | Streamer goes online                                                               |
| `STREAMER_OFFLINE`      | ⚫     | Streamer goes offline                                                              |
| `GAIN_FOR_RAID`         | 💵    | Points gained from a raid                                                          |
| `GAIN_FOR_CLAIM`        | 💵    | Points gained from claiming bonus                                                  |
| `GAIN_FOR_WATCH`        | 💵    | Points gained from watching                                                        |
| `GAIN_FOR_WATCH_STREAK` | 💵    | Points gained from watch streak                                                    |
| `BET_WIN`               | 🏆    | Prediction bet won                                                                 |
| `BET_LOSE`              | 💸    | Prediction bet lost                                                                |
| `BET_REFUND`            | ↩️    | Prediction bet refunded                                                            |
| `BET_FILTERS`           | 🎰    | Prediction filtered by settings                                                    |
| `BET_GENERAL`           | 🎰    | General prediction info                                                            |
| `BET_FAILED`            | 🎰    | Prediction bet failed                                                              |
| `BET_START`             | 🎰    | Prediction started                                                                 |
| `BONUS_CLAIM`           | 💵    | Bonus claimed                                                                      |
| `MOMENT_CLAIM`          | 🎉    | Community moment claimed                                                           |
| `JOIN_RAID`             | ⚔️    | Joined a raid                                                                      |
| `DROP_CLAIM`            | 📦    | Drop claimed                                                                       |
| `DROP_CLAIM_AVAILABLE`  | 📦    | Drop available to claim (one-time notification per drop)                           |
| `DROP_STATUS`           | 📦    | Drop progress status                                                               |
| `CHAT_MENTION`          | 💬    | Mentioned in chat                                                                  |
| `GIFTED_SUB`            | 🎁    | Received a gifted sub (via IRC)                                                    |
| `MINER_STARTED`              | 🚀    | Miner started (with version info)                                                  |
| `MINER_STOPPED`              | 🛑    | Miner stopped gracefully                                                           |
| `MINER_CRASHED`              | 💥    | Miner crashed (with error details); miner auto-restarts with exponential backoff   |
| `ACCOUNT_CONFIG_RELOADED`    | 🔄    | Account config changed in DB or YAML and miner was restarted (DB mode and file mode) |
| `TEST`                       | —     | Test notification (see below)                                                      |

> **Note:** Emojis are prepended to log messages and notifications automatically. The emoji mappings are defined in [`eventEmoji`](internal/logger/logger.go:19). The event type constants are defined in [`internal/model/settings.go`](internal/model/settings.go:7).
>
> **Note:** Lifecycle events (`MINER_STARTED`, `MINER_STOPPED`, `MINER_CRASHED`, `ACCOUNT_CONFIG_RELOADED`) are always sent immediately — they bypass notification batching entirely.
>
> **Note:** Use `--no-lifecycle-notify` flag to suppress `MINER_STARTED`, `MINER_STOPPED`, and `MINER_CRASHED` for a single run — useful when restarting the process frequently (e.g. during deploys).

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

### 1.7.4. Notification Batching

Instead of receiving a separate notification for every single event, you can enable **batching** to group events by streamer or category and deliver them as a single message at a configurable interval.

**Global defaults with per-provider overrides:**

```yaml
notifications:
  batch:
    enabled: true
    interval: 30m # flush buffered events every 30 minutes
    max_entries: 15 # max lines per message; splits into multiple if exceeded
    immediate_events: # these bypass batching (empty list = batch everything)
      - "BET_WIN"
      - "BET_LOSE"
      - "DROP_CLAIM"

  discord:
    enabled: true
    batch:
      interval: 15m # override: Discord flushes every 15 minutes
  telegram:
    enabled: true
    batch:
      enabled: false # override: Telegram sends instantly
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

### 1.7.5. Testing Notifications

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

## 1.8. Authentication

Authentication is automatic — on first run the miner walks through a priority chain until one method succeeds:

| Priority | Method                        | Description                                                                                                                                       |
|----------|-------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------|
| 1        | **Cookie file**               | Saved from a previous successful login. Reused automatically. If expired, a refresh token flow is attempted first.                                |
| 2        | **Auth token from config**    | Token set directly in the YAML config file (`auth_token` field).                                                                                  |
| 3        | **`TWITCH_AUTH_TOKEN_*` env** | Fallback — read via `os.Getenv()`.                                                                                                                |
| 4        | **`TWITCH_PASSWORD_*` env**   | Last resort — password login via `os.Getenv()`. May require 2FA.                                                                                  |
| 5        | **Device code flow**          | Interactive — displays a code in the terminal and waits for activation at [twitch.tv/activate](https://www.twitch.tv/activate). **(RECOMMENDED)** |

Once authenticated, the token is validated against the Twitch OAuth2 endpoint to confirm it belongs to the expected user (derived from the config filename). If there's a mismatch — for example, you completed the device code flow with the wrong Twitch account — the system will show a clear error like:

> `authenticated as "wrong_user" but config expects "your_username" — please log in with the correct account`

The validated token is then saved to a cookie file and reused on subsequent starts — so the device code flow is typically a one-time step.

> See [`internal/auth/auth.go`](internal/auth/auth.go) for the full authentication implementation.

> **Warning:** Password login (priority 4) may trigger Twitch's two-factor authentication (2FA) prompt, making it less reliable in fully headless environments. **Prefer `TWITCH_AUTH_TOKEN_<USERNAME>` over `TWITCH_PASSWORD_<USERNAME>` whenever possible.** Use the password method only as a last resort when you cannot obtain an OAuth token.

### 1.8.1. When to use the env vars

The `TWITCH_AUTH_TOKEN_<USERNAME>` env var is the **recommended** fallback when the interactive device code flow is impractical:

- **Headless deployments** — servers or containers without interactive terminal access (e.g., Fly.io, Docker hosts, cloud VMs)
- **Multi-account setups** — pre-seed tokens for several accounts without running the device flow for each
- **CI/CD environments** — automated pipelines where no human is present to complete the device flow

The variable name is `TWITCH_AUTH_TOKEN_` followed by the **uppercase** username with hyphens replaced by underscores. Examples:

| Username    | Env var                       |
|-------------|-------------------------------|
| `guliveer_` | `TWITCH_AUTH_TOKEN_GULIVEER_` |
| `my-user`   | `TWITCH_AUTH_TOKEN_MY_USER`   |

> **Note:** Both `TWITCH_AUTH_TOKEN_<USERNAME>` and `TWITCH_PASSWORD_<USERNAME>` are read directly in the auth flow (`os.Getenv()`), not through the config layer or `applyEnvOverrides()`.

## 1.9. Docker

```bash
# Build
docker build -t twitch-miner-go .

# Run (DATA_DIR is required to persist cookies across container restarts)
docker run -d \
  -p 8080:8080 \
  -v miner_data:/data \
  -v $(pwd)/configs:/configs:ro \
  -e DATA_DIR=/data \
  --env-file .env \
  twitch-miner-go

# Run with auth token (recommended — for headless environments)
docker run -d \
  -p 8080:8080 \
  -v miner_data:/data \
  -v $(pwd)/configs:/configs:ro \
  -e DATA_DIR=/data \
  --env-file .env \
  -e TWITCH_AUTH_TOKEN_YOUR_USERNAME=your_oauth_token \
  twitch-miner-go

# Run with password (optional — last resort, may require 2FA)
docker run -d \
  -p 8080:8080 \
  -v miner_data:/data \
  -v $(pwd)/configs:/configs:ro \
  -e DATA_DIR=/data \
  --env-file .env \
  -e TWITCH_PASSWORD_YOUR_USERNAME=your_twitch_password \
  twitch-miner-go
```

### 1.9.1. Docker Compose

```bash
docker compose up -d
```

The included [`docker-compose.yml`](docker-compose.yml) uses the published GHCR image by default:

```text
ghcr.io/guliveer/twitch-miner-go:latest
```

Set `TWITCH_MINER_IMAGE` in `.env` if you want to pin a version or use a different registry/tag.

The compose setup mounts:

- `./configs` to `/configs` as read-only account configuration
- a named volume to `/data` for cookies and persisted session state
- `.env` for required Twitch client identifiers, dashboard auth, and account secrets

### 1.9.2. GitHub Container Registry

This repository publishes Docker images to GHCR with GitHub Actions.

- `main` pushes publish `latest` and a short SHA tag
- version tags such as `v1.2.3` publish `1.2.3` and `1.2`

Example image references:

- `ghcr.io/guliveer/twitch-miner-go:latest`
- `ghcr.io/guliveer/twitch-miner-go:1.2.3`
- `ghcr.io/guliveer/twitch-miner-go:sha-abcdef1`

## 1.10. Linux Service (systemd / OpenRC)

Run twitch-miner-go as a native Linux service with automatic restarts and boot startup. The interactive installer auto-detects the init system (systemd or OpenRC/Alpine).

```bash
# Build the binary first
./_run.sh   # Ctrl+C after build completes

# Run the installer wizard
sudo ./_install-service.sh install
```

The wizard will prompt for service name, paths, port, user, and optionally enable + start the service.

### 1.10.1. Managing the Service

```bash
# systemd
systemctl status twitch-miner-go
systemctl restart twitch-miner-go
journalctl -u twitch-miner-go -f     # follow logs

# OpenRC (Alpine)
rc-service twitch-miner-go status
rc-service twitch-miner-go restart
tail -f /var/log/twitch-miner-go.log  # follow logs
```

### 1.10.2. Uninstalling

```bash
sudo ./_install-service.sh uninstall
```

### 1.10.3. Default File Locations

| Item           | Path                             |
|----------------|----------------------------------|
| Binary         | `/usr/local/bin/twitch-miner-go` |
| Configs        | `/etc/twitch-miner-go/configs/`  |
| Environment    | `/etc/twitch-miner-go/.env`      |
| Data (cookies) | `/var/lib/twitch-miner-go/`      |

> See [DEPLOYMENT.md](DEPLOYMENT.md) for the full Linux service deployment guide.

## 1.11. Windows Service

Run twitch-miner-go as a Windows service with automatic restarts. The binary is rebuilt from source on every start, so config and code changes are always picked up. Uses [NSSM](https://nssm.cc/) (auto-downloaded if not installed).

```bat
REM Right-click and select "Run as administrator"
_install-service.bat install
```

The wizard will prompt for config directory, port, and log level.

### 1.11.1. Managing the Service

```bat
_install-service.bat start       REM starts (rebuilds the binary first)
_install-service.bat stop
_install-service.bat restart     REM restart with a fresh rebuild
_install-service.bat status
```

### 1.11.2. Uninstalling

```bat
_install-service.bat uninstall
```

> See [DEPLOYMENT.md](DEPLOYMENT.md) for the full Windows service deployment guide.

## 1.12. Deploy to Fly.io

The repo includes [`fly.toml`](fly.toml) — the Fly.io deployment config. Fly.io is a personal preference and comes pre-configured, but the miner is portable and runs on any platform that supports Go (AWS, GCP, Azure, DigitalOcean, etc.).

### 1.12.1. Setup

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

# 6. Set required Twitch runtime identifiers
fly secrets set TWITCH_CLIENT_ID_TV=your_tv_client_id
fly secrets set TWITCH_CLIENT_ID_BROWSER=your_browser_client_id
fly secrets set TWITCH_CLIENT_VERSION=your_client_version

# 7. (Optional) Set auth token for headless login — skips the interactive device code flow (recommended)
fly secrets set TWITCH_AUTH_TOKEN_YOUR_USERNAME=your_oauth_token

# 8. (Optional) Set password for last-resort login — less reliable than auth token, may require 2FA
fly secrets set TWITCH_PASSWORD_YOUR_USERNAME=your_twitch_password

# 9. Set notification secrets (replace YOUR_USERNAME with your Twitch username in uppercase)
fly secrets set TELEGRAM_TOKEN_YOUR_USERNAME=your_bot_token
fly secrets set TELEGRAM_CHAT_ID_YOUR_USERNAME=your_chat_id
```

### 1.12.2. CI/CD Auto-Deploy

Pushes to `main` are automatically deployed via the [CI workflow](.github/workflows/ci.yml) after build and version bump succeed. This requires a `FLY_API_TOKEN` GitHub secret:

```bash
# 1. Generate a deploy token scoped to your app
flyctl tokens create deploy -a twitch-miner-go

# 2. Set it as a GitHub repo secret
gh secret set FLY_API_TOKEN --repo <owner>/<repo>
# (paste the token when prompted)
```

> If `FLY_API_TOKEN` is not set, the deployment step is **skipped gracefully** — build and version bump still run normally.

### 1.12.3. Manual Deploy

```bash
fly deploy

# View logs
fly logs

# Check health
curl https://your-app-name.fly.dev/health
```

### 1.12.4. Alternative Deployment

For self-hosted deployments, Docker Compose is also supported — see the [Docker Compose](#191-docker-compose) section above and [DEPLOYMENT.md](DEPLOYMENT.md) for a comprehensive guide covering both Fly.io and Docker workflows.

## 1.13. Development

This project uses [Conventional Commits](https://www.conventionalcommits.org/) and automated versioning. See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the full commit convention, git hooks setup, and versioning workflow.

## 1.14. Auto-Update

On startup, the miner automatically checks for new releases in the background. If a newer version is available, a notification is printed to the terminal. This check is non-blocking and does not affect startup time.

### 1.14.1. Automatic updates

Pass `-auto-update` to have the miner download and apply the update itself:

```bash
./_run.sh -auto-update
```

When a new release is detected:
1. The platform-specific binary is downloaded from [GitHub Releases](https://github.com/Guliveer/twitch-miner-go/releases).
2. The current binary is replaced atomically.
3. The process exits with code 0 — your service manager (systemd, NSSM, OpenRC) restarts it automatically with the new binary.

If the download or replacement fails (e.g. no write permission, Docker read-only filesystem), the miner continues running and prints the usual update notification instead.

**To enable in a systemd unit**, add `-auto-update` to `ExecStart`:
```ini
ExecStart=/usr/local/bin/twitch-miner-go -config /etc/twitch-miner-go/configs -auto-update
```

**To enable in a Windows NSSM service**, re-run the installer or edit the service arguments in NSSM GUI to include `-auto-update`.

> **Note:** Auto-update is opt-in and disabled by default. It is not useful for Docker or Fly.io deployments where the image is the unit of update.

## 1.15. License

This project is licensed under the GNU GPL v3.0 License. See the [LICENSE](LICENSE.txt) file for details.
