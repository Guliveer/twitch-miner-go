# Deployment Guide

This project supports multiple deployment strategies with modular CI/CD pipelines. Choose the approach that fits your infrastructure.

## Table of Contents

1. [Deployment Options](#deployment-options)
2. [Linux Service (systemd / OpenRC)](#linux-service-systemd--openrc)
3. [Windows Service](#windows-service)
4. [Docker Compose (GHCR)](#docker-compose-ghcr)
5. [Fly.io](#flyio)
6. [CI/CD Pipelines](#cicd-pipelines)
7. [Required Configuration](#required-configuration)

---

## Deployment Options

| Method | Best For | Pros | Cons |
|--------|----------|------|------|
| **Linux Service** | Bare-metal Linux, VPS, Alpine | Native, zero overhead, auto-restart, systemd + OpenRC | Linux only, manual updates |
| **Windows Service** | Windows desktop/server | Native, auto-rebuild on start, no visible window | Windows only, requires Go toolchain |
| **Docker Compose** | Self-hosted, VPS | Full control, customizable | Requires infrastructure management |
| **Fly.io** | Quick cloud deploy | Managed platform, auto-scaling | Costs for high usage |

All methods work with the same codebase and configuration files. Choose based on your infrastructure preferences.

---

## Linux Service (systemd / OpenRC)

The installer auto-detects the init system — systemd (most distros) or OpenRC (Alpine Linux). OpenRC support requires `supervise-daemon` (OpenRC >= 0.21).

### Quick Start

```bash
# 1. Clone and build
git clone https://github.com/Guliveer/twitch-miner-go.git
cd twitch-miner-go
./run.sh  # builds the binary (Ctrl+C to stop after build)

# 2. Run the interactive installer
sudo ./install-service.sh install
```

The wizard will ask for service name, paths, port, user, and optionally enable + start the service.

### Managing the Service

```bash
# Check status and recent logs (works with both init systems)
sudo ./install-service.sh status

# systemd
systemctl status twitch-miner-go
systemctl restart twitch-miner-go
journalctl -u twitch-miner-go -f

# OpenRC (Alpine)
rc-service twitch-miner-go status
rc-service twitch-miner-go restart
tail -f /var/log/twitch-miner-go.log
```

### Uninstalling

```bash
sudo ./install-service.sh uninstall
```

This stops and removes the service unit/init script. Binary, configs, and data are preserved — remove them manually if needed.

### File Locations (defaults)

| Item | Path |
|------|------|
| Binary | `/usr/local/bin/twitch-miner-go` |
| Configs | `/etc/twitch-miner-go/configs/` |
| Environment | `/etc/twitch-miner-go/.env` |
| Data (cookies) | `/var/lib/twitch-miner-go/` |
| Service (systemd) | `/etc/systemd/system/twitch-miner-go.service` |
| Service (OpenRC) | `/etc/init.d/twitch-miner-go` |
| Logs (OpenRC) | `/var/log/twitch-miner-go.log` |

---

## Windows Service

### Quick Start

```bat
REM Right-click install-service.bat and select "Run as administrator"
install-service.bat install
```

The wizard will ask for config directory, port, and log level. [NSSM](https://nssm.cc/) (Non-Sucking Service Manager) is auto-downloaded if not already installed.

### How It Works

The installer uses NSSM to wrap the application as a native Windows service. On every start (including automatic restarts after a crash), the service:

1. Rebuilds the binary from source (`go build`)
2. Launches the freshly built binary

This ensures config changes and code updates are always picked up without manual rebuilds.

### Managing the Service

```bat
install-service.bat start       REM starts (triggers a rebuild first)
install-service.bat stop
install-service.bat restart     REM restart with a fresh rebuild
install-service.bat status
```

### Uninstalling

```bat
install-service.bat uninstall
```

This stops and removes the Windows service. The binary, configs, and logs are preserved — remove them manually if needed.

### File Locations (defaults)

| Item | Path |
|------|------|
| Binary | `<project-dir>\twitch-miner-go.exe` |
| Configs | `<project-dir>\configs\` |
| Service wrapper | `<project-dir>\_service-start.bat` |
| Logs | `<project-dir>\logs\service.log` |
| NSSM | `<project-dir>\tools\nssm\nssm.exe` |

### Prerequisites

- **Go toolchain** in the system PATH (not just user PATH — the service runs under the SYSTEM account)
- **Git** in the system PATH (for version embedding during builds)
- **Administrator privileges** to install/manage Windows services

---

## Docker Compose (GHCR)

### Quick Start (Docker)

```bash
# 1. Clone and configure
git clone https://github.com/Guliveer/twitch-miner-go.git
cd twitch-miner-go
cp .env.example .env
cp configs/example.yaml.example configs/your_username.yaml

# 2. Edit .env with required Twitch identifiers
nano .env  # Set TWITCH_CLIENT_ID_TV, TWITCH_CLIENT_ID_BROWSER, TWITCH_CLIENT_VERSION

# 3. Configure your account
nano configs/your_username.yaml

# 4. Start with Docker Compose
docker compose up -d

# 5. Check logs
docker compose logs -f
```

### Configuration (Docker)

The `docker-compose.yml` uses the published GHCR image by default:

```yaml
services:
  twitch-miner-go:
    image: ${TWITCH_MINER_IMAGE:-ghcr.io/guliveer/twitch-miner-go:latest}
```

#### Image Versions

Set `TWITCH_MINER_IMAGE` in `.env` to pin a specific version:

```bash
# Use latest (rolling)
TWITCH_MINER_IMAGE=ghcr.io/guliveer/twitch-miner-go:latest

# Use specific version
TWITCH_MINER_IMAGE=ghcr.io/guliveer/twitch-miner-go:1.6.0

# Use commit SHA
TWITCH_MINER_IMAGE=ghcr.io/guliveer/twitch-miner-go:sha-abc1234
```

#### Volume Mounts

- `./configs:/configs:ro` — Account configuration (read-only)
- `twitch-miner-data:/data` — Persistent cookies and state

#### Health Checks

The container includes automatic health checks:

```yaml
healthcheck:
  test: ["CMD", "/twitch-miner-go", "-healthcheck-url", "http://127.0.0.1:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 15s
```

#### Updating

```bash
# Pull latest image
docker compose pull

# Restart with new image
docker compose up -d

# Clean up old images
docker image prune -f
```

---

## Fly.io

### Quick Start (Fly.io)

```bash
# 1. Install flyctl
curl -L https://fly.io/install.sh | sh

# 2. Login
fly auth login

# 3. Clone and configure
git clone https://github.com/Guliveer/twitch-miner-go.git
cd twitch-miner-go
cp configs/example.yaml.example configs/your_username.yaml

# 4. Configure your account
nano configs/your_username.yaml

# 5. Create app (first time only)
fly launch --no-deploy

# 6. Create persistent volume
fly volumes create miner_data --region fra --size 1

# 7. Set required Twitch runtime identifiers
fly secrets set TWITCH_CLIENT_ID_TV=your_tv_client_id
fly secrets set TWITCH_CLIENT_ID_BROWSER=your_browser_client_id
fly secrets set TWITCH_CLIENT_VERSION=your_client_version

# 8. (Optional) Set auth token for headless login
fly secrets set TWITCH_AUTH_TOKEN_YOUR_USERNAME=your_oauth_token

# 9. Deploy
fly deploy
```

### Configuration (Fly.io)

The `fly.toml` includes optimized settings for the miner:

```toml
[env]
  PORT = '8080'
  DATA_DIR = '/data'
  GOMEMLIMIT = "400MiB"  # Memory limit for Go GC
  GOGC = "50"            # Aggressive GC for low-memory environment

[[vm]]
  cpu_kind = 'shared'
  cpus = 1
  memory_mb = 256        # Minimal footprint
```

#### Secrets Management

Fly.io uses encrypted secrets instead of `.env` files:

```bash
# Set secrets
fly secrets set KEY=VALUE

# List secrets (values hidden)
fly secrets list

# Remove secret
fly secrets unset KEY
```

#### Scaling

```bash
# Scale to multiple regions
fly scale count 2 --region fra,ams

# Change VM size
fly scale vm shared-cpu-1x --memory 512

# Auto-stop/start (cost optimization)
fly scale count 0  # Stop
fly scale count 1  # Start
```

#### Monitoring

```bash
# View logs
fly logs

# SSH into instance
fly ssh console

# Check status
fly status

# Open dashboard
fly open
```

---

## CI/CD Pipelines

The repository includes modular GitHub Actions workflows:

### Workflow Overview

| Workflow | Trigger | Purpose | Auto-runs |
|----------|---------|---------|-----------|
| **CI** ([ci.yml](.github/workflows/ci.yml)) | Push, PR | Test, lint, version bump | Always |
| **Docker Publish** ([docker-publish.yml](.github/workflows/docker-publish.yml)) | Push to main, tags | Build and push to GHCR | When enabled |
| **Fly.io Deploy** ([fly-deploy.yml](.github/workflows/fly-deploy.yml)) | Push to main, tags | Deploy to Fly.io | When `FLY_API_TOKEN` exists |

### Enabling Workflows

#### Docker Publish (GHCR)

**Auto-enabled** for public repositories. For private repos:

1. Go to repository Settings → Actions → General
2. Enable "Read and write permissions" for `GITHUB_TOKEN`
3. Workflow will run on next push to `main`

No additional configuration needed!

#### Fly.io Deploy

**Enabled when `FLY_API_TOKEN` secret is configured:**

1. Get your Fly.io API token:

   ```bash
   fly auth token
   ```

2. Add to GitHub:
   - Go to repository Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Name: `FLY_API_TOKEN`
   - Value: (paste your token)

3. Workflow will auto-run on next push to `main`

### Manual Triggers

Both workflows support manual dispatch:

```bash
# From GitHub UI: Actions → Select workflow → Run workflow

# Or via GitHub CLI:
gh workflow run docker-publish.yml
gh workflow run fly-deploy.yml
```

### Disabling Workflows

**To disable a workflow without deleting it:**

```bash
# Method 1: Remove trigger secret
# For Fly.io: delete FLY_API_TOKEN secret

# Method 2: Edit workflow file
# Change 'on:' section to only 'workflow_dispatch:'
on:
  workflow_dispatch:
```

---

## Required Configuration

### Twitch Runtime Identifiers

**Required for ALL deployment methods:**

| Variable | Description | How to Obtain |
|----------|-------------|---------------|
| `TWITCH_CLIENT_ID_TV` | TV app client ID | Inspect Twitch TV app traffic |
| `TWITCH_CLIENT_ID_BROWSER` | Browser client ID | DevTools Network tab → `gql` request header `Client-Id` |
| `TWITCH_CLIENT_VERSION` | Browser version | DevTools Network tab → `gql` request header `Client-Version` |

#### Getting Browser Values

1. Open <https://www.twitch.tv> in your browser
2. Open DevTools (F12) → Network tab
3. Filter for `gql`
4. Click any request to `https://gql.twitch.tv/gql`
5. Find Request Headers:
   - `Client-Id` → `TWITCH_CLIENT_ID_BROWSER`
   - `Client-Version` → `TWITCH_CLIENT_VERSION`

#### Getting TV Client ID

**Option 1:** Android TV emulator with traffic proxy
**Option 2:** Reuse existing TV client ID until Twitch invalidates it
**Option 3:** Check community resources for current IDs

### Account Configuration

Create one YAML file per account in `configs/`. The filename (without extension) becomes the username — no `username` field in the YAML:

```yaml
# configs/your_twitch_username.yaml
enabled: true

features:
  claim_drops_startup: false
  enable_analytics: true

max_watch_streams: 2

priority:
  - STREAK
  - DROPS
  - ORDER

streamers:
  - username: "streamer_name"
    settings:
      make_predictions: true
```

### Authentication

**Recommended:** Use auth tokens for headless environments

```bash
# Docker Compose (.env)
TWITCH_AUTH_TOKEN_YOUR_USERNAME=your_oauth_token

# Fly.io (secrets)
fly secrets set TWITCH_AUTH_TOKEN_YOUR_USERNAME=your_oauth_token
```

**Alternative:** Password (less reliable, may require 2FA)

```bash
TWITCH_PASSWORD_YOUR_USERNAME=your_password
```

**Interactive:** Device code flow (local development only)

---

## Comparison Matrix

| Feature | Docker Compose | Fly.io |
|---------|----------------|--------|
| **Cost** | Infrastructure cost only | ~$2-5/month (256MB VM) |
| **Setup Time** | 5 minutes | 5 minutes |
| **Maintenance** | Manual updates | Auto-deploy via CI/CD |
| **Scaling** | Manual | Automatic |
| **Monitoring** | External tools | Built-in dashboard |
| **Secrets** | `.env` file | Encrypted secrets |
| **Backup** | Volume backups | Volume snapshots |
| **Migration** | Portable containers | Vendor lock-in |

---

## Troubleshooting

### Troubleshooting: Docker Compose

**Container won't start:**

```bash
docker compose logs
# Check for missing Twitch identifiers
```

**Config changes not applied:**

```bash
docker compose down
docker compose up -d
```

**Old cookies causing issues:**

```bash
docker compose down
docker volume rm twitch-miner-go_twitch-miner-data
docker compose up -d
```

### Troubleshooting: Fly.io

**Deployment fails:**

```bash
fly logs
# Check for missing secrets
fly secrets list
```

**Out of memory:**

```bash
# Increase memory
fly scale vm shared-cpu-1x --memory 512
```

**Volume issues:**

```bash
fly volumes list
fly volumes destroy <volume_id>
fly volumes create miner_data --region fra --size 1
```

---

## Security Best Practices

1. **Never commit secrets** — use `.env` (gitignored) or secret managers
2. **Use auth tokens** instead of passwords when possible
3. **Enable dashboard auth** — set `DASHBOARD_USER` and `DASHBOARD_PASSWORD_SHA256`
4. **Rotate tokens** periodically
5. **Monitor logs** for authentication failures
6. **Use HTTPS** for external access (Fly.io includes this automatically)

---

## Next Steps

- **Configure notifications:** See [README.md](README.md#17-notifications)
- **Customize betting:** See [configs/example.yaml.example](configs/example.yaml.example)
- **Monitor performance:** Access dashboard at `http://localhost:8080` (Docker) or your Fly.io URL
- **Join community:** Check GitHub Discussions for tips and support
