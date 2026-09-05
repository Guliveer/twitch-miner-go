# Architecture

This page describes the internal structure of `twitch-miner-go` for contributors and anyone curious about how the pieces fit together.

## Package map

```
cmd/twitch-miner-go/
└── main.go              Entry point: flag parsing, .env loading, miner startup,
                         HTTP server startup, FileWatcher or DB Poller

cmd/db-seed/
└── main.go              One-shot tool: seed PostgreSQL from existing YAML configs
                         (idempotent upsert; --dry-run flag)

internal/
├── model/               Pure data types — no external dependencies, no I/O
│   ├── settings.go      Event, Priority, Strategy constants
│   ├── prediction.go    Prediction model + strategy types
│   ├── drop.go          Drop model + progress tracking
│   ├── campaign.go      Campaign tracking
│   ├── stream.go        Stream state
│   ├── streamer.go      Streamer state
│   ├── message.go       PubSub message types
│   └── ...              (game_registry, community_goal, raid, etc.)
│
├── config/              YAML config loading + env var overrides + JSON serialisation
│
├── store/               Account persistence interface and implementations
│   ├── store.go         Store interface (ListAccounts, GetAccount, UpsertAccount, …)
│   ├── postgres.go      PostgreSQL implementation; LISTEN/NOTIFY + goose migrations
│   ├── noop.go          No-op implementation used in file mode (satisfies interface)
│   └── migrations/      Versioned SQL files (goose format, embedded in binary)
│
├── managedminer/        Dynamic miner lifecycle management
│   ├── manager.go       Manager: Start/Stop/Restart miners; exponential-backoff auto-restart
│   ├── poller.go        Poller: syncs DB → Manager (DB mode); driven by NOTIFY + ticker
│   └── filewatcher.go   FileWatcher: syncs configs/ → Manager (file mode); driven by mtime poll
│
├── auth/                OAuth2 flow, token refresh, cookie persistence
│   ├── auth.go          Priority chain (cookie → auth_token → env → device code)
│   ├── device_code.go   Interactive device code flow
│   └── cookies.go       Cookie read/write
│
├── gql/                 Twitch GraphQL client
│   ├── client.go        HTTP transport, request building, Android client ID default
│   └── operations.go    Individual GQL operation functions
│
├── pubsub/              Twitch PubSub WebSocket pool
│   ├── pool.go          Connection pool manager, topic subscription
│   └── connection.go    Single WebSocket connection lifecycle + ping/pong
│
├── twitch/              High-level Twitch API facade
│   ├── client.go        Main client struct + HTTP setup
│   ├── drops.go         Drop campaign sync + claim
│   ├── predictions.go   Prediction bet placement
│   └── minute_watched.go Spade-based minute-watched tracking
│
├── miner/               Core orchestrator — one instance per account
│   ├── miner.go         Run(), auth → PubSub → chat → watchers → main loop
│   ├── handler.go       PubSub message dispatch
│   ├── prediction_handler.go  Prediction event handling + strategy execution
│   ├── scheduler.go     Periodic tasks (drop sync, watch selection, streak)
│   └── streamer_manager.go   Streamer list management, resolver, priority sort
│
├── notify/              Notification dispatcher
│   ├── notifier.go      Dispatcher interface + event routing
│   ├── batcher.go       Event batching engine (buffer → flush on interval)
│   ├── telegram.go      Telegram provider
│   ├── discord.go       Discord webhook provider
│   ├── webhook.go       Generic HTTP webhook provider
│   ├── matrix.go        Matrix provider
│   ├── pushover.go      Pushover provider
│   └── gotify.go        Gotify provider
│
├── chat/                IRC chat manager
│   ├── chat.go          TMI connection, channel join/leave
│   └── handler.go       Mention detection + gifted sub detection
│
├── watcher/             Auto-discovery engines
│   ├── category.go      Discover streams by game category (poll loop)
│   └── team.go          Discover streams by Twitch team (poll loop)
│
├── server/              Built-in HTTP analytics server
│   ├── analytics.go     Dashboard data, route registration, HTTP Basic Auth
│   ├── handlers.go      HTTP handlers (health, streamers, stats, test-notification, etc.)
│   └── accounts_api.go  REST API for account CRUD (DB mode only; returns 501 otherwise)
│
├── runtimecfg/          Twitch client IDs and version (with env var overrides)
├── constants/           Global constants (timeouts, limits)
├── logger/              Structured logging (slog-based, color support)
├── telemetry/           Anonymous heartbeat sender
│   └── telemetry.go     Periodic HTTP POST with instance ID, version, OS, arch,
│                        deployment, and running accounts count;
│                        env-based config (TELEMETRY_AGREE, TELEMETRY_URL, etc.)
├── updater/             GitHub release version check (non-blocking, on startup)
├── version/             Version string embedding (from VERSION file + git commit)
├── workerpool/          Generic concurrent task executor
├── configeditor/        Embedded web config editor server
│   └── server.go        HTTP handlers to read/write YAML account configs via the
│                        browser at 127.0.0.1:8070; the same package powers the
│                        standalone cmd/config-editor binary
├── tray/                System tray icon (fyne.io/systray)
│   ├── tray.go          Logic + icon selection (ICO on Windows, PNG elsewhere)
│   ├── session_windows.go  Session detection: no tray in Windows session 0 (service)
│   └── session_other.go    Tray available on interactive Linux/macOS/BSD desktops
└── jsonutil/            JSON parsing helpers
```

## Data flow

### Startup

```
main.go
  ├─ Load .env
  ├─ Parse flags (config dir, port, config-editor-port, no-tray, log level, …)
  ├─ Start HTTP analytics server (server/)                     [port 8080]
  ├─ Start embedded config editor (configeditor/)              [127.0.0.1:8070]
  ├─ Start system tray (tray/) — when enabled and in an interactive session
  │    └─ left-click: dashboard; menu: Dashboard / Config Editor / Exit
  ├─ Start update checker background goroutine (updater/)
  ├─ Create Manager (managedminer/)
  └─ Choose account source:
       ├─ DB_ENABLED=true
       │    ├─ Connect to PostgreSQL (DB_DSN)
       │    ├─ Run goose migrations automatically
       │    ├─ Start DB Poller goroutine (DB_POLL_INTERVAL, default 30s)
       │    │    └─ On NOTIFY or tick → sync DB rows → Manager.Start/RestartChanged/Stop
       │    └─ Wire accounts REST API (/api/accounts)
       └─ DB_ENABLED=false (default)
            └─ Start FileWatcher goroutine (FILE_POLL_INTERVAL, default 5s)
                 └─ On mtime change → Manager.Start/RestartChanged/Stop
```

### Config editor & tray runtime

The embedded config editor and the system tray are responsible for making the
miner manageable without a terminal:

```
tray.Run()                    (desktops only)
  ├─ Show icon in tray/menu bar
  ├─ Left-click link  →  open analytics dashboard URL
  ├─ Menu "Dashboard" →  open analytics dashboard URL
  ├─ Menu "Config Editor" →  open embedded editor 127.0.0.1:8070
  └─ Menu "Exit" →  cancel the root context → graceful shutdown
```

```
configeditor server           (bound to 127.0.0.1 only)
  ├─ Serve web UI at http://127.0.0.1:8070
  ├─ List account configs from the config dir / store
  ├─ Save edits back to disk → FileWatcher detects the change
  └─ Hot-reload: the account's miner is restarted automatically
```

Two runtime gates decide whether the tray runs at all:

- **Interactive session.** On Windows, a process running as a service lives in
  session 0, which has no desktop — the tray is skipped there. Detection is in
  `tray/session_windows.go` (`ProcessIdToSessionId`). Linux/macOS/BSD are
  treated as always interactive (`session_other.go`).
- **Opt-out.** `-no-tray` flag or `NO_TRAY=true` env var disables the tray
  (containers, Fly.io).

macOS note: the tray uses `fyne.io/systray`, which requires cgo on macOS
(Objective-C/Cocoa). Released darwin binaries are built on a macOS CI runner
with `CGO_ENABLED=1`. Linux and Windows trays are pure Go — they build with
`CGO_ENABLED=0`.

### Per-account miner lifecycle

```
miner.Run()
  ├─ auth/             — authenticate, get token, validate username
  ├─ twitch/drops      — (optional) claim drops on startup
  ├─ workerpool/       — concurrent channel ID resolution for all streamers
  ├─ pubsub/pool       — subscribe to PubSub topics for all streamers
  ├─ chat/             — connect to IRC, join channels
  ├─ watcher/          — start category/team watcher goroutines
  └─ event loop
       ├─ pubsub messages  →  miner/handler.go
       │    ├─ channel-points events  →  claim bonus / watch streak
       │    ├─ prediction events      →  miner/prediction_handler.go
       │    ├─ drop progress          →  twitch/drops (claim if ready)
       │    ├─ community goals        →  contribute
       │    └─ raids                  →  follow raid
       ├─ scheduler (ticker)
       │    ├─ select streams to watch  (priority engine)
       │    ├─ send minute-watched      (twitch/minute_watched)
       │    ├─ sync drop campaigns      (twitch/drops)
       │    ├─ refresh analytics data
       │    └─ poll watchers
       └─ chat events
            ├─ mention detection  →  notify CHAT_MENTION
            └─ gifted sub        →  notify GIFTED_SUB
```

### Telemetry flow

```
main.go
  └─ LoadConfigFromEnv() → nil if TELEMETRY_AGREE=false
       └─ Sender.Start() → background goroutine
            └─ Every interval (TELEMETRY_INTERVAL, default 10 min):
                 ├─ Collect: instance_id, version, os, arch, deployment,
                 │           running_accounts
                 ├─ POST to TELEMETRY_URL/api/heartbeat with X-API-Key header
                 └─ Log debug line with payload fields
```

**Collected data is anonymous** — instance ID (random UUID), version, OS, architecture, deployment label, and running accounts count. No personal data, channel names, or IP addresses are sent. See the [telemetry dashboard](https://github.com/Guliveer/twitch-miner-go-telemetry) for the server-side collector.

**Configuration** — `TELEMETRY_AGREE=false` disables; `TELEMETRY_URL` overrides the server (for forks); `TELEMETRY_INTERVAL` controls frequency; `HEARTBEAT_API_KEY` authenticates with the server.

### Notification flow

```
Event fires in miner/
  → notify.Dispatcher.Send(event, data)
       ├─ Check event against each provider's event filter list
       ├─ Lifecycle events (STARTED/STOPPED/CRASHED) → send immediately
       ├─ immediate_events → send immediately
       └─ other events → push to batcher buffer
            └─ On interval tick (or shutdown) → flush → provider.Send()
```

## Key design decisions

**`model/` has no dependencies.** All data types live in one package with no imports from other internal packages. This prevents import cycles and makes testing straightforward.

**PubSub uses a connection pool.** Each connection can handle a limited number of topic subscriptions. The pool distributes topics across multiple connections and reconnects transparently on failure.

**Authentication falls through a priority chain.** The most persistent method (cookie) is tried first; the most interactive (device code) is tried last. This makes the miner work in both headless and interactive environments without config changes.

**Android client ID is the default for GQL requests.** Twitch's GraphQL API accepts requests from the Android client ID without requiring the stricter integrity checks applied to browser clients.

**Notification batching is provider-level.** Each provider has an independent batcher. Per-provider `batch` config overrides the global defaults, allowing Telegram to be instant while Discord is batched.

**Store interface abstracts the persistence layer.** Both `PostgresStore` and `NoopStore` implement the same `Store` interface. The `Poller` (DB mode) and `FileWatcher` (file mode) both drive the same `Manager` interface, keeping the miner core agnostic to where configs come from.

**Miners auto-restart on crash.** The `Manager.launchFn` wraps `miner.Run()` in a retry loop with exponential backoff (10 s initial, 5 min cap). A cancelled context (graceful shutdown) exits the loop immediately without retrying.

**One config-editor package serves both entry points.** `internal/configeditor` is imported by the main binary to embed the editor on `127.0.0.1:8070`, and by the standalone `cmd/config-editor` binary. This keeps the editor UI and save/validation logic in a single place, so the embedded and standalone variants never drift apart.

**The tray is a thin launcher, not a feature dependency.** `internal/tray` only opens URLs and triggers graceful shutdown through the root context. The miner core does not import `tray`, and the tray can be compiled out via `-no-tray` / `NO_TRAY` without affecting any mining logic.
