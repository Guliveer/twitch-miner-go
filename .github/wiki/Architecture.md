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
├── updater/             GitHub release version check (non-blocking, on startup)
├── version/             Version string embedding (from VERSION file + git commit)
├── workerpool/          Generic concurrent task executor
└── jsonutil/            JSON parsing helpers
```

## Data flow

### Startup

```
main.go
  ├─ Load .env
  ├─ Parse flags (config dir, port, log level, --no-lifecycle-notify, …)
  ├─ Start HTTP analytics server (server/)
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
