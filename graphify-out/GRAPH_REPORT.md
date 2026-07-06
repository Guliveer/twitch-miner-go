# Graph Report - twitch-miner-go  (2026-07-06)

## Corpus Check
- 153 files · ~135,363 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1912 nodes · 3553 edges · 113 communities (89 shown, 24 thin omitted)
- Extraction: 91% EXTRACTED · 9% INFERRED · 0% AMBIGUOUS · INFERRED: 330 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `c1faed1f`
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
- [[_COMMUNITY_Community 98|Community 98]]
- [[_COMMUNITY_Community 99|Community 99]]
- [[_COMMUNITY_Community 100|Community 100]]
- [[_COMMUNITY_Community 101|Community 101]]
- [[_COMMUNITY_Community 103|Community 103]]
- [[_COMMUNITY_Community 104|Community 104]]
- [[_COMMUNITY_Community 106|Community 106]]
- [[_COMMUNITY_Community 107|Community 107]]
- [[_COMMUNITY_Community 108|Community 108]]
- [[_COMMUNITY_Community 109|Community 109]]
- [[_COMMUNITY_Community 110|Community 110]]
- [[_COMMUNITY_Community 111|Community 111]]
- [[_COMMUNITY_Community 112|Community 112]]

## God Nodes (most connected - your core abstractions)
1. `DefaultBetSettings()` - 41 edges
2. `T` - 37 edges
3. `Miner` - 36 edges
4. `makeBet()` - 36 edges
5. `main()` - 35 edges
6. `Server` - 31 edges
7. `Connection` - 31 edges
8. `Streamer` - 30 edges
9. `Client` - 27 edges
10. `twoOutcomes()` - 27 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `NewServer()`  [INFERRED]
  cmd/config-editor/main.go → internal/configeditor/server.go
- `main()` --calls--> `RunTUI()`  [INFERRED]
  cmd/config-editor/main.go → internal/configeditor/tui.go
- `main()` --calls--> `getEnv()`  [INFERRED]
  cmd/db-seed/main.go → internal/config/config.go
- `main()` --calls--> `OpenPostgres()`  [INFERRED]
  cmd/db-seed/main.go → internal/store/postgres.go
- `main()` --calls--> `Parse()`  [INFERRED]
  cmd/db-seed/main.go → internal/version/version.go

## Import Cycles
- None detected.

## Communities (113 total, 24 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.05
Nodes (38): Context, Miner, Streamer, Campaign, Duration, GameInfo, Tag, Time (+30 more)

### Community 1 - "Community 1"
Cohesion: 0.08
Nodes (58): addCategoryItem(), addStreamerItem(), addTag(), addTeamItem(), api(), assignNum(), assignTriToggle(), collectCategories() (+50 more)

### Community 2 - "Community 2"
Cohesion: 0.06
Nodes (74): B, Bet, BetSettings, Mutex, Streamer, Time, T, BetSettings (+66 more)

### Community 3 - "Community 3"
Cohesion: 0.04
Nodes (47): Account Configuration, Authentication, CI/CD Pipelines, Comparison Matrix, Configuration (Docker), Configuration (Fly.io), Deployment Guide, Deployment Options (+39 more)

### Community 4 - "Community 4"
Cohesion: 0.18
Nodes (27): apiRequest(), newTestServer(), TestCleanConfig(), TestCreateAccount(), TestCreateAccountConflict(), TestCreateAccountInvalidName(), TestCreateAccountValidationFail(), TestDeleteAccount() (+19 more)

### Community 5 - "Community 5"
Cohesion: 0.19
Nodes (17): accountMeta, Server, cleanConfig(), isValidDuration(), mergeSecretsBack(), readJSON(), removeEmpty(), sendError() (+9 more)

### Community 6 - "Community 6"
Cohesion: 0.05
Nodes (43): 1.10.1. Managing the Service, 1.10.2. Uninstalling, 1.10.3. Default File Locations, 1.10. Linux Service (systemd / OpenRC), 1.11.1. Managing the Service, 1.11.2. Uninstalling, 1.11. Windows Service, 1.12.1. Setup (+35 more)

### Community 7 - "Community 7"
Cohesion: 0.06
Nodes (39): BetSettingsConfig, CategoryConfig, ResolveBatchConfig(), AuthConfig, boolPtr(), TestIsBatchEnabled_Nil(), TestResolveBatchConfig_BothNil(), TestResolveBatchConfig_GlobalOnly() (+31 more)

### Community 8 - "Community 8"
Cohesion: 0.08
Nodes (29): circuitBreaker, isRetryableGQLError(), IsTransientError(), NewClient(), NewClientForTest(), TestIsRetryableGQLError(), TestIsTransientError(), wrapTransientGQLError() (+21 more)

### Community 9 - "Community 9"
Cohesion: 0.07
Nodes (29): Batcher, FeaturesConfig, FollowersConfig, AuthConfig, CategoryWatcherConfig, AccountConfig, NotificationsConfig, Priority (+21 more)

### Community 10 - "Community 10"
Cohesion: 0.09
Nodes (14): generateDeviceID(), NewAuthenticator(), NewForTest(), CookieJar, AccountConfig, Authenticator, AuthConfig, Client (+6 more)

### Community 11 - "Community 11"
Cohesion: 0.06
Nodes (25): NewManager(), Handler, NewHandler(), Manager, Client, Context, Handler, Logger (+17 more)

### Community 12 - "Community 12"
Cohesion: 0.08
Nodes (35): ConnectionSnapshot, DebugPredictionEntry, DebugWatchingEntry, Miner, Time, Context, Miner, Streamer (+27 more)

### Community 13 - "Community 13"
Cohesion: 0.09
Nodes (16): GoalContribution, ChannelPointsContext, GameResp, GoalContribution, PlaybackAccessToken, StreamInfoResponse, TeamMember, TopStream (+8 more)

### Community 14 - "Community 14"
Cohesion: 0.10
Nodes (17): Authenticator, AccountConfig, Context, Logger, Provider, RWMutex, Streamer, Time (+9 more)

### Community 15 - "Community 15"
Cohesion: 0.12
Nodes (28): BatchConfig, batchKey, Context, Duration, Event, Logger, Mutex, Once (+20 more)

### Community 16 - "Community 16"
Cohesion: 0.22
Nodes (9): Conn, Logger, Message, Mutex, Once, Provider, Connection, Time (+1 more)

### Community 17 - "Community 17"
Cohesion: 0.17
Nodes (25): Time, T, Message, ParseMessage(), serverTime(), splitTopic(), TestMessage_String(), TestParseMessage_ChannelIDFallbackToTopicUser() (+17 more)

### Community 18 - "Community 18"
Cohesion: 0.13
Nodes (22): historyAggregate, HistoryEntry, Request, ResponseWriter, AnalyticsServer, Streamer, T, errorResponse (+14 more)

### Community 19 - "Community 19"
Cohesion: 0.06
Nodes (47): Int64, T, CommunityGoal, Context, Event, Message, Miner, Outcome (+39 more)

### Community 20 - "Community 20"
Cohesion: 0.13
Nodes (21): Client, Context, Mutex, Request, Response, T, inventoryDrop, claimFailedResponse() (+13 more)

### Community 21 - "Community 21"
Cohesion: 0.09
Nodes (13): Connection, Context, Mutex, Provider, T, newTestConnection(), TestHandleResponse_ERR_BADAUTH_AlreadyRefreshedByAnother(), TestHandleResponse_ERR_BADAUTH_RefreshesAndResubscribes() (+5 more)

### Community 22 - "Community 22"
Cohesion: 0.12
Nodes (21): Attr, Context, Event, Handler, Level, Mutex, Logger, T (+13 more)

### Community 23 - "Community 23"
Cohesion: 0.18
Nodes (28): _install-service.sh script, ask(), banner(), confirm(), DEFAULT_CONFIG_DIR, DEFAULT_DATA_DIR, DEFAULT_ENV_FILE, DEFAULT_INSTALL_DIR (+20 more)

### Community 24 - "Community 24"
Cohesion: 0.16
Nodes (12): Drop, GameInfo, Time, Campaign, Context, RawMessage, Streamer, Client (+4 more)

### Community 25 - "Community 25"
Cohesion: 0.17
Nodes (11): Connection, Context, Logger, Message, Mutex, Provider, Pool, PubSubTopic (+3 more)

### Community 26 - "Community 26"
Cohesion: 0.07
Nodes (26): For /graphify add and --watch, For /graphify query, For the commit hook and native CLAUDE.md integration, For --update and --cluster-only, /graphify, Honesty Rules, Interpreter guard for subcommands, Part A - Structural extraction for code files (+18 more)

### Community 27 - "Community 27"
Cohesion: 0.08
Nodes (24): `401 Unauthorized` errors in logs, All bets on one outcome keep losing, `authenticated as "X" but config expects "Y"`, Authentication errors, Config changes not taking effect, Config issues, Drop issues, Drops not being claimed (+16 more)

### Community 28 - "Community 28"
Cohesion: 0.10
Nodes (20): API, CategoryWatcher, Dispatcher, AccountConfig, Bool, DeviceCodeStatus, Event, EventPrediction (+12 more)

### Community 29 - "Community 29"
Cohesion: 0.09
Nodes (11): CommunityGoal, HistoryEntry, PointsMultiplier, RWMutex, StreamerSettings, Time, HistoryEntry, PointsMultiplier (+3 more)

### Community 30 - "Community 30"
Cohesion: 0.12
Nodes (18): Client, Context, EventPrediction, Message, Miner, Mutex, Provider, Streamer (+10 more)

### Community 31 - "Community 31"
Cohesion: 0.15
Nodes (17): Handler, Logger, RWMutex, Server, AnalyticsServer, Store, Streamer, checkCredentials() (+9 more)

### Community 32 - "Community 32"
Cohesion: 0.11
Nodes (18): `batch` (global batching defaults), `bet.filter_condition` (optional), `bet` (nested under `streamer_defaults` and per-streamer `settings`), `blacklist`, `category_blacklist`, `category_watcher`, Configuration Reference, `features` (+10 more)

### Community 33 - "Community 33"
Cohesion: 0.18
Nodes (21): AnalyticsServer, Context, Duration, Level, Manager, Store, getEnv(), ColorSupported() (+13 more)

### Community 34 - "Community 34"
Cohesion: 0.07
Nodes (48): AccountConfigFromJSON(), AccountConfigToJSON(), applyDefaults(), applyEnvOverrides(), isOwnerAccount(), LoadAccountConfig(), LoadAllAccountConfigs(), parseProxyURL() (+40 more)

### Community 35 - "Community 35"
Cohesion: 0.09
Nodes (32): AccountConfig, Manager, T, AccountConfig, AccountRow, Logger, Store, T (+24 more)

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
Cohesion: 0.30
Nodes (6): DeviceCodeResponse, DeviceCodeStatus, TokenErrorResponse, TokenResponse, Authenticator, Context

### Community 42 - "Community 42"
Cohesion: 0.18
Nodes (27): AccountRow, AnalyticsServer, ResponseRecorder, Store, T, accountsRequest(), minimalAccountCfg(), newFakeStore() (+19 more)

### Community 44 - "Community 44"
Cohesion: 0.23
Nodes (6): DB, AccountRow, Once, OpenPostgres(), scanRow(), PostgresStore

### Community 45 - "Community 45"
Cohesion: 0.42
Nodes (4): promptLine(), loginResponse, Authenticator, Context

### Community 46 - "Community 46"
Cohesion: 0.20
Nodes (11): CancelFunc, AccountConfig, Context, Event, Logger, Miner, RWMutex, Twitch (+3 more)

### Community 47 - "Community 47"
Cohesion: 0.42
Nodes (5): Streamer, NewStreamerTopic(), NewUserTopic(), PubSubTopic, PubSubTopicType

### Community 48 - "Community 48"
Cohesion: 0.22
Nodes (8): graphify reference: extra exports and benchmark, Step 6b - Wiki (only if --wiki flag), Step 7 - Neo4j export (only if --neo4j or --neo4j-push flag), Step 7a - FalkorDB export (only if --falkordb or --falkordb-push flag), Step 7b - SVG export (only if --svg flag), Step 7c - GraphML export (only if --graphml flag), Step 7d - MCP server (only if --mcp flag), Step 8 - Token reduction benchmark (only if total_words > 5000)

### Community 49 - "Community 49"
Cohesion: 0.20
Nodes (10): 1. Clone and configure, 2. Set required environment variables, 3. Run, 4. Authenticate, 5. Verify it's working, Automatic updates, Getting Started, Next steps (+2 more)

### Community 51 - "Community 51"
Cohesion: 0.29
Nodes (6): RawMessage, MessageData, Request, RequestData, Response, RequestData

### Community 52 - "Community 52"
Cohesion: 0.25
Nodes (8): Architecture, Data flow, Key design decisions, Notification flow, Package map, Per-account miner lifecycle, Startup, Telemetry flow

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
Cohesion: 0.25
Nodes (20): editAccountFields, applyCategoryWatcherSection(), applyEditFields(), applyFeaturesSection(), applyStreamersSection(), applyTeamWatcherSection(), boolVal(), intVal() (+12 more)

### Community 73 - "Community 73"
Cohesion: 0.67
Nodes (3): Key resources, Twitch Channel Points Miner — Go Edition, Wiki pages

### Community 98 - "Community 98"
Cohesion: 0.26
Nodes (3): Context, RawMessage, Response

### Community 99 - "Community 99"
Cohesion: 0.21
Nodes (6): Cookie, CookieJar, CookieFileExists(), NewCookieJar(), RWMutex, Time

### Community 100 - "Community 100"
Cohesion: 0.16
Nodes (14): main(), openBrowser(), baseNotifier, Client, Context, Event, T, Webhook (+6 more)

### Community 103 - "Community 103"
Cohesion: 0.48
Nodes (6): T, TestNoopStore_ChangesIsNil(), TestNoopStore_DeleteIsNoop(), TestNoopStore_GetAccountReturnsNil(), TestNoopStore_ListAccountsReturnsEmpty(), TestNoopStore_UpsertIsNoop()

### Community 104 - "Community 104"
Cohesion: 1.00
Nodes (3): Time, AccountRow, Store

### Community 106 - "Community 106"
Cohesion: 0.20
Nodes (12): Client, Context, Duration, Logger, Config, heartbeatPayload, Sender, detectDeployment() (+4 more)

### Community 107 - "Community 107"
Cohesion: 0.37
Nodes (13): NewServer(), HandlerFunc, T, TestCheckForUpdate_DevVersion(), TestCheckForUpdate_NewerAvailable(), TestCheckForUpdate_NoMatchingAsset(), TestCheckForUpdate_PopulatesAssetURL(), TestCheckForUpdate_ServerError() (+5 more)

### Community 108 - "Community 108"
Cohesion: 0.20
Nodes (17): Logger, ghAsset, Context, Logger, runAutoUpdate(), ghAsset, ghRelease, UpdateInfo (+9 more)

### Community 110 - "Community 110"
Cohesion: 0.31
Nodes (8): Logger, T, Twitch, envOrDefault(), LoadTwitchFromEnv(), TestClientIDsForGQL_Dedup(), TestLoadTwitchFromEnv_Defaults(), TestLoadTwitchFromEnv_EnvOverride()

### Community 111 - "Community 111"
Cohesion: 0.18
Nodes (7): GenerateHex(), PubSubTopic, Request, Context, EventPrediction, Streamer, Client

### Community 112 - "Community 112"
Cohesion: 0.22
Nodes (8): Attribution, Code of Conduct, Corrective Action Guide, Enforcement, Our Pledge, Our Standards, Reporting, Scope

## Knowledge Gaps
- **473 isolated node(s):** `_edit-config.sh script`, `DEFAULT_SERVICE_NAME`, `DEFAULT_INSTALL_DIR`, `DEFAULT_CONFIG_DIR`, `DEFAULT_DATA_DIR` (+468 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **24 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `main()` connect `Community 33` to `Community 34`, `Community 100`, `Community 106`, `Community 11`, `Community 44`, `Community 108`, `Community 110`, `Community 22`, `Community 28`, `Community 31`?**
  _High betweenness centrality (0.215) - this node is a cross-community bridge._
- **Why does `Parse()` connect `Community 100` to `Community 33`, `Community 34`, `Community 108`, `Community 17`, `Community 19`, `Community 24`?**
  _High betweenness centrality (0.137) - this node is a cross-community bridge._
- **Why does `Setup()` connect `Community 22` to `Community 33`, `Community 35`, `Community 4`, `Community 15`, `Community 20`, `Community 21`, `Community 30`?**
  _High betweenness centrality (0.136) - this node is a cross-community bridge._
- **Are the 39 inferred relationships involving `DefaultBetSettings()` (e.g. with `BenchmarkBetCalculate()` and `BenchmarkFilterConditionSkip()`) actually correct?**
  _`DefaultBetSettings()` has 39 INFERRED edges - model-reasoned connections that need verification._
- **Are the 23 inferred relationships involving `main()` (e.g. with `.Load()` and `getEnv()`) actually correct?**
  _`main()` has 23 INFERRED edges - model-reasoned connections that need verification._
- **What connects `_edit-config.sh script`, `DEFAULT_SERVICE_NAME`, `DEFAULT_INSTALL_DIR` to the rest of the system?**
  _473 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.05069124423963134 - nodes in this community are weakly interconnected._