# Getting Started

This page walks you from a fresh clone to a running miner in under 5 minutes. It starts with the simplest explanations and gets more technical toward the end.

## What are you actually running?

Twitch gives you **channel points** as you watch a channel, and periodically runs **Drops** and **prediction events**. Watching streams manually to collect those adds up to a lot of idle time. This miner automates it: it keeps tabs open on your chosen channels in the background, collects the points, claims Drops, joins raids when they trigger, and places prediction bets according to your settings.

The important part for a first-time user: **you do not need to touch any code.** You write neither code nor configuration by hand unless you want to.

## The two easiest ways to manage your accounts

There are two built-in, graphical ways to control the miner once it is running:

- **Embedded config editor.** The miner opens a small web page at `http://localhost:8070` (only reachable from your own machine). In your browser you can edit your account settings and they take effect immediately — no restart, no command line.
- **System tray icon.** On **Windows, Linux and macOS** desktops, a small icon appears in your system tray / menu bar. Left-click opens the live dashboard. Right-click shows a menu with **Dashboard**, **Config Editor**, and **Exit**. `Exit` stops the miner gracefully.

> If the tray icon does not appear, the most common reasons are: the miner is running as a background service (a Windows service runs in "session 0", which has no desktop — the tray is intentionally skipped there), it is running in a container (Docker/Fly), or `-no-tray` / `NO_TRAY=true` is set. See [Troubleshooting](Troubleshooting).

If you would rather run without any GUI — for example on a server — everything below still works exactly the same; you just configure via the `configs/` folder and environment variables instead.

## Prerequisites

- [Go 1.25+](https://go.dev/dl/) — required for local builds
- A Twitch account
- Twitch runtime identifiers (see [Authentication](Authentication) and [How to obtain Twitch runtime identifiers](https://github.com/Guliveer/twitch-miner-go#how-to-obtain-twitch-runtime-identifiers))

## 1. Clone and configure

```bash
git clone https://github.com/Guliveer/twitch-miner-go.git
cd twitch-miner-go

# Create a config file for your account (filename = Twitch username)
cp configs/example.yaml.example configs/your_twitch_username.yaml
```

> **Note:** The repository includes the maintainer's own account configs. They are skipped automatically — you do not need to delete or disable them. Set `RUN_OWNER_ACCOUNTS=true` only if you are the maintainer running on your own infrastructure.

Open the file and set at least:

```yaml
enabled: true

streamers:
  - username: "some_streamer"
```

See [Configuration Reference](Configuration-Reference) for all options.

## 2. Set required environment variables

Create a `.env` file (copy `.env.example` as a starting point):

```dotenv
TWITCH_CLIENT_ID_TV=your_tv_client_id
TWITCH_CLIENT_ID_BROWSER=your_browser_client_id
TWITCH_CLIENT_VERSION=your_client_version
```

See [How to obtain Twitch runtime identifiers](https://github.com/Guliveer/twitch-miner-go#how-to-obtain-twitch-runtime-identifiers) in the README for step-by-step instructions.

## 3. Run

**Linux / macOS:**
```bash
./_run.sh
```

**Windows:**
```bat
_run.bat
```

Both scripts build the binary and run it immediately.

## 4. Authenticate

On first run the miner walks through the [authentication chain](Authentication). The easiest path is the **device code flow** — the miner will print:

```
To activate, visit https://www.twitch.tv/activate and enter code: ABCD-1234
```

Visit the URL, enter the code, and the miner saves a cookie for future runs — no re-auth needed.

## 5. Verify it's working

```bash
# Health check
curl http://localhost:8080/health

# Test notifications (if configured)
curl -X POST http://localhost:8080/api/test-notification
```

The analytics dashboard is available at `http://localhost:8080`.

## Telemetry

By default the miner sends anonymous usage heartbeats — instance ID, version, OS, architecture, deployment label, and running accounts count. No personal data, channel names, or IP addresses are transmitted. Set `TELEMETRY_AGREE=false` in your environment to disable.

The server-side dashboard is open source: [twitch-miner-go-telemetry](https://github.com/Guliveer/twitch-miner-go-telemetry).

## Automatic updates

Pass `-auto-update` to have the miner update itself on startup:

```bash
./_run.sh -auto-update
```

When a new release is found, the miner downloads the new binary, replaces itself, and exits — the service manager restarts it with the new version. If anything goes wrong, it falls back to printing the usual update notice.

For a **systemd** service, add the flag to `ExecStart`:
```ini
ExecStart=/usr/local/bin/twitch-miner-go -config /etc/twitch-miner-go/configs -auto-update
```

For a **Windows NSSM** service, add `-auto-update` to the service arguments (re-run `_install-service.bat` or edit via NSSM GUI).

> Not useful for Docker or Fly.io — those update by pulling a new image.

## Next steps

- Open `http://localhost:8070` while the miner runs to visually edit account settings — changes hot-reload without a restart.
- Add more streamers or enable [category/team watchers](Configuration-Reference#watcher-options)
- Configure [notifications](Notifications)
- Set up a [prediction strategy](Prediction-Strategies)
- Deploy as a persistent service — see [Docker/Fly.io/systemd/Windows service in the README](https://github.com/Guliveer/twitch-miner-go#19-docker)

## Going deeper

When you are comfortable running the miner, the pages below get progressively more detailed:

1. [Configuration Reference](Configuration-Reference) — every setting you can change.
2. [Authentication](Authentication) — how logins work and how to automate them.
3. [Prediction Strategies](Prediction-Strategies) and [Notifications](Notifications) — the two most complex feature areas.
4. [Architecture](Architecture) — how the internal packages fit together, for contributors and technical evaluators.
