# Graph Report - twitch-miner-go  (2026-06-27)

## Corpus Check
- 128 files · ~95,024 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1544 nodes · 2733 edges · 98 communities (76 shown, 22 thin omitted)
- Extraction: 91% EXTRACTED · 9% INFERRED · 0% AMBIGUOUS · INFERRED: 252 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `024455a2`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 26|Community 26]]
- [[_COMMUNITY_Community 27|Community 27]]
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 32|Community 32]]
- [[_COMMUNITY_Community 33|Community 33]]
- [[_COMMUNITY_Community 34|Community 34]]
- [[_COMMUNITY_Community 35|Community 35]]
- [[_COMMUNITY_Community 36|Community 36]]
- [[_COMMUNITY_Community 37|Community 37]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Community 39|Community 39]]
- [[_COMMUNITY_Community 40|Community 40]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Community 44|Community 44]]
- [[_COMMUNITY_Community 45|Community 45]]
- [[_COMMUNITY_Community 46|Community 46]]
- [[_COMMUNITY_Community 47|Community 47]]
- [[_COMMUNITY_Community 48|Community 48]]
- [[_COMMUNITY_Community 49|Community 49]]
- [[_COMMUNITY_Community 50|Community 50]]
- [[_COMMUNITY_Community 51|Community 51]]
- [[_COMMUNITY_Community 52|Community 52]]
- [[_COMMUNITY_Community 53|Community 53]]
- [[_COMMUNITY_Community 54|Community 54]]
- [[_COMMUNITY_Community 55|Community 55]]
- [[_COMMUNITY_Community 56|Community 56]]
- [[_COMMUNITY_Community 57|Community 57]]
- [[_COMMUNITY_Community 58|Community 58]]
- [[_COMMUNITY_Community 59|Community 59]]
- [[_COMMUNITY_Community 60|Community 60]]
- [[_COMMUNITY_Community 61|Community 61]]
- [[_COMMUNITY_Community 62|Community 62]]
- [[_COMMUNITY_Community 63|Community 63]]
- [[_COMMUNITY_Community 64|Community 64]]
- [[_COMMUNITY_Community 65|Community 65]]
- [[_COMMUNITY_Community 66|Community 66]]
- [[_COMMUNITY_Community 67|Community 67]]
- [[_COMMUNITY_Community 68|Community 68]]
- [[_COMMUNITY_Community 69|Community 69]]
- [[_COMMUNITY_Community 70|Community 70]]
- [[_COMMUNITY_Community 71|Community 71]]
- [[_COMMUNITY_Community 72|Community 72]]
- [[_COMMUNITY_Community 73|Community 73]]
- [[_COMMUNITY_Community 74|Community 74]]
- [[_COMMUNITY_Community 75|Community 75]]
- [[_COMMUNITY_Community 76|Community 76]]
- [[_COMMUNITY_Community 77|Community 77]]
- [[_COMMUNITY_Community 78|Community 78]]
- [[_COMMUNITY_Community 79|Community 79]]
- [[_COMMUNITY_Community 80|Community 80]]
- [[_COMMUNITY_Community 81|Community 81]]
- [[_COMMUNITY_Community 82|Community 82]]
- [[_COMMUNITY_Community 83|Community 83]]
- [[_COMMUNITY_Community 84|Community 84]]
- [[_COMMUNITY_Community 87|Community 87]]
- [[_COMMUNITY_Community 88|Community 88]]
- [[_COMMUNITY_Community 90|Community 90]]
- [[_COMMUNITY_Community 91|Community 91]]
- [[_COMMUNITY_Community 96|Community 96]]

## God Nodes (most connected - your core abstractions)
1. `DefaultBetSettings()` - 41 edges
2. `T` - 37 edges
3. `makeBet()` - 36 edges
4. `Server` - 31 edges
5. `Miner` - 31 edges
6. `Connection` - 31 edges
7. `Streamer` - 30 edges
8. `Client` - 27 edges
9. `twoOutcomes()` - 27 edges
10. `Client` - 24 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `NewServer()`  [INFERRED]
  cmd/config-editor/main.go → internal/configeditor/server.go
- `main()` --calls--> `RunTUI()`  [INFERRED]
  cmd/config-editor/main.go → internal/configeditor/tui.go
- `main()` --calls--> `Bool`  [INFERRED]
  cmd/twitch-miner-go/main.go → internal/miner/miner.go
- `main()` --calls--> `getEnv()`  [INFERRED]
  cmd/twitch-miner-go/main.go → internal/config/config.go
- `main()` --calls--> `LoadAllAccountConfigs()`  [INFERRED]
  cmd/twitch-miner-go/main.go → internal/config/config.go

## Import Cycles
- None detected.

## Communities (98 total, 22 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.05
Nodes (38): Context, Miner, Streamer, Campaign, Duration, GameInfo, Tag, Time (+30 more)

### Community 1 - "Community 1"
Cohesion: 0.08
Nodes (58): addCategoryItem(), addStreamerItem(), addTag(), addTeamItem(), api(), assignNum(), assignTriToggle(), collectCategories() (+50 more)

### Community 2 - "Community 2"
Cohesion: 0.08
Nodes (29): B, BetSettings, Mutex, Streamer, Time, T, BenchmarkBetCalculate(), BenchmarkFilterConditionSkip() (+21 more)

### Community 3 - "Community 3"
Cohesion: 0.04
Nodes (47): Account Configuration, Authentication, CI/CD Pipelines, Comparison Matrix, Configuration (Docker), Configuration (Fly.io), Deployment Guide, Deployment Options (+39 more)

### Community 4 - "Community 4"
Cohesion: 0.17
Nodes (45): Bet, BetSettings, Outcome, T, DefaultBetSettings(), makeBet(), TestCalculate_AllOutcomesZeroUsers(), TestCalculate_AmountCappedByMaxPoints() (+37 more)

### Community 5 - "Community 5"
Cohesion: 0.13
Nodes (26): accountMeta, Server, cleanConfig(), isValidDuration(), mergeSecretsBack(), readJSON(), removeEmpty(), sendError() (+18 more)

### Community 6 - "Community 6"
Cohesion: 0.05
Nodes (42): 1.10. Linux Service (systemd / OpenRC), 1.11. Windows Service, 1.12.1. Setup, 1.12.2. CI/CD Auto-Deploy, 1.12.3. Manual Deploy, 1.12.4. Alternative Deployment, 1.12. Deploy to Fly.io, 1.13. Development (+34 more)

### Community 7 - "Community 7"
Cohesion: 0.07
Nodes (38): BetSettingsConfig, CategoryConfig, ResolveBatchConfig(), AuthConfig, boolPtr(), TestIsBatchEnabled_Nil(), TestResolveBatchConfig_BothNil(), TestResolveBatchConfig_GlobalOnly() (+30 more)

### Community 8 - "Community 8"
Cohesion: 0.08
Nodes (28): circuitBreaker, isRetryableGQLError(), NewClient(), NewClientForTest(), TestIsRetryableGQLError(), TestIsTransientError(), wrapTransientGQLError(), gqlError (+20 more)

### Community 9 - "Community 9"
Cohesion: 0.07
Nodes (28): Batcher, FeaturesConfig, FollowersConfig, AuthConfig, CategoryWatcherConfig, AccountConfig, NotificationsConfig, Priority (+20 more)

### Community 10 - "Community 10"
Cohesion: 0.07
Nodes (18): generateDeviceID(), NewAuthenticator(), NewForTest(), Cookie, CookieJar, CookieFileExists(), NewCookieJar(), CookieJar (+10 more)

### Community 11 - "Community 11"
Cohesion: 0.06
Nodes (22): NewManager(), Handler, NewHandler(), Manager, Client, Context, Handler, Logger (+14 more)

### Community 12 - "Community 12"
Cohesion: 0.09
Nodes (27): ConnectionSnapshot, DebugPredictionEntry, DebugWatchingEntry, Miner, Time, Context, Miner, Streamer (+19 more)

### Community 13 - "Community 13"
Cohesion: 0.09
Nodes (16): GoalContribution, ChannelPointsContext, GameResp, GoalContribution, PlaybackAccessToken, StreamInfoResponse, TeamMember, TopStream (+8 more)

### Community 14 - "Community 14"
Cohesion: 0.09
Nodes (18): Authenticator, IsTransientError(), AccountConfig, Context, Logger, Provider, RWMutex, Streamer (+10 more)

### Community 15 - "Community 15"
Cohesion: 0.12
Nodes (28): BatchConfig, batchKey, Context, Duration, Event, Logger, Mutex, Once (+20 more)

### Community 16 - "Community 16"
Cohesion: 0.11
Nodes (15): GenerateHex(), Conn, Context, Logger, Message, Mutex, Once, Provider (+7 more)

### Community 17 - "Community 17"
Cohesion: 0.12
Nodes (31): Int64, Time, T, baseNotifier, Client, Context, Event, Message (+23 more)

### Community 18 - "Community 18"
Cohesion: 0.13
Nodes (22): historyAggregate, HistoryEntry, Request, ResponseWriter, AnalyticsServer, Streamer, T, errorResponse (+14 more)

### Community 19 - "Community 19"
Cohesion: 0.14
Nodes (18): Outcome, Context, EventPrediction, Message, Miner, Streamer, BoolFromMap(), FloatFromAny() (+10 more)

### Community 20 - "Community 20"
Cohesion: 0.13
Nodes (21): Client, Context, Mutex, Request, Response, T, inventoryDrop, claimFailedResponse() (+13 more)

### Community 21 - "Community 21"
Cohesion: 0.09
Nodes (13): Connection, Context, Mutex, Provider, T, newTestConnection(), TestHandleResponse_ERR_BADAUTH_AlreadyRefreshedByAnother(), TestHandleResponse_ERR_BADAUTH_RefreshesAndResubscribes() (+5 more)

### Community 22 - "Community 22"
Cohesion: 0.13
Nodes (19): Attr, Context, Event, Handler, Mutex, Level, colorHandler, Config (+11 more)

### Community 23 - "Community 23"
Cohesion: 0.18
Nodes (28): install-service.sh script, ask(), banner(), confirm(), DEFAULT_CONFIG_DIR, DEFAULT_DATA_DIR, DEFAULT_ENV_FILE, DEFAULT_INSTALL_DIR (+20 more)

### Community 24 - "Community 24"
Cohesion: 0.11
Nodes (16): Drop, GameInfo, Time, Time, Campaign, Context, RawMessage, Streamer (+8 more)

### Community 25 - "Community 25"
Cohesion: 0.17
Nodes (11): Connection, Context, Logger, Message, Mutex, Provider, Pool, PubSubTopic (+3 more)

### Community 26 - "Community 26"
Cohesion: 0.08
Nodes (23): For /graphify add and --watch, For /graphify query, For the commit hook and native CLAUDE.md integration, For --update and --cluster-only, /graphify, Honesty Rules, Interpreter guard for subcommands, Part A - Structural extraction for code files (+15 more)

### Community 27 - "Community 27"
Cohesion: 0.08
Nodes (24): `401 Unauthorized` errors in logs, All bets on one outcome keep losing, `authenticated as "X" but config expects "Y"`, Authentication errors, Config changes not taking effect, Config issues, Drop issues, Drops not being claimed (+16 more)

### Community 28 - "Community 28"
Cohesion: 0.11
Nodes (18): API, CategoryWatcher, Dispatcher, AccountConfig, EventPrediction, Logger, Miner, Mutex (+10 more)

### Community 29 - "Community 29"
Cohesion: 0.09
Nodes (11): CommunityGoal, HistoryEntry, PointsMultiplier, RWMutex, StreamerSettings, Time, HistoryEntry, PointsMultiplier (+3 more)

### Community 30 - "Community 30"
Cohesion: 0.30
Nodes (9): CommunityGoal, Context, Event, Message, Miner, Streamer, extractNestedInt(), mapReasonToEvent() (+1 more)

### Community 31 - "Community 31"
Cohesion: 0.18
Nodes (15): Handler, Logger, RWMutex, AnalyticsServer, Streamer, Server, checkCredentials(), NewAnalyticsServer() (+7 more)

### Community 32 - "Community 32"
Cohesion: 0.11
Nodes (18): `batch` (global batching defaults), `bet.filter_condition` (optional), `bet` (nested under `streamer_defaults` and per-streamer `settings`), `blacklist`, `category_blacklist`, `category_watcher`, Configuration Reference, `features` (+10 more)

### Community 33 - "Community 33"
Cohesion: 0.29
Nodes (16): NewServer(), HandlerFunc, T, DownloadAsset(), platformAsset(), ReplaceBinary(), TestCheckForUpdate_DevVersion(), TestCheckForUpdate_NewerAvailable() (+8 more)

### Community 34 - "Community 34"
Cohesion: 0.26
Nodes (12): applyDefaults(), applyEnvOverrides(), getEnv(), LoadAccountConfig(), LoadAllAccountConfigs(), parseProxyURL(), TestApplyDefaultsSetsMaxWatchStreams(), TestValidateRejectsInvalidMaxWatchStreams() (+4 more)

### Community 35 - "Community 35"
Cohesion: 0.19
Nodes (11): main(), openBrowser(), Context, Event, T, Version, Compare(), Parse() (+3 more)

### Community 36 - "Community 36"
Cohesion: 0.27
Nodes (13): buildFilterParams(), clearFilters(), escapeHTML(), fetchJSON(), formatPoints(), handleSort(), loadFilters(), populateSelect() (+5 more)

### Community 37 - "Community 37"
Cohesion: 0.31
Nodes (13): buildFilterParams(), clearFilters(), debounce(), escapeHTML(), fetchJSON(), formatPoints(), initFilterListeners(), loadFilters() (+5 more)

### Community 38 - "Community 38"
Cohesion: 0.17
Nodes (8): StreamerSettings, StreamerSettings, BetSettings, ChatPresence, DefaultStreamerSettings(), ParseChatPresence(), ShouldJoinChat(), StreamerSettings

### Community 39 - "Community 39"
Cohesion: 0.17
Nodes (12): Discord, Event reference, Gotify, Matrix, Notification batching, Notifications, Per-account vs global credentials, Provider setup (+4 more)

### Community 40 - "Community 40"
Cohesion: 0.17
Nodes (12): Betting amount calculation, Delay modes, Filter conditions, `HIGH_ODDS`, `MOST_VOTED`, `NUMBER_1` through `NUMBER_8`, `PERCENTAGE`, Prediction Strategies (+4 more)

### Community 41 - "Community 41"
Cohesion: 0.35
Nodes (5): DeviceCodeResponse, TokenErrorResponse, TokenResponse, Authenticator, Context

### Community 42 - "Community 42"
Cohesion: 0.27
Nodes (8): Logger, T, Twitch, envOrDefault(), LoadTwitchFromEnv(), TestClientIDsForGQL_Dedup(), TestLoadTwitchFromEnv_Defaults(), TestLoadTwitchFromEnv_EnvOverride()

### Community 44 - "Community 44"
Cohesion: 0.33
Nodes (10): ghAsset, Context, ghAsset, ghRelease, UpdateInfo, CheckForUpdate(), checkWithURL(), findAssetURL() (+2 more)

### Community 45 - "Community 45"
Cohesion: 0.42
Nodes (4): promptLine(), loginResponse, Authenticator, Context

### Community 46 - "Community 46"
Cohesion: 0.28
Nodes (7): Bool, Logger, ColorSupported(), main(), playStartupAnimation(), runHealthcheck(), ExitForRestart()

### Community 47 - "Community 47"
Cohesion: 0.42
Nodes (5): Streamer, NewStreamerTopic(), NewUserTopic(), PubSubTopic, PubSubTopicType

### Community 48 - "Community 48"
Cohesion: 0.22
Nodes (8): graphify reference: extra exports and benchmark, Step 6b - Wiki (only if --wiki flag), Step 7 - Neo4j export (only if --neo4j or --neo4j-push flag), Step 7a - FalkorDB export (only if --falkordb or --falkordb-push flag), Step 7b - SVG export (only if --svg flag), Step 7c - GraphML export (only if --graphml flag), Step 7d - MCP server (only if --mcp flag), Step 8 - Token reduction benchmark (only if total_words > 5000)

### Community 49 - "Community 49"
Cohesion: 0.22
Nodes (9): 1. Clone and configure, 2. Set required environment variables, 3. Run, 4. Authenticate, 5. Verify it's working, Automatic updates, Getting Started, Next steps (+1 more)

### Community 51 - "Community 51"
Cohesion: 0.29
Nodes (6): RawMessage, MessageData, Request, RequestData, Response, RequestData

### Community 52 - "Community 52"
Cohesion: 0.29
Nodes (7): Architecture, Data flow, Key design decisions, Notification flow, Package map, Per-account miner lifecycle, Startup

### Community 53 - "Community 53"
Cohesion: 0.29
Nodes (7): Authentication, Cookie persistence, Obtaining an OAuth token, Priority chain, Recommended approach per environment, Token scopes required, Variable naming convention

### Community 54 - "Community 54"
Cohesion: 0.33
Nodes (6): Automated Versioning, Commit Convention, Contributing, Documentation and Wiki, Pull Requests, Setting Up Git Hooks

### Community 56 - "Community 56"
Cohesion: 0.40
Nodes (5): baseNotifier, Client, Context, Event, Discord

### Community 57 - "Community 57"
Cohesion: 0.40
Nodes (5): baseNotifier, Client, Context, Event, Gotify

### Community 58 - "Community 58"
Cohesion: 0.40
Nodes (5): baseNotifier, Client, Context, Event, Pushover

### Community 59 - "Community 59"
Cohesion: 0.40
Nodes (5): baseNotifier, Client, Context, Event, Telegram

### Community 60 - "Community 60"
Cohesion: 0.40
Nodes (3): Connection, Pool, ConnectionSnapshot

### Community 61 - "Community 61"
Cohesion: 0.33
Nodes (5): For /graphify explain, For /graphify path, graphify reference: query, path, explain, Step 0 — Constrained query expansion (REQUIRED before traversal), Step 1 — Traversal

### Community 62 - "Community 62"
Cohesion: 0.40
Nodes (4): Reporting a Vulnerability, Scope, Security Policy, Supported Versions

### Community 63 - "Community 63"
Cohesion: 0.50
Nodes (3): Checklist, Notes for reviewer, What does this PR do?

### Community 64 - "Community 64"
Cohesion: 0.50
Nodes (3): Context, T, Run()

### Community 65 - "Community 65"
Cohesion: 0.50
Nodes (3): For /graphify add, For --watch, graphify reference: add a URL and watch a folder

### Community 66 - "Community 66"
Cohesion: 0.50
Nodes (3): For git commit hook, For native CLAUDE.md integration, graphify reference: commit hook and native CLAUDE.md integration

### Community 67 - "Community 67"
Cohesion: 0.50
Nodes (3): For --cluster-only, For --update (incremental re-extraction), graphify reference: incremental update and cluster-only

### Community 69 - "Community 69"
Cohesion: 1.00
Nodes (3): baseNotifier, Client, Webhook

### Community 73 - "Community 73"
Cohesion: 0.67
Nodes (3): Key resources, Twitch Channel Points Miner — Go Edition, Wiki pages

## Knowledge Gaps
- **421 isolated node(s):** `edit-config.sh script`, `github.com/Guliveer/twitch-miner-go`, `DEFAULT_SERVICE_NAME`, `DEFAULT_INSTALL_DIR`, `DEFAULT_CONFIG_DIR` (+416 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **22 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `main()` connect `Community 46` to `Community 33`, `Community 34`, `Community 35`, `Community 42`, `Community 11`, `Community 44`, `Community 22`, `Community 28`, `Community 31`?**
  _High betweenness centrality (0.125) - this node is a cross-community bridge._
- **Why does `Parse()` connect `Community 35` to `Community 34`, `Community 44`, `Community 46`, `Community 17`, `Community 19`, `Community 24`?**
  _High betweenness centrality (0.101) - this node is a cross-community bridge._
- **Why does `Setup()` connect `Community 22` to `Community 5`, `Community 46`, `Community 15`, `Community 20`, `Community 21`?**
  _High betweenness centrality (0.095) - this node is a cross-community bridge._
- **Are the 39 inferred relationships involving `DefaultBetSettings()` (e.g. with `BenchmarkBetCalculate()` and `BenchmarkFilterConditionSkip()`) actually correct?**
  _`DefaultBetSettings()` has 39 INFERRED edges - model-reasoned connections that need verification._
- **What connects `edit-config.sh script`, `github.com/Guliveer/twitch-miner-go`, `DEFAULT_SERVICE_NAME` to the rest of the system?**
  _421 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.05069124423963134 - nodes in this community are weakly interconnected._
- **Should `Community 1` be split into smaller, more focused modules?**
  _Cohesion score 0.07834101382488479 - nodes in this community are weakly interconnected._