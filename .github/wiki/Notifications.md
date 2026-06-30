# Notifications

The miner supports 6 notification providers. Providers are configured in the `notifications` section of your account YAML. Credentials are injected via environment variables — never hardcode tokens in config files.

## Provider setup

### Telegram

1. Create a bot via [@BotFather](https://t.me/BotFather) and copy the token.
2. Get your chat ID — send a message to the bot, then visit `https://api.telegram.org/bot<TOKEN>/getUpdates` and find `chat.id`.
3. Set env vars:
   ```dotenv
   TELEGRAM_TOKEN=123456:ABC-DEF...
   TELEGRAM_CHAT_ID=987654321
   ```
4. Enable in config:
   ```yaml
   notifications:
     telegram:
       enabled: true
       events:
         - "BET_WIN"
         - "DROP_CLAIM"
         - "MINER_CRASHED"
   ```

### Discord

1. In your Discord server: **Server Settings → Integrations → Webhooks → New Webhook**.
2. Copy the webhook URL.
3. Set env var:
   ```dotenv
   DISCORD_WEBHOOK=https://discord.com/api/webhooks/...
   ```
4. Enable in config:
   ```yaml
   notifications:
     discord:
       enabled: true
       events:
         - "BET_WIN"
         - "DROP_CLAIM"
   ```

### Webhook (generic)

Sends a POST or GET request to any URL when events fire. Useful for custom integrations.

```dotenv
WEBHOOK_URL=https://your-server.example.com/miner-events
```

```yaml
notifications:
  webhook:
    enabled: true
    method: "POST"   # GET | POST
    events:
      - "BET_WIN"
```

### Matrix

1. Create an access token for your Matrix account.
2. Find your room ID from the room settings.
3. Set env vars:
   ```dotenv
   MATRIX_HOMESERVER=https://matrix.example.com
   MATRIX_ROOM_ID=!roomid:matrix.example.com
   MATRIX_ACCESS_TOKEN=syt_...
   ```

### Pushover

1. Register an application at [pushover.net](https://pushover.net) to get an API token.
2. Copy your user key from the dashboard.
3. Set env vars:
   ```dotenv
   PUSHOVER_TOKEN=your_app_token
   PUSHOVER_USER_KEY=your_user_key
   ```

### Gotify

1. Create an application in your Gotify instance and copy the token.
2. Set env vars:
   ```dotenv
   GOTIFY_URL=https://gotify.example.com
   GOTIFY_TOKEN=your_app_token
   ```

---

## Per-account vs global credentials

All notification env vars support two scopes. Per-account takes precedence; global is the fallback.

| Global | Per-account (suffix = uppercase username) |
|--------|-------------------------------------------|
| `TELEGRAM_TOKEN` | `TELEGRAM_TOKEN_GULIVEER_` |
| `TELEGRAM_CHAT_ID` | `TELEGRAM_CHAT_ID_GULIVEER_` |
| `DISCORD_WEBHOOK` | `DISCORD_WEBHOOK_GULIVEER_` |
| `WEBHOOK_URL` | `WEBHOOK_URL_GULIVEER_` |
| `MATRIX_HOMESERVER` | `MATRIX_HOMESERVER_GULIVEER_` |
| `MATRIX_ROOM_ID` | `MATRIX_ROOM_ID_GULIVEER_` |
| `MATRIX_ACCESS_TOKEN` | `MATRIX_ACCESS_TOKEN_GULIVEER_` |
| `PUSHOVER_TOKEN` | `PUSHOVER_TOKEN_GULIVEER_` |
| `PUSHOVER_USER_KEY` | `PUSHOVER_USER_KEY_GULIVEER_` |
| `GOTIFY_URL` | `GOTIFY_URL_GULIVEER_` |
| `GOTIFY_TOKEN` | `GOTIFY_TOKEN_GULIVEER_` |

Multiple accounts can share one global Telegram channel, or each have their own.

---

## Event reference

| Event | Emoji | Description |
|-------|-------|-------------|
| `STREAMER_ONLINE` | 🟢 | Streamer goes live |
| `STREAMER_OFFLINE` | ⚫ | Streamer goes offline |
| `GAIN_FOR_RAID` | 💵 | Points gained from a raid |
| `GAIN_FOR_CLAIM` | 💵 | Points gained from claiming a bonus |
| `GAIN_FOR_WATCH` | 💵 | Points gained from watching |
| `GAIN_FOR_WATCH_STREAK` | 💵 | Points gained from a watch streak |
| `BET_WIN` | 🏆 | Prediction bet won |
| `BET_LOSE` | 💸 | Prediction bet lost |
| `BET_REFUND` | ↩️ | Prediction bet refunded |
| `BET_FILTERS` | 🎰 | Prediction skipped by filter condition |
| `BET_GENERAL` | 🎰 | General prediction information |
| `BET_FAILED` | 🎰 | Prediction bet failed to place |
| `BET_START` | 🎰 | New prediction started |
| `BONUS_CLAIM` | 💵 | Bonus chest claimed |
| `MOMENT_CLAIM` | 🎉 | Community moment claimed |
| `JOIN_RAID` | ⚔️ | Joined a raid |
| `DROP_CLAIM` | 📦 | Drop reward claimed |
| `DROP_STATUS` | 📦 | Drop campaign progress update |
| `CHAT_MENTION` | 💬 | Your account was mentioned in chat |
| `GIFTED_SUB` | 🎁 | Received a gifted subscription |
| `MINER_STARTED`           | 🚀 | Miner started (includes version) |
| `MINER_STOPPED`           | 🛑 | Miner stopped gracefully |
| `MINER_CRASHED`           | 💥 | Miner crashed (includes error); miner auto-restarts with backoff |
| `ACCOUNT_CONFIG_RELOADED` | 🔄 | Config changed in DB or YAML file and miner was restarted |
| `TEST`                    | — | Test notification (via API endpoint) |

If the `events` list is **empty or omitted**, all events are sent. Lifecycle events (`MINER_STARTED`, `MINER_STOPPED`, `MINER_CRASHED`, `ACCOUNT_CONFIG_RELOADED`) always bypass batching.

---

## Notification batching

Batching groups notifications and delivers them as a single message at a configurable interval. Reduces notification noise for high-activity streams.

```yaml
notifications:
  batch:
    enabled: true
    interval: 30m          # flush every 30 minutes
    max_entries: 15        # split into multiple messages if exceeded
    immediate_events:      # these bypass batching and send instantly
      - "BET_WIN"
      - "BET_LOSE"
      - "DROP_CLAIM"
      - "CHAT_MENTION"
      - "GIFTED_SUB"
```

Per-provider overrides:

```yaml
  discord:
    enabled: true
    batch:
      interval: 15m        # Discord: flush every 15 minutes (overrides global)

  telegram:
    enabled: true
    batch:
      enabled: false       # Telegram: send every notification instantly
```

On graceful shutdown all pending batched events are flushed before the process exits.

---

## Testing notifications

```bash
curl -X POST http://localhost:8080/api/test-notification
```

Sends a test event to all enabled providers, bypassing event filters. Check the response:

```json
{ "status": "ok", "message": "Test notification sent to all enabled notifiers" }
```

A partial failure returns:

```json
{ "status": "partial", "errors": ["telegram: 401 Unauthorized"] }
```
