# Graph Report - twitch-miner-go  (2026-09-04)

## Corpus Check
- 183 files · ~166,243 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2005 nodes · 4544 edges · 132 communities (92 shown, 23 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 293 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `6a5696a8`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Stream
- web/app.js
- time.Time
- Deployment Guide
- testing.T
- DefaultBetSettings
- twitch-miner-go - Efficient Auto Drops & Points Claim for Twitch
- AccountConfig
- Client
- Dispatcher
- Authenticator
- Manager
- What You Must Do When Invoked
- context.Context
- Client
- Batcher
- install-service.sh
- Message
- net/http.ResponseWriter
- .handlePredictionCreated
- newMockTransport
- sync.Mutex
- colorHandler
- Troubleshooting
- SelectStreamersToWatch
- Connection
- main
- server_test.go
- Configuration Reference
- Streamer
- tray.go
- AnalyticsServer
- newTestConfigServer
- time.Duration
- fakeMgr
- NewStreamer
- static/app.js
- logs.js
- autostart_unix.go
- Getting Started
- operations_test.go
- .pollForToken
- NewBatcher
- newTestAccountsServer
- .handle2FA
- Manager
- Provider setup
- Prediction Strategies
- Logger
- Miner
- newTestManager
- poller_test.go
- Priority
- Contributing
- Event
- Poller
- newTestMiner
- tokenChangingMock
- graphify reference: extra exports and benchmark
- Data flow
- Advanced Guide
- Security Policy
- PULL_REQUEST_TEMPLATE.md
- Run
- ResolveBatchConfig
- Authentication
- FAQ
- LoadTwitchFromEnv
- NewDrop
- Raid
- graphify reference: query, path, explain
- autostart_windows.go
- Twitch Channel Points Miner — Go Edition
- AGENTS.md
- os/exec.Cmd
- tui.go
- _run.sh
- CategoryWatcher
- Validate
- CLAUDE.md
- .claude/CLAUDE.md
- GEMINI.md
- copilot-instructions.md
- graphify reference: add a URL and watch a folder
- graphify reference: commit hook and native CLAUDE.md integration
- graphify reference: incremental update and cluster-only
- graphify reference: GitHub clone and cross-repo merge
- graphify reference: transcribe video and audio
- Troubleshooting
- graphify.js
- rules/graphify.md
- workflows/graphify.md
- extraction-spec.md
- Operations
- github.com/Guliveer/twitch-miner-go
- commit-msg
- Sender
- checkWithURL
- pre-commit
- mockAuthProvider
- pre-push
- noopAuth
- Code of Conduct
- _run-localdev.sh
- edit-config.sh
- gen-cookie-key.sh
- gen-dashboard-auth.sh
- install-hooks.sh
- CI/CD Pipelines
- config_api.go
- Configuration (Docker)
- Windows Service
- Required Configuration
- Configuration (Fly.io)
- Linux Service (systemd / OpenRC)

## God Nodes (most connected - your core abstractions)
1. `Streamer` - 98 edges
2. `Logger` - 51 edges
3. `newMockTransport()` - 44 edges
4. `DefaultBetSettings()` - 41 edges
5. `Miner` - 36 edges
6. `makeBet()` - 36 edges
7. `newTestClientWithCapture()` - 31 edges
8. `Connection` - 30 edges
9. `main()` - 28 edges
10. `Authenticator` - 28 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `NewServer()`  [EXTRACTED]
  cmd/config-editor/main.go → internal/configeditor/server.go
- `main()` --calls--> `OpenPostgres()`  [EXTRACTED]
  cmd/db-seed/main.go → internal/store/postgres.go
- `main()` --calls--> `applyNoConsole()`  [INFERRED]
  cmd/twitch-miner-go/main.go → cmd/twitch-miner-go/tray_enabled.go
- `main()` --calls--> `ColorSupported()`  [EXTRACTED]
  cmd/twitch-miner-go/main.go → internal/logger/color_support.go
- `main()` --calls--> `Setup()`  [EXTRACTED]
  cmd/twitch-miner-go/main.go → internal/logger/logger.go

## Import Cycles
- None detected.

## Communities (132 total, 23 thin omitted)

### Community 0 - "Stream"
Cohesion: 0.09
Nodes (11): StreamInfoResponse, Campaign, NewCampaign(), GameInfo, Stream, Tag, NewStream(), campaignMatchesStreamer() (+3 more)

### Community 1 - "web/app.js"
Cohesion: 0.08
Nodes (69): addCategoryItem(), addStreamerItem(), addTag(), addTeamItem(), api(), assignFloatFromEl(), assignNum(), assignNumFromEl() (+61 more)

### Community 2 - "time.Time"
Cohesion: 0.06
Nodes (38): BetSettingsConfig, FilterConditionConfig, testing.B, time.Time, Miner, BenchmarkBetCalculate(), BenchmarkFilterConditionSkip(), BenchmarkParseMessage() (+30 more)

### Community 3 - "Deployment Guide"
Cohesion: 0.25
Nodes (6): Comparison Matrix, Deployment Guide, Deployment Options, Next Steps, Security Best Practices, Table of Contents

### Community 4 - "testing.T"
Cohesion: 0.10
Nodes (37): testing.T, ParseMessage(), splitTopic(), TestMessage_String(), TestParseMessage_ChannelIDFallbackToTopicUser(), TestParseMessage_ChannelIDFromDataChannelID(), TestParseMessage_ClaimAvailable(), TestParseMessage_EmptyObject() (+29 more)

### Community 5 - "DefaultBetSettings"
Cohesion: 0.17
Nodes (38): DefaultBetSettings(), makeBet(), TestCalculate_AllOutcomesZeroUsers(), TestCalculate_AmountCappedByMaxPoints(), TestCalculate_AmountFormula(), TestCalculate_EqualOdds(), TestCalculate_HighOdds(), TestCalculate_MostVoted() (+30 more)

### Community 6 - "twitch-miner-go - Efficient Auto Drops & Points Claim for Twitch"
Cohesion: 0.04
Nodes (46): 1.10.1. Managing the Service, 1.10.2. Uninstalling, 1.10.3. Default File Locations, 1.10. Linux Service (systemd / OpenRC), 1.11.1. Managing the Service, 1.11.2. Uninstalling, 1.11. Windows Service, 1.12.1. Setup (+38 more)

### Community 7 - "AccountConfig"
Cohesion: 0.24
Nodes (14): DiscordConfig, FeaturesConfig, FollowersConfig, GotifyConfig, MatrixConfig, PushoverConfig, StreamerConfig, StreamerSettingsConfig (+6 more)

### Community 8 - "Client"
Cohesion: 0.11
Nodes (16): net/http.Transport, circuitBreaker, gqlError, gqlExtensions, gqlRequest, gqlResponse, operationBehavior, persistedQuery (+8 more)

### Community 9 - "Dispatcher"
Cohesion: 0.15
Nodes (7): sync/atomic.Bool, NotifyFunc, Dispatcher, NewDispatcher(), parseEvents(), Notifier, notifierEntry

### Community 10 - "Authenticator"
Cohesion: 0.05
Nodes (45): Cookie, CookieJar, generateDeviceID(), GenerateHex(), Authenticator, NewAuthenticator(), CookieFileExists(), NewCookieJar() (+37 more)

### Community 11 - "Manager"
Cohesion: 0.10
Nodes (9): Handler, Manager, twitch.Client, NewManager(), NewHandler(), twitch.PrivateMessage, twitch.UserJoinMessage, twitch.UserNoticeMessage (+1 more)

### Community 12 - "What You Must Do When Invoked"
Cohesion: 0.07
Nodes (26): For /graphify add and --watch, For /graphify query, For the commit hook and native CLAUDE.md integration, For --update and --cluster-only, /graphify, Honesty Rules, Interpreter guard for subcommands, Part A - Structural extraction for code files (+18 more)

### Community 13 - "context.Context"
Cohesion: 0.07
Nodes (10): context.Context, encoding/json.RawMessage, GameResp, GoalContribution, PlaybackAccessToken, TeamMember, TopStream, Client (+2 more)

### Community 14 - "Client"
Cohesion: 0.11
Nodes (8): sync.Map, sync.RWMutex, Client, isStaleUsernameError(), NewClient(), TestIsStaleUsernameError(), spadeCache, spadeCacheEntry

### Community 15 - "Batcher"
Cohesion: 0.31
Nodes (4): batchKey, sync.Once, batchEntry, Batcher

### Community 16 - "install-service.sh"
Cohesion: 0.18
Nodes (28): ask(), banner(), confirm(), DEFAULT_CONFIG_DIR, DEFAULT_DATA_DIR, DEFAULT_ENV_FILE, DEFAULT_INSTALL_DIR, DEFAULT_LOG_LEVEL (+20 more)

### Community 17 - "Message"
Cohesion: 0.19
Nodes (6): extractNestedInt(), Miner, mapReasonToEvent(), Message, serverTime(), MessageType

### Community 18 - "net/http.ResponseWriter"
Cohesion: 0.06
Nodes (46): accountMeta, Server, io.Reader, net/http.HandlerFunc, net/http.Request, net/http.Response, net/http.ResponseWriter, net/http.ServeMux (+38 more)

### Community 19 - ".handlePredictionCreated"
Cohesion: 0.09
Nodes (22): ChannelPointsContext, BoolFromMap(), FloatFromAny(), IntFromAny(), IntFromMap(), StringFromAny(), StringFromMap(), TestBoolFromMap() (+14 more)

### Community 20 - "newMockTransport"
Cohesion: 0.10
Nodes (64): bytes.Buffer, NewForTest(), NewClientForTest(), boolPtr(), Client, makeStreamerWithDrop(), makeStreamerWithDropAtProgress(), milestoneEventCount() (+56 more)

### Community 21 - "sync.Mutex"
Cohesion: 0.12
Nodes (5): sync.Mutex, pollLoop(), TeamWatcher, NewTeamWatcher(), mockAuthProvider

### Community 22 - "colorHandler"
Cohesion: 0.21
Nodes (9): io.Writer, log/slog.Attr, log/slog.Handler, log/slog.Level, log/slog.Record, copyAttrs(), newColorHandler(), colorHandler (+1 more)

### Community 23 - "Troubleshooting"
Cohesion: 0.07
Nodes (27): `401 Unauthorized` errors in logs, All bets on one outcome keep losing, `authenticated as "X" but config expects "Y"`, Authentication errors, Config changes not taking effect, Config editor & tray issues, Config issues, Drop issues (+19 more)

### Community 24 - "SelectStreamersToWatch"
Cohesion: 0.24
Nodes (22): SelectStreamersToWatch(), makeOnlineStreamer(), TestSelectStreamersToWatch_DropsOnly_IncludedWhenHasCampaignIDs(), TestSelectStreamersToWatch_DropsOnly_SkippedWhenNoCampaignIDs(), TestSelectStreamersToWatch_EndingSoonest(), TestSelectStreamersToWatch_EndingSoonest_PicksFirstEnd(), TestSelectStreamersToWatch_EndingSoonest_SkipsNoCampaigns(), TestSelectStreamersToWatch_Freeze_AllFrozenReturnsNone() (+14 more)

### Community 25 - "Connection"
Cohesion: 0.06
Nodes (23): github.com/coder/websocket.Conn, Provider, PubSubTopic, NewStreamerTopic(), NewUserTopic(), Connection, NewConnection(), Connection (+15 more)

### Community 26 - "main"
Cohesion: 0.16
Nodes (18): main(), playStartupAnimation(), resolveFileWatchInterval(), resolveLogDir(), resolveLogFormat(), resolveLogLevel(), resolveLogNoTime(), resolveNoBanner() (+10 more)

### Community 27 - "server_test.go"
Cohesion: 0.30
Nodes (18): apiRequest(), newTestServer(), TestCleanConfig(), TestCreateAccount(), TestCreateAccountConflict(), TestCreateAccountInvalidName(), TestCreateAccountValidationFail(), TestDeleteAccount() (+10 more)

### Community 28 - "Configuration Reference"
Cohesion: 0.11
Nodes (18): `batch` (global batching defaults), `bet.filter_condition` (optional), `bet` (nested under `streamer_defaults` and per-streamer `settings`), `blacklist`, `category_blacklist`, `category_watcher`, Configuration Reference, `features` (+10 more)

### Community 29 - "Streamer"
Cohesion: 0.10
Nodes (14): Miner, Streamer, applyPriority(), applyPriorityDrops(), applyPriorityEndingSoonest(), applyPriorityLowAvailability(), applyPriorityOrder(), applyPriorityPoints() (+6 more)

### Community 30 - "tray.go"
Cohesion: 0.19
Nodes (13): applyNoConsole(), startTray(), fyne.io/systray.MenuItem, RunServiceAction(), ServiceScriptAvailable(), serviceScriptName(), Available(), clickCh() (+5 more)

### Community 31 - "AnalyticsServer"
Cohesion: 0.12
Nodes (21): net/http.Handler, net/http.Server, checkCredentials(), generateRequestID(), newRateLimiter(), RequestIDFromContext(), withAuth(), withLogging() (+13 more)

### Community 32 - "newTestConfigServer"
Cohesion: 0.30
Nodes (15): net/http/httptest.ResponseRecorder, configRequest(), AnalyticsServer, newTestConfigServer(), TestConfigGenerate_InvalidConfig(), TestConfigGenerate_InvalidJSON(), TestConfigGenerate_MissingUsername(), TestConfigGenerate_Valid() (+7 more)

### Community 33 - "time.Duration"
Cohesion: 0.14
Nodes (6): CategoryConfig, jsonDuration, TeamConfig, time.Duration, CategoryWatcherConfig, TeamWatcherConfig

### Community 35 - "NewStreamer"
Cohesion: 0.27
Nodes (11): DefaultStreamerSettings(), StreamerSettings, NewStreamer(), ParseChatPresence(), ShouldJoinChat(), TestSetOnline_AlreadyOnline_NoOp(), TestSetOnline_Carryover_FirstOnline(), TestSetOnline_Carryover_StreakResolved_LongGap() (+3 more)

### Community 36 - "static/app.js"
Cohesion: 0.30
Nodes (14): buildFilterParams(), clearFilters(), debounce(), escapeHTML(), fetchJSON(), formatPoints(), initFilterListeners(), loadFilters() (+6 more)

### Community 37 - "logs.js"
Cohesion: 0.27
Nodes (13): buildFilterParams(), clearFilters(), escapeHTML(), fetchJSON(), formatPoints(), handleSort(), loadFilters(), populateSelect() (+5 more)

### Community 38 - "autostart_unix.go"
Cohesion: 0.30
Nodes (12): autostartDirPath(), AutostartEnabled(), autostartFilePath(), ClearAutostart(), escapePlist(), launchAgentPath(), quote(), SetAutostart() (+4 more)

### Community 39 - "Getting Started"
Cohesion: 0.15
Nodes (13): 1. Clone and configure, 2. Set required environment variables, 3. Run, 4. Authenticate, 5. Verify it's working, Automatic updates, Getting Started, Going deeper (+5 more)

### Community 40 - "operations_test.go"
Cohesion: 0.40
Nodes (10): mockTransport, assertVarEquals(), assertVarPresent(), Client, newTestGQLClient(), TestGetChannelPointsContext_SendsChannelLogin(), TestGetPlaybackAccessToken_PlatformIsNonEmptyString(), TestGetPlaybackAccessToken_SendAllRequiredVariables() (+2 more)

### Community 41 - ".pollForToken"
Cohesion: 0.31
Nodes (4): DeviceCodeResponse, TokenErrorResponse, TokenResponse, Authenticator

### Community 42 - "NewBatcher"
Cohesion: 0.49
Nodes (10): NewBatcher(), boolPtr(), newTestLogger(), TestBatcher_BuffersAndFlushes(), TestBatcher_ImmediateEventsBypassBuffer(), TestBatcher_MaxEntriesSplitsMessages(), TestBatcher_SingleEntryNotBatched(), TestBatcher_StopFlushesRemaining() (+2 more)

### Community 44 - "newTestAccountsServer"
Cohesion: 0.06
Nodes (30): database/sql.DB, accountsRequest(), AnalyticsServer, minimalAccountCfg(), newFakeStore(), newTestAccountsServer(), TestCreateAccount_InvalidJSON(), TestCreateAccount_MissingUsername() (+22 more)

### Community 45 - ".handle2FA"
Cohesion: 0.42
Nodes (3): loginResponse, Authenticator, promptLine()

### Community 46 - "Manager"
Cohesion: 0.24
Nodes (6): context.CancelFunc, Manager, NewManager(), NewMiner(), Twitch, entry

### Community 47 - "Provider setup"
Cohesion: 0.17
Nodes (12): Discord, Event reference, Gotify, Matrix, Notification batching, Notifications, Per-account vs global credentials, Provider setup (+4 more)

### Community 48 - "Prediction Strategies"
Cohesion: 0.17
Nodes (12): Betting amount calculation, Delay modes, Filter conditions, `HIGH_ODDS`, `MOST_VOTED`, `NUMBER_1` through `NUMBER_8`, `PERCENTAGE`, Prediction Strategies (+4 more)

### Community 49 - "Logger"
Cohesion: 0.35
Nodes (8): sync/atomic.Value, DefaultConfig(), Logger, ParseLevel(), Setup(), discardLogger(), newTestLogger(), Config

### Community 50 - "Miner"
Cohesion: 0.16
Nodes (4): time.Timer, Miner, oneTimeEventMessage(), API

### Community 51 - "newTestManager"
Cohesion: 0.42
Nodes (10): newTestManager(), testConfig(), TestManager_EntriesReturnsSnapshot(), TestManager_RestartReplacesEntry(), TestManager_StartAddsEntry(), TestManager_StartIdempotent(), TestManager_StopAllClearsEntries(), TestManager_StopRemovesEntry() (+2 more)

### Community 52 - "poller_test.go"
Cohesion: 0.44
Nodes (10): minimalConfigJSON(), newTestPoller(), TestPoller_DisabledAccountStopsMiner(), TestPoller_InvalidConfigSkipped(), TestPoller_MultipleAccounts(), TestPoller_NewAccountStartsMiner(), TestPoller_RemovedAccountStopsMiner(), TestPoller_SameTimestampNoRestart() (+2 more)

### Community 53 - "Priority"
Cohesion: 0.29
Nodes (6): AllEvents(), Priority, ParseEvent(), ParseFollowersOrder(), ParsePriority(), FollowersOrder

### Community 54 - "Contributing"
Cohesion: 0.33
Nodes (6): Automated Versioning, Commit Convention, Contributing, Documentation and Wiki, Pull Requests, Setting Up Git Hooks

### Community 55 - "Event"
Cohesion: 0.09
Nodes (10): net/http.Client, sync/atomic.Int64, Event, baseNotifier, Discord, Gotify, Matrix, Pushover (+2 more)

### Community 56 - "Poller"
Cohesion: 0.33
Nodes (4): NewPoller(), Store, minerManager, Poller

### Community 57 - "newTestMiner"
Cohesion: 0.40
Nodes (9): newTestMiner(), raidMessage(), TestHandleRaid_Dedup(), TestHandleRaid_EmptyRaidID(), TestHandleRaid_FollowRaidDisabled(), TestHandleRaid_MissingRaidData(), TestHandleRaid_NoStreamer(), TestHandleRaid_WrongMessageType() (+1 more)

### Community 59 - "graphify reference: extra exports and benchmark"
Cohesion: 0.22
Nodes (8): graphify reference: extra exports and benchmark, Step 6b - Wiki (only if --wiki flag), Step 7 - Neo4j export (only if --neo4j or --neo4j-push flag), Step 7a - FalkorDB export (only if --falkordb or --falkordb-push flag), Step 7b - SVG export (only if --svg flag), Step 7c - GraphML export (only if --graphml flag), Step 7d - MCP server (only if --mcp flag), Step 8 - Token reduction benchmark (only if total_words > 5000)

### Community 60 - "Data flow"
Cohesion: 0.22
Nodes (9): Architecture, Config editor & tray runtime, Data flow, Key design decisions, Notification flow, Package map, Per-account miner lifecycle, Startup (+1 more)

### Community 61 - "Advanced Guide"
Cohesion: 0.25
Nodes (8): Advanced Guide, Data & storage, Drops & campaigns, How watching works, Performance notes, Prediction strategies, PubSub (live event stream), Reliability & recovery

### Community 62 - "Security Policy"
Cohesion: 0.40
Nodes (4): Reporting a Vulnerability, Scope, Security Policy, Supported Versions

### Community 63 - "PULL_REQUEST_TEMPLATE.md"
Cohesion: 0.50
Nodes (3): Checklist, Notes for reviewer, What does this PR do?

### Community 64 - "Run"
Cohesion: 0.31
Nodes (8): T, Run(), TestRunAllItems(), TestRunContextCancellation(), TestRunEmpty(), TestRunReturnsFirstError(), TestRunWorkersBound(), TestRunZeroWorkersDefaultsToOne()

### Community 65 - "ResolveBatchConfig"
Cohesion: 0.46
Nodes (7): ResolveBatchConfig(), boolPtr(), TestIsBatchEnabled_Nil(), TestResolveBatchConfig_BothNil(), TestResolveBatchConfig_GlobalOnly(), TestResolveBatchConfig_ProviderDisablesGlobal(), TestResolveBatchConfig_ProviderOverridesGlobal()

### Community 66 - "Authentication"
Cohesion: 0.29
Nodes (7): Authentication, Cookie persistence, Obtaining an OAuth token, Priority chain, Recommended approach per environment, Token scopes required, Variable naming convention

### Community 67 - "FAQ"
Cohesion: 0.29
Nodes (7): Accounts & configuration, Behaviour & resources, Config editor & tray, Docker & Fly.io, FAQ, Running, Telemetry & data

### Community 68 - "LoadTwitchFromEnv"
Cohesion: 0.38
Nodes (5): envOrDefault(), LoadTwitchFromEnv(), TestClientIDsForGQL_Dedup(), TestLoadTwitchFromEnv_Defaults(), TestLoadTwitchFromEnv_EnvOverride()

### Community 69 - "NewDrop"
Cohesion: 0.17
Nodes (21): NewDrop(), Percentage(), refTime(), TestDropUpdate_FieldUpdate_DropInstanceID(), TestDropUpdate_FieldUpdate_HasPreconditionsMet(), TestDropUpdate_FieldUpdate_IsClaimed(), TestDropUpdate_IsClaimable_Claimed(), TestDropUpdate_IsClaimable_EmptyInstanceID() (+13 more)

### Community 71 - "graphify reference: query, path, explain"
Cohesion: 0.33
Nodes (5): For /graphify explain, For /graphify path, graphify reference: query, path, explain, Step 0 — Constrained query expansion (REQUIRED before traversal), Step 1 — Traversal

### Community 72 - "autostart_windows.go"
Cohesion: 0.53
Nodes (4): autostartCommand(), ClearAutostart(), SetAutostart(), SyncAutostart()

### Community 73 - "Twitch Channel Points Miner — Go Edition"
Cohesion: 0.40
Nodes (5): Key resources, Twitch Channel Points Miner — Go Edition, What this program does (in plain English), Where do I start?, Wiki pages

### Community 74 - "AGENTS.md"
Cohesion: 0.15
Nodes (12): Before writing code, CI/CD verification, Code quality, Comments, Completion checklist, Dependencies, Documentation, Error handling (+4 more)

### Community 75 - "os/exec.Cmd"
Cohesion: 0.40
Nodes (3): os/exec.Cmd, detach(), detach()

### Community 76 - "tui.go"
Cohesion: 0.18
Nodes (22): main(), openBrowser(), editAccountFields, applyCategoryWatcherSection(), applyEditFields(), applyFeaturesSection(), applyStreamersSection(), applyTeamWatcherSection() (+14 more)

### Community 78 - "CategoryWatcher"
Cohesion: 0.26
Nodes (5): LookupGameSlug(), RegisterGameSlug(), CategoryWatcher, NewCategoryWatcher(), categoryEntry

### Community 79 - "Validate"
Cohesion: 0.10
Nodes (32): fatalf(), main(), net/url.URL, AccountConfigFromJSON(), AccountConfigToJSON(), applyDefaults(), applyEnvOverrides(), getEnv() (+24 more)

### Community 80 - "CLAUDE.md"
Cohesion: 0.15
Nodes (12): Before writing code, CI/CD verification, Code quality, Comments, Completion checklist, Dependencies, Documentation, Error handling (+4 more)

### Community 81 - ".claude/CLAUDE.md"
Cohesion: 0.15
Nodes (12): Before writing code, CI/CD verification, Code quality, Comments, Completion checklist, Dependencies, Documentation, Error handling (+4 more)

### Community 82 - "GEMINI.md"
Cohesion: 0.15
Nodes (12): Before writing code, CI/CD verification, Code quality, Comments, Completion checklist, Dependencies, Documentation, Error handling (+4 more)

### Community 83 - "copilot-instructions.md"
Cohesion: 0.15
Nodes (12): Before writing code, CI/CD verification, Code quality, Comments, Completion checklist, Dependencies, Documentation, Error handling (+4 more)

### Community 84 - "graphify reference: add a URL and watch a folder"
Cohesion: 0.50
Nodes (3): For /graphify add, For --watch, graphify reference: add a URL and watch a folder

### Community 85 - "graphify reference: commit hook and native CLAUDE.md integration"
Cohesion: 0.50
Nodes (3): For git commit hook, For native CLAUDE.md integration, graphify reference: commit hook and native CLAUDE.md integration

### Community 86 - "graphify reference: incremental update and cluster-only"
Cohesion: 0.50
Nodes (3): For --cluster-only, For --update (incremental re-extraction), graphify reference: incremental update and cluster-only

### Community 89 - "Troubleshooting"
Cohesion: 0.67
Nodes (3): Troubleshooting, Troubleshooting: Docker Compose, Troubleshooting: Fly.io

### Community 106 - "Sender"
Cohesion: 0.26
Nodes (9): log/slog.Logger, detectDeployment(), LoadConfigFromEnv(), loadOrGenerateInstanceID(), NewSender(), newUUID(), Config, heartbeatPayload (+1 more)

### Community 107 - "checkWithURL"
Cohesion: 0.11
Nodes (29): runAutoUpdate(), CheckForUpdate(), checkWithURL(), DownloadAsset(), ExitForRestart(), findAssetURL(), FormatNotification(), isGitRepo() (+21 more)

### Community 112 - "Code of Conduct"
Cohesion: 0.22
Nodes (8): Attribution, Code of Conduct, Corrective Action Guide, Enforcement, Our Pledge, Our Standards, Reporting, Scope

### Community 125 - "CI/CD Pipelines"
Cohesion: 0.29
Nodes (7): CI/CD Pipelines, Disabling Workflows, Docker Publish (GHCR), Enabling Workflows, Fly.io Deploy, Manual Triggers, Workflow Overview

### Community 127 - "Configuration (Docker)"
Cohesion: 0.25
Nodes (8): Configuration (Docker), Docker Compose (GHCR), Health Checks, Image Versions, Quick Start (Docker), System Tray, Updating, Volume Mounts

### Community 128 - "Windows Service"
Cohesion: 0.25
Nodes (8): Config Editor & Tray in Service Mode, File Locations (defaults), How It Works, Managing the Service, Prerequisites, Quick Start, Uninstalling, Windows Service

### Community 130 - "Required Configuration"
Cohesion: 0.33
Nodes (6): Account Configuration, Authentication, Getting Browser Values, Getting TV Client ID, Required Configuration, Twitch Runtime Identifiers

### Community 131 - "Configuration (Fly.io)"
Cohesion: 0.33
Nodes (6): Configuration (Fly.io), Fly.io, Monitoring, Quick Start (Fly.io), Scaling, Secrets Management

### Community 134 - "Linux Service (systemd / OpenRC)"
Cohesion: 0.40
Nodes (5): File Locations (defaults), Linux Service (systemd / OpenRC), Managing the Service, Quick Start, Uninstalling

## Knowledge Gaps
- **316 isolated node(s):** `_run-localdev.sh script`, `_run.sh script`, `github.com/Guliveer/twitch-miner-go`, `TokenErrorResponse`, `operationBehavior` (+311 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 557 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **23 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Streamer` connect `Streamer` to `Stream`, `time.Time`, `NewStreamer`, `Raid`, `context.Context`, `Client`, `CategoryWatcher`, `Message`, `Miner`, `.handlePredictionCreated`, `net/http.ResponseWriter`, `newMockTransport`, `sync.Mutex`, `SelectStreamersToWatch`, `Connection`, `AnalyticsServer`?**
  _High betweenness centrality (0.073) - this node is a cross-community bridge._
- **Why does `Logger` connect `Logger` to `Client`, `Dispatcher`, `Authenticator`, `Manager`, `Client`, `Batcher`, `net/http.ResponseWriter`, `newMockTransport`, `sync.Mutex`, `Connection`, `main`, `tray.go`, `AnalyticsServer`, `NewBatcher`, `Manager`, `Miner`, `Event`, `Poller`, `CategoryWatcher`, `Validate`, `Sender`, `checkWithURL`?**
  _High betweenness centrality (0.066) - this node is a cross-community bridge._
- **Why does `Miner` connect `Miner` to `time.Time`, `AccountConfig`, `Dispatcher`, `Authenticator`, `Manager`, `Client`, `Manager`, `CategoryWatcher`, `Logger`, `sync.Mutex`, `Priority`, `Event`, `Connection`, `Streamer`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Are the 23 inferred relationships involving `newMockTransport()` (e.g. with `TestLogDropProgress_MixedPrintable()` and `TestLogDropProgress_NilPreconditions()`) actually correct?**
  _`newMockTransport()` has 23 INFERRED edges - model-reasoned connections that need verification._
- **What connects `_run-localdev.sh script`, `_run.sh script`, `github.com/Guliveer/twitch-miner-go` to the rest of the system?**
  _316 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Stream` be split into smaller, more focused modules?**
  _Cohesion score 0.09411764705882353 - nodes in this community are weakly interconnected._
- **Should `web/app.js` be split into smaller, more focused modules?**
  _Cohesion score 0.0825508607198748 - nodes in this community are weakly interconnected._