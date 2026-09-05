# Advanced Guide

Deep-dives into how the miner's features behave under the hood: the strategies,
the limits, the background loops and the more advanced configuration. This is
the technical companion to the [FAQ](FAQ) — where the FAQ answers "what do I
do?", this page answers "how does it actually work?".

Every number and behaviour below is taken from the code, so it reflects the real
implementation rather than the idealised feature description.

---

## Prediction strategies

<details>
<summary>Which prediction strategies exist and how does each one decide?</summary>

The miner ships with the strategies declared in `internal/model/prediction.go`:

| Strategy | Decision |
|----------|----------|
| `MOST_VOTED` | Bet on the outcome with the most voters. |
| `HIGH_ODDS` | Bet on the outcome with the highest odds. |
| `PERCENTAGE` | Bet on the outcome with the highest odds percentage. |
| `SMART_MONEY` | Bet on the outcome with the highest "top predictor" points. |
| `SMART` | Hybrid: high odds when the field is close, most voted otherwise. |
| `NUMBER_1` … `NUMBER_8` | Always bet on outcome index 0–7. |

`SMART` is the default; an unknown strategy string in the YAML silently falls
back to `SMART` (`ParseStrategy` default case). It tries to pick the higher-risk
/ higher-reward outcome only when the prediction is close, otherwise it plays
the crowd.

</details>

<details>
<summary>What do `delay` and `delay_mode` actually control?</summary>

The prediction bet is not placed instantly. `delay` (seconds) waits before the
bet, and `delay_mode` changes what the delay is measured against:

- `FROM_START` — N seconds after the prediction window **starts**.
- `FROM_END` — N seconds before the window **ends** (the default).
- `PERCENTAGE` — delay as a percentage of the window duration.

This exists so you can wait for the field to evolve (better data near the end)
before committing points. See `DelayMode` and `ParseDelayMode` in
`internal/model/prediction.go`.

</details>

<details>
<summary>Can I filter which predictions I bet on?</summary>

Yes — `bet` accepts filter conditions. Each condition targets an outcome field
(`outcomePercentage`, `totalVotes`, odds, etc.) with a comparison operator
(`Condition`) and a threshold value. A prediction that does not satisfy your
filters is skipped, and `BET_FILTERS` is logged (and optionally notified). This
is how you can, for example, only bet when the favourite has > 60% of votes, or
when the pool exceeds a minimum size. See `FilterCondition` in
`internal/model/prediction.go`.

</details>

<details>
<summary>How much do I risk per bet?</summary>

Three settings govern the stake:

- `percentage` — fraction of your current channel-points balance to stake.
- `max_points` — hard cap so a large balance never risks everything.
- `min_points` (if set) — skip the bet entirely when your balance is too low.

The actual points sent to Twitch come from the `MakePrediction` GraphQL mutation
and are validated against your balance before the bet is placed; the upstream
result is checked for a prediction error code.

</details>

---

## How watching works

<details>
<summary>How does the miner "watch" a stream — really?</summary>

It does not play the video. It sends **minute-watched events** to Twitch's
backend on a fixed cadence (`DefaultMinuteWatchedInterval`, 20 s). Each tick it
selects up to `max_watch_streams` streamers from your watched set (see
`SelectStreamersToWatch`) and sends the event for each. Ticking every 20 s keeps
the watch credit flowing even though branches can run concurrently.

</details>

<details>
<summary>Why would a streamer stop being selected even though they are online?</summary>

The selector applies your `priority` list (e.g. `STREAK`, `DROPS`, `ORDER`) and
the `max_watch_streams` cap. There is also a **freeze-detection** heuristic: a
streamer that has not received a successful minute-watched credit within
`FreezeDetectionThreshold` is treated as frozen and excluded from selection
until it recovers — this prevents the miner from burning events on an account
that is not actually counting.

</details>

<details>
<summary>How often are streamers checked for being online?</summary>

A dedicated `runMonitorLoop` checks each streamer's online status on a
**randomised** interval (20–60 s, jittered with `20+rand.IntN(40)`). The
randomisation avoids a thundering-herd of requests on a fixed cadence, which
keeps the load on Twitch's API predictable.

</details>

---

## Drops & campaigns

<details>
<summary>How often are drops synced, and what happens on each sync?</summary>

`runCampaignSync` polls on `DefaultCampaignSyncInterval` **(10 minutes)**. Each
tick it:

1. Claims everything claimable from your drop inventory first
   (`ClaimAllDropsFromInventory`).
2. Pulls the drop dashboard and campaign details.
3. Keeps only campaigns inside their time window with unclaimed drops.
4. Cross-references your inventory to mark progress/claim status.
5. Matches campaigns to streamers that have `claim_drops` enabled and
   pre-populates their campaign IDs so progress can be reported immediately.

Drops-only watching starts correctly even for streamers discovered later by the
category/team watchers.

</details>

<details>
<summary>What are "vanished", "synthetic", and "duplicate" drops?</summary>

Drop storage deals with three edge cases:

- **Vanished** — a drop that disappears from your inventory across repeated
  polls is flagged after it is missing 3 consecutive times (`detectVanishedDrops`),
  so a transient API inconsistency does not immediately trigger re-claiming.
- **Synthetic** — a drop advertised by a campaign that never actually appears
  in your inventory is tracked and eventually marked synthetic
  (`detectSyntheticDrops`, `SynthSkipPolls`), so the miner stops treating it as
  claimable.
- **Duplicate** — the same drop definition can appear across multiple campaigns
  with different instance IDs. `ClaimAllDropsFromInventory` dedups by the drop
  definition ID and claims one instance; Twitch marks the rest claimed
  server-side.

</details>

<details>
<summary>Why do some drops fail to claim with "PRECONDITIONS_NOT_MET"?</summary>

The drop needs an external account link (e.g. a game account connected to
Twitch) before it can be claimed. The miner logs a clear hint:
"Link your game account to Twitch to claim this drop". It is not a bug — the
drop is simply not claimable until the external prerequisite is satisfied.

</details>

<details>
<summary>Is there a throttle between drop claims?</summary>

Yes. Between inventory claims the miner sleeps a randomised **5–10 seconds**
(`5+rand.IntN(5)`). This is deliberate rate-limiting so a batch of claims does
not hammer the API and looks human. The `ClaimAllDropsFromInventory` function
also marks a drop as attempted *before* calling the API so a retry cannot
double-claim.

</details>

---

## PubSub (live event stream)

<details>
<summary>How many streams can I watch before the connection layer splits?</summary>

Each PubSub WebSocket connection supports **up to 50 topics**
(`MaxTopicsPerConn`). Beyond that, the pool opens additional connections up to a
maximum of **10** (`MaxPubSubConns`). With two topics per streamer
(channel-points + video-playback), one connection covers roughly 25 streamers and
the pool handles far more before you ever hit its ceiling. `HasCapacity` and the
pool choose the next connection with free capacity.

</details>

<details>
<summary>What keeps the PubSub connection alive and resurrects it?</summary>

Three mechanisms in `internal/pubsub/connection.go`:

- **PING/PONG** — a PING is sent every 4 minutes and a PONG must arrive within
  10 s; if no PONG is seen for over 5 minutes the connection is considered dead
  and torn down.
- **Server-driven reconnect** — a `RECONNECT` message causes the connection to
  close cleanly (`TypeReconnect`), and the pool opens a fresh one.
- **Duplicate suppression** — identical `Identifier` + timestamp pairs are
  dropped (`handleMessage`), so reconnects do not double-fire events.

</details>

<details>
<summary>What happens when Twitch returns ERR_BADAUTH on a subscription?</summary>

The connection refreshes its auth token and re-subscribes the failed topic
(`retryAfterRefresh`). If another connection refreshed the token first, it
detects that and still retries the subscription. Failed LISTEN errors are logged
with the topic and nonce so you can see exactly which subscription was rejected.

</details>

---

## Reliability & recovery

<details>
<summary>What happens in detail when a miner crashes?</summary>

The `Manager` restarts it with **exponential backoff**: the first retry waits
10 s, then 20 s, 40 s, … doubling up to a **5-minute cap**
(`initialRestartDelay=10s`, `maxRestartDelay=5m`). A `MINER_CRASHED` lifecycle
notification is dispatched (if you have notifications). On a clean context
cancellation the miner exits without restarting; on `ErrSkippedUnauth` it also
gives up (no point retrying headless with no credentials).

</details>

<details>
<summary>How is a "stopped" account different from a "crash" one?</summary>

- A config set to `enabled: false` (or removed) triggers an orderly `Stop` via
  the file watcher — the miner unwinds, sends `MINER_STOPPED`, and is removed
  from the manager.
- A crash is a runtime error inside `Miner.Run`. The same manager then re-runs
  it with backoff and sends `MINER_CRASHED`.

One is a deliberate lifecycle change; the other is fault recovery. Both are
first-class because the manager tracks every entry in a map rather than a static
slice.

</details>

<details>
<summary>Why does the config editor restart only my changed account on edit?</summary>

The file watcher mtime-compares each YAML against a saved watermark
(`filewatcher.go`). On a change it calls `RestartChanged(username)` — which
stops only that account's miner and starts it fresh. Other accounts keep running.
This is what makes hot-reload cheap: a re-edit of one YAML does not bounce every
miner.

</details>

---

## Data & storage

<details>
<summary>How is my cookie/token stored, and can it be encrypted?</summary>

Authentication tokens are stored in a cookie file reused on later starts. By
default they are written as plaintext. If you set `COOKIE_ENCRYPTION_KEY` (a
Base64-encoded 32-byte AES-256 key), cookie values are encrypted at rest with
AES-256-GCM and transparently decrypted on load. Existing plaintext cookies are
migrated on the first save after you enable it. Losing the key means losing
access to those cookies.

</details>

<details>
<summary>Where does the instance ID live and how can I reset it?</summary>

`.instance_id` is written once, hidden + read-only, into `DATA_DIR` (default
the working directory). The miner only reads it afterwards, which is why "read-only"
never breaks a restart. To reset it: stop the miner, delete `.instance_id` from
`DATA_DIR`, and start again — a fresh UUID is generated. See the [FAQ](FAQ) for
how to disable telemetry entirely.

</details>

<details>
<summary>What is the difference between file mode and DB mode?</summary>

- **File mode** (default) — account configs are YAML files; a `FileWatcher`
  polls the directory for changes.
- **DB mode** (`DB_ENABLED=true`) — configs are rows in PostgreSQL; a `Poller`
  reacts to changes via PostgreSQL `LISTEN/NOTIFY` (instant) plus a periodic
  ticker (default 30 s) as a fallback.

Both expose the same `Manager` lifecycle (start/restart/stop per account), so
switching storage does not change how miners are managed. In DB mode the REST
API (`/api/accounts`) is available; in file mode those endpoints return
`501 Not Implemented`.

</details>

---

## Performance notes

<details>
<summary>How much concurrency does the miner really use?</summary>

Startup resolves streamers concurrently with up to `StartupWorkers` (5) workers,
and the monitoring/refresh loops run as goroutines alongside the PubSub pool and
IRC chat manager. The result is a few OS threads (~4–5) driving many goroutines,
which is why the memory footprint stays in the tens of MB. The reference
comparison is in the [README](https://github.com/Guliveer/twitch-miner-go#13-resource-comparison).

</details>

<details>
<summary>Why is the analytics/health port separate from the config editor port?</summary>

They serve different purposes and audiences:

- `-port` (default 8080) is the health/analytics server, optionally protected by
  dashboard basic auth (`DASHBOARD_USER` / `DASHBOARD_PASSWORD_SHA256`).
- `-config-editor-port` (default 8070) is the embedded editor, bound to
  `127.0.0.1` only and *never* exposed to the network.

Keeping them apart means you can expose the dashboard behind your own auth while
the editor stays strictly local.

</details>