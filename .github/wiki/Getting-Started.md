# Getting Started

This page walks you from a fresh clone to a running miner in under 5 minutes.

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

- Add more streamers or enable [category/team watchers](Configuration-Reference#watcher-options)
- Configure [notifications](Notifications)
- Set up a [prediction strategy](Prediction-Strategies)
- Deploy as a persistent service — see [Docker/Fly.io/systemd/Windows service in the README](https://github.com/Guliveer/twitch-miner-go#19-docker)
