# Troubleshooting

## Quick checks

Before reporting a bug, run through this list:

- [ ] Running the [latest release](https://github.com/Guliveer/twitch-miner-go/releases/latest)?
- [ ] Logs show anything with `WARN` or `ERROR`? (run with `-log-level debug` for more detail)
- [ ] Config file is named exactly `configs/<twitch_username>.yaml` (case-sensitive)?
- [ ] No `.yaml.example` extension on the file?
- [ ] `TWITCH_CLIENT_ID_TV`, `TWITCH_CLIENT_ID_BROWSER`, and `TWITCH_CLIENT_VERSION` are set?

---

## Authentication errors

### `authenticated as "X" but config expects "Y"`

**Cause:** You completed the device code flow or provided a token for a different Twitch account than the config filename implies.

**Fix:** Delete the corresponding cookie file in `{DATA_DIR}/cookies/` and re-authenticate with the correct account, or rename the config file to match the account you authenticated as.

---

### Miner asks for device code on every startup

**Cause:** Cookies are not being persisted — either `DATA_DIR` is not set, or in Docker you're not mounting a persistent volume.

**Fix:**
```bash
# Docker: mount a named volume
docker run -v miner_data:/data -e DATA_DIR=/data ...

# Fly.io: mount the volume
# (already configured in fly.toml — ensure the volume exists)
fly volumes create miner_data --region fra --size 1
```

---

### `401 Unauthorized` errors in logs

**Cause:** The OAuth token has expired and the refresh failed, or the `TWITCH_AUTH_TOKEN_*` env var is stale.

**Fix:** Re-authenticate via the device code flow or obtain a fresh token and update the env var.

---

## Config issues

### No streamers being watched

**Cause:** The config file isn't being loaded. Common reasons:
- File has the wrong extension (`.yaml.example` instead of `.yaml` or `.yml`)
- File is in the wrong directory (check the `-config` flag value)
- `enabled: false` is set at the top level

**Fix:**
```bash
./twitch-miner-go -log-level debug -config configs
# Look for "loaded config" log lines to see which files were read
```

---

### YAML parsing error on startup

**Cause:** Syntax error in the config file — common culprits are tabs instead of spaces, missing quotes around values with special characters, or incorrect indentation.

**Fix:** Validate your YAML:
```bash
python3 -c "import yaml; yaml.safe_load(open('configs/your_user.yaml'))"
# or
npx js-yaml configs/your_user.yaml
```

Or use the visual config editor:
```bash
./_edit-config.sh   # Linux/macOS
_edit-config.bat    # Windows
```

---

### Config changes not taking effect

**Cause:** The miner does not hot-reload config files. Changes require a restart.

**Fix:** Restart the miner after editing any config file.

---

## Drop issues

### Drops not being claimed

**Cause (most common):** Drop preconditions not met — the streamer must be watched for the required duration before the drop becomes claimable.

**Cause:** `claim_drops: false` in config.

**Cause:** The streamer is not part of the active drop campaign. Verify at `https://www.twitch.tv/drops/campaigns`.

**Fix:** Enable debug logging and look for `DROP_STATUS` log lines to see current progress.

---

### `drops_only` streamers not being skipped

**Cause:** The feature skips a streamer only when **all** active campaigns it's part of are fully completed. If there's still a campaign with remaining progress, the streamer continues to be watched.

---

## Prediction issues

### Predictions not being placed

**Cause:** `make_predictions: false` in config.

**Cause:** `minimum_points` threshold not met — the account doesn't have enough points to bet.

**Cause:** `filter_condition` is blocking the bet. Look for `BET_FILTERS` in the logs.

**Cause:** The prediction closed before the configured delay elapsed.

**Fix:** Run with `-log-level debug` and look for `BET_` log lines.

---

### All bets on one outcome keep losing

**Fix:** Try a different strategy. `SMART` is the recommended default. See [Prediction Strategies](Prediction-Strategies) for a full comparison.

---

## Notification issues

### Notifications not arriving

**Cause:** Provider not enabled in config (`enabled: false`).

**Cause:** Wrong or expired credentials in env vars.

**Cause:** Event is being batched — it will arrive at the next batch flush interval.

**Fix:** Test the notification pipeline directly:
```bash
curl -X POST http://localhost:8080/api/test-notification
```
A `"status": "partial"` response will name the failing provider.

---

### Notifications arrive but are delayed

**Cause:** Notification batching is enabled. Events are buffered until the `interval` elapses.

**Fix:** Add the event to `immediate_events` to bypass batching, or reduce the `interval`, or set `batch.enabled: false` for that provider.

---

## Runtime / API issues

### GQL requests failing silently

**Cause:** Stale Twitch client IDs. The built-in defaults may become invalid when Twitch updates their clients.

**Fix:** Obtain fresh values and set them explicitly:
```dotenv
TWITCH_CLIENT_ID_TV=...
TWITCH_CLIENT_ID_BROWSER=...
TWITCH_CLIENT_VERSION=...
```
See [How to obtain Twitch runtime identifiers](https://github.com/Guliveer/twitch-miner-go#how-to-obtain-twitch-runtime-identifiers) in the README.

---

### `PubSub connection overflowed`

**Cause:** The miner is subscribed to too many topics and Twitch is sending more than ~2,000 messages per second on that connection.

**Fix:** Reduce `max_watch_streams` or the number of streamers in the config. The miner will reconnect automatically, but may miss some events during the reconnect window.

---

### Miner crashes repeatedly

**Fix:**
1. Run with `-log-level debug` to capture full logs.
2. Check for `MINER_CRASHED` notification — it includes the error details.
3. If it's a panic, please [open a bug report](https://github.com/Guliveer/twitch-miner-go/issues/new?template=bug_report.yml) with the full stack trace.

---

## Still stuck?

Open a [Configuration Help issue](https://github.com/Guliveer/twitch-miner-go/issues/new?template=config_help.yml) or start a [Discussion](https://github.com/Guliveer/twitch-miner-go/discussions).
