# Graph Report - twitch-miner-go  (2026-09-06)

## Corpus Check
- 200 files · ~178,072 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2183 nodes · 5039 edges · 158 communities (115 shown, 27 thin omitted)
- Extraction: 93% EXTRACTED · 7% INFERRED · 0% AMBIGUOUS · INFERRED: 374 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `046a8478`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- NewDrop
- web/app.js
- prediction.go
- Deployment Guide
- ParseMessage
- DefaultBetSettings
- twitch-miner-go - Efficient Auto Drops & Points Claim for Twitch
- Server
- Client
- Dispatcher
- CookieJar
- Manager
- What You Must Do When Invoked
- context.Context
- Client
- Batcher
- install-service.sh
- Message
- net/http.ResponseWriter
- CommunityGoal
- newMockTransport
- mockAuthProvider
- colorHandler
- Troubleshooting
- SelectStreamersToWatch
- Connection
- main
- Campaign
- Configuration Reference
- Drop
- tray.go
- middleware.go
- newTestConfigServer
- AnalyticsServer
- account.go
- .handlePredictionCreated
- static/app.js
- logs.js
- buildBody
- Getting Started
- noopAuth
- .pollForToken
- NewBatcher
- AccountRow
- .handle2FA
- Manager
- Provider setup
- Prediction Strategies
- Logger
- Miner
- manager_test.go
- newTestAccountsServer
- settings.go
- Contributing
- Event
- newTestFileWatcher
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
- server_test.go
- Stream
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
- Authenticator
- graphify.js
- rules/graphify.md
- workflows/graphify.md
- extraction-spec.md
- Operations
- github.com/Guliveer/twitch-miner-go
- findProc
- commit-msg
- testing.B
- checkWithURL
- pre-commit
- mockAuthProvider
- pre-push
- streams_api.go
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
- Streamer
- Required Configuration
- Configuration (Fly.io)
- newTestLogger
- SelectWatchSet
- Linux Service (systemd / OpenRC)
- sync.Mutex
- AccountConfig
- DebugSnapshot
- clientWithStreaks
- TeamWatcher
- time.Time
- Miner
- testing.T
- 1.6. Environment Variables
- capturingNotifier
- 1.7. Notifications
- autostart_unix.go
- NewAuthenticator
- 1.12. Deploy to Fly.io
- 1.5. Configuration
- 1.11. Windows Service
- newTestConnection
- Field
- NewStream
- NewServer
- NewStreamer
- StreamerSettings
- .RoundTrip

## God Nodes (most connected - your core abstractions)
1. `Streamer` - 110 edges
2. `Logger` - 53 edges
3. `newMockTransport()` - 44 edges
4. `DefaultBetSettings()` - 41 edges
5. `Miner` - 39 edges
6. `makeBet()` - 36 edges
7. `Event` - 31 edges
8. `Connection` - 31 edges
9. `newTestClientWithCapture()` - 31 edges
10. `main()` - 29 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `RunTUI()`  [EXTRACTED]
  cmd/config-editor/main.go → internal/configeditor/tui.go
- `main()` --calls--> `AccountConfigToJSON()`  [EXTRACTED]
  cmd/db-seed/main.go → internal/config/config.go
- `main()` --calls--> `OpenPostgres()`  [EXTRACTED]
  cmd/db-seed/main.go → internal/store/postgres.go
- `main()` --calls--> `applyNoConsole()`  [INFERRED]
  cmd/twitch-miner-go/main.go → cmd/twitch-miner-go/tray_enabled.go
- `main()` --calls--> `ParseKey()`  [EXTRACTED]
  cmd/twitch-miner-go/main.go → internal/encryption/encryption.go

## Import Cycles
- None detected.

## Communities (158 total, 27 thin omitted)

### Community 0 - "NewDrop"
Cohesion: 0.27
Nodes (19): NewDrop(), refTime(), TestDropUpdate_FieldUpdate_DropInstanceID(), TestDropUpdate_FieldUpdate_HasPreconditionsMet(), TestDropUpdate_FieldUpdate_IsClaimed(), TestDropUpdate_IsClaimable_Claimed(), TestDropUpdate_IsClaimable_EmptyInstanceID(), TestDropUpdate_IsClaimable_WithInstanceID() (+11 more)

### Community 1 - "web/app.js"
Cohesion: 0.08
Nodes (69): addCategoryItem(), addStreamerItem(), addTag(), addTeamItem(), api(), assignFloatFromEl(), assignNum(), assignNumFromEl() (+61 more)

### Community 2 - "prediction.go"
Cohesion: 0.13
Nodes (17): BetSettingsConfig, FilterConditionConfig, GetPredictionWindow(), BetSettings, Outcome, NewBet(), NewEventPrediction(), ParseCondition() (+9 more)

### Community 3 - "Deployment Guide"
Cohesion: 0.18
Nodes (9): Comparison Matrix, Deployment Guide, Deployment Options, Next Steps, Security Best Practices, Table of Contents, Troubleshooting, Troubleshooting: Docker Compose (+1 more)

### Community 4 - "ParseMessage"
Cohesion: 0.15
Nodes (21): ParseMessage(), splitTopic(), TestMessage_String(), TestParseMessage_ChannelIDFallbackToTopicUser(), TestParseMessage_ChannelIDFromDataChannelID(), TestParseMessage_ClaimAvailable(), TestParseMessage_EmptyObject(), TestParseMessage_Identifier() (+13 more)

### Community 5 - "DefaultBetSettings"
Cohesion: 0.17
Nodes (38): DefaultBetSettings(), makeBet(), TestCalculate_AllOutcomesZeroUsers(), TestCalculate_AmountCappedByMaxPoints(), TestCalculate_AmountFormula(), TestCalculate_EqualOdds(), TestCalculate_HighOdds(), TestCalculate_MostVoted() (+30 more)

### Community 6 - "twitch-miner-go - Efficient Auto Drops & Points Claim for Twitch"
Cohesion: 0.10
Nodes (21): 1.10.1. Managing the Service, 1.10.2. Uninstalling, 1.10.3. Default File Locations, 1.10. Linux Service (systemd / OpenRC), 1.13. Development, 1.14.1. Automatic updates, 1.14. Auto-Update, 1.15. License (+13 more)

### Community 7 - "Server"
Cohesion: 0.16
Nodes (18): accountMeta, Server, io.Reader, net/http.ServeMux, cleanConfig(), isValidDuration(), mergeSecretsBack(), readJSON() (+10 more)

### Community 8 - "Client"
Cohesion: 0.11
Nodes (16): net/http.Transport, circuitBreaker, gqlError, gqlExtensions, gqlRequest, gqlResponse, operationBehavior, persistedQuery (+8 more)

### Community 9 - "Dispatcher"
Cohesion: 0.18
Nodes (6): sync/atomic.Bool, Dispatcher, NewDispatcher(), parseEvents(), Notifier, notifierEntry

### Community 10 - "CookieJar"
Cohesion: 0.08
Nodes (45): Cookie, CookieJar, CookieFileExists(), NewCookieJar(), NewCookieJarWithEncryption(), TestAutoMigrationPlaintextToEncrypted(), TestEncryptedCookieWithoutKey(), testEncryptionKey() (+37 more)

### Community 11 - "Manager"
Cohesion: 0.10
Nodes (9): Handler, Manager, twitch.Client, NewManager(), NewHandler(), twitch.PrivateMessage, twitch.UserJoinMessage, twitch.UserNoticeMessage (+1 more)

### Community 12 - "What You Must Do When Invoked"
Cohesion: 0.07
Nodes (26): For /graphify add and --watch, For /graphify query, For the commit hook and native CLAUDE.md integration, For --update and --cluster-only, /graphify, Honesty Rules, Interpreter guard for subcommands, Part A - Structural extraction for code files (+18 more)

### Community 13 - "context.Context"
Cohesion: 0.07
Nodes (11): context.Context, encoding/json.RawMessage, ChannelPointsContext, GameResp, GoalContribution, PlaybackAccessToken, TeamMember, TopStream (+3 more)

### Community 14 - "Client"
Cohesion: 0.09
Nodes (10): sync.Map, sync.RWMutex, LookupGameSlug(), RegisterGameSlug(), Client, isStaleUsernameError(), TestIsStaleUsernameError(), spadeCache (+2 more)

### Community 15 - "Batcher"
Cohesion: 0.31
Nodes (4): batchKey, sync.Once, batchEntry, Batcher

### Community 16 - "install-service.sh"
Cohesion: 0.18
Nodes (28): ask(), banner(), confirm(), DEFAULT_CONFIG_DIR, DEFAULT_DATA_DIR, DEFAULT_ENV_FILE, DEFAULT_INSTALL_DIR, DEFAULT_LOG_LEVEL (+20 more)

### Community 18 - "net/http.ResponseWriter"
Cohesion: 0.20
Nodes (11): net/http.Request, net/http.ResponseWriter, AnalyticsServer, AnalyticsServer, AnalyticsServer, filterStreamers(), AnalyticsServer, parsePagination() (+3 more)

### Community 19 - "CommunityGoal"
Cohesion: 0.12
Nodes (20): BoolFromMap(), FloatFromAny(), IntFromAny(), IntFromMap(), StringFromAny(), StringFromMap(), TestBoolFromMap(), TestFloatFromAny() (+12 more)

### Community 20 - "newMockTransport"
Cohesion: 0.10
Nodes (64): bytes.Buffer, NewForTest(), NewClientForTest(), boolPtr(), Client, makeStreamerWithDrop(), makeStreamerWithDropAtProgress(), milestoneEventCount() (+56 more)

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
Nodes (16): github.com/coder/websocket.Conn, Provider, PubSubTopic, NewStreamerTopic(), NewUserTopic(), Connection, NewConnection(), Connection (+8 more)

### Community 26 - "main"
Cohesion: 0.16
Nodes (18): main(), playStartupAnimation(), resolveFileWatchInterval(), resolveLogDir(), resolveLogFormat(), resolveLogLevel(), resolveLogNoTime(), resolveNoBanner() (+10 more)

### Community 27 - "Campaign"
Cohesion: 0.16
Nodes (7): Campaign, NewCampaign(), BenchmarkParseCampaign(), campaignMatchesStreamer(), Client, isQuarterMilestone(), parseCampaign()

### Community 28 - "Configuration Reference"
Cohesion: 0.11
Nodes (18): `batch` (global batching defaults), `bet.filter_condition` (optional), `bet` (nested under `streamer_defaults` and per-streamer `settings`), `blacklist`, `category_blacklist`, `category_watcher`, Configuration Reference, `features` (+10 more)

### Community 29 - "Drop"
Cohesion: 0.25
Nodes (4): Percentage(), Percentage(), TestPercentage(), Drop

### Community 30 - "tray.go"
Cohesion: 0.19
Nodes (13): applyNoConsole(), startTray(), fyne.io/systray.MenuItem, RunServiceAction(), ServiceScriptAvailable(), serviceScriptName(), Available(), clickCh() (+5 more)

### Community 31 - "middleware.go"
Cohesion: 0.22
Nodes (11): net/http.Handler, generateRequestID(), newRateLimiter(), RequestIDFromContext(), withLogging(), withRateLimit(), withRequestID(), contextKey (+3 more)

### Community 32 - "newTestConfigServer"
Cohesion: 0.30
Nodes (15): net/http/httptest.ResponseRecorder, configRequest(), AnalyticsServer, newTestConfigServer(), TestConfigGenerate_InvalidConfig(), TestConfigGenerate_InvalidJSON(), TestConfigGenerate_MissingUsername(), TestConfigGenerate_Valid() (+7 more)

### Community 33 - "AnalyticsServer"
Cohesion: 0.19
Nodes (11): net/http.Server, checkCredentials(), withAuth(), AnalyticsServer, NewAnalyticsServer(), AuthStatusFunc, DashboardAuth, DebugSnapshotFunc (+3 more)

### Community 34 - "account.go"
Cohesion: 0.16
Nodes (13): CategoryConfig, DiscordConfig, GotifyConfig, jsonDuration, MatrixConfig, PushoverConfig, TeamConfig, TelegramConfig (+5 more)

### Community 35 - ".handlePredictionCreated"
Cohesion: 0.26
Nodes (3): Miner, isTerminalPredictionStatus(), isTransientPredictionError()

### Community 36 - "static/app.js"
Cohesion: 0.30
Nodes (14): buildFilterParams(), clearFilters(), debounce(), escapeHTML(), fetchJSON(), formatPoints(), initFilterListeners(), loadFilters() (+6 more)

### Community 37 - "logs.js"
Cohesion: 0.27
Nodes (13): buildFilterParams(), clearFilters(), escapeHTML(), fetchJSON(), formatPoints(), handleSort(), loadFilters(), populateSelect() (+5 more)

### Community 38 - "buildBody"
Cohesion: 0.19
Nodes (13): sync/atomic.Int64, buildBody(), TestBuildBodyBetGeneralFallsBackToMessage(), TestBuildBodyBetKeepsUnknownFields(), TestBuildBodyBetResultNamesChannel(), TestBuildBodyBetStartShowsCountdown(), TestBuildBodyNonBetKeepsCompactForm(), TestBuildBodyNonBetWithoutFields() (+5 more)

### Community 39 - "Getting Started"
Cohesion: 0.15
Nodes (13): 1. Clone and configure, 2. Set required environment variables, 3. Run, 4. Authenticate, 5. Verify it's working, Automatic updates, Getting Started, Going deeper (+5 more)

### Community 41 - ".pollForToken"
Cohesion: 0.22
Nodes (5): DeviceCodeResponse, TokenErrorResponse, TokenResponse, Authenticator, DeviceCodeStatus

### Community 42 - "NewBatcher"
Cohesion: 0.49
Nodes (10): NewBatcher(), boolPtr(), newTestLogger(), TestBatcher_BuffersAndFlushes(), TestBatcher_ImmediateEventsBypassBuffer(), TestBatcher_MaxEntriesSplitsMessages(), TestBatcher_SingleEntryNotBatched(), TestBatcher_StopFlushesRemaining() (+2 more)

### Community 44 - "AccountRow"
Cohesion: 0.05
Nodes (22): database/sql.DB, NewPoller(), minimalConfigJSON(), newTestPoller(), TestPoller_DisabledAccountStopsMiner(), TestPoller_InvalidConfigSkipped(), TestPoller_MultipleAccounts(), TestPoller_NewAccountStartsMiner() (+14 more)

### Community 45 - ".handle2FA"
Cohesion: 0.42
Nodes (3): loginResponse, Authenticator, promptLine()

### Community 46 - "Manager"
Cohesion: 0.24
Nodes (5): context.CancelFunc, Manager, NewManager(), NewMiner(), entry

### Community 47 - "Provider setup"
Cohesion: 0.17
Nodes (12): Discord, Event reference, Gotify, Matrix, Notification batching, Notifications, Per-account vs global credentials, Provider setup (+4 more)

### Community 48 - "Prediction Strategies"
Cohesion: 0.17
Nodes (12): Betting amount calculation, Delay modes, Filter conditions, `HIGH_ODDS`, `MOST_VOTED`, `NUMBER_1` through `NUMBER_8`, `PERCENTAGE`, Prediction Strategies (+4 more)

### Community 49 - "Logger"
Cohesion: 0.22
Nodes (10): os.File, sync/atomic.Value, DefaultConfig(), Logger, NotifyFunc, ParseLevel(), Setup(), TestLogFileNameEncodesFullStartupTimestamp() (+2 more)

### Community 50 - "Miner"
Cohesion: 0.13
Nodes (5): time.Timer, Miner, oneTimeEventMessage(), streakDir(), API

### Community 51 - "manager_test.go"
Cohesion: 0.31
Nodes (14): newLiveManager(), newTestManager(), testConfig(), TestManager_EntriesReturnsSnapshot(), TestManager_LiveCountCountsStartingMiners(), TestManager_LiveCountExcludesExitedMiners(), TestManager_LiveCountZeroWithoutEntries(), TestManager_RestartReplacesEntry() (+6 more)

### Community 52 - "newTestAccountsServer"
Cohesion: 0.28
Nodes (22): accountsRequest(), AnalyticsServer, minimalAccountCfg(), newFakeStore(), newTestAccountsServer(), TestCreateAccount_InvalidJSON(), TestCreateAccount_MissingUsername(), TestCreateAccount_NoStore() (+14 more)

### Community 53 - "settings.go"
Cohesion: 0.47
Nodes (4): AllEvents(), ParseEvent(), ParseFollowersOrder(), FollowersOrder

### Community 54 - "Contributing"
Cohesion: 0.33
Nodes (6): Automated Versioning, Commit Convention, Contributing, Documentation and Wiki, Pull Requests, Setting Up Git Hooks

### Community 55 - "Event"
Cohesion: 0.10
Nodes (9): net/http.Client, Event, isBetEvent(), baseNotifier, Discord, Gotify, Pushover, Telegram (+1 more)

### Community 56 - "newTestFileWatcher"
Cohesion: 0.24
Nodes (13): NewFileWatcher(), newTestFileWatcher(), TestFileWatcher_ChangedMtimeRestartsMiner(), TestFileWatcher_EmptyDirNoStarts(), TestFileWatcher_InvalidYAMLSkipped(), TestFileWatcher_MinimalYAMLValid(), TestFileWatcher_MultipleAccounts(), TestFileWatcher_NewFileStartsMiner() (+5 more)

### Community 57 - "newTestMiner"
Cohesion: 0.14
Nodes (29): Miner, newTestMiner(), raidMessage(), TestHandleRaid_Dedup(), TestHandleRaid_EmptyRaidID(), TestHandleRaid_FollowRaidDisabled(), TestHandleRaid_MissingRaidData(), TestHandleRaid_NoStreamer() (+21 more)

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

### Community 68 - "server_test.go"
Cohesion: 0.33
Nodes (17): apiRequest(), newTestServer(), TestCreateAccount(), TestCreateAccountConflict(), TestCreateAccountInvalidName(), TestCreateAccountValidationFail(), TestDeleteAccount(), TestDeleteAccountNotFound() (+9 more)

### Community 69 - "Stream"
Cohesion: 0.15
Nodes (5): time.Duration, StreamInfoResponse, GameInfo, Stream, Tag

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
Cohesion: 0.22
Nodes (20): editAccountFields, applyCategoryWatcherSection(), applyEditFields(), applyFeaturesSection(), applyStreamersSection(), applyTeamWatcherSection(), boolVal(), intVal() (+12 more)

### Community 78 - "CategoryWatcher"
Cohesion: 0.50
Nodes (3): CategoryWatcher, NewCategoryWatcher(), categoryEntry

### Community 79 - "Validate"
Cohesion: 0.17
Nodes (20): fatalf(), main(), net/url.URL, AccountConfigFromJSON(), applyDefaults(), applyEnvOverrides(), getEnv(), AccountConfig (+12 more)

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

### Community 104 - "findProc"
Cohesion: 0.23
Nodes (9): golang.org/x/sys/windows.LazyProc, findProc(), HideConsole(), showConsole(), TestConsoleToggleStateStartsVisible(), TestGetConsoleWindowResolves(), TestShowConsoleWithoutHideIsNoop(), TestShowWindowResolves() (+1 more)

### Community 106 - "testing.B"
Cohesion: 0.22
Nodes (14): AccountConfig, testing.B, AccountConfigToJSON(), benchConfig(), BenchmarkAccountConfigFromJSON(), BenchmarkAccountConfigToJSON(), BenchmarkValidate(), T (+6 more)

### Community 107 - "checkWithURL"
Cohesion: 0.08
Nodes (38): runAutoUpdate(), log/slog.Logger, detectDeployment(), LoadConfigFromEnv(), loadOrGenerateInstanceID(), NewSender(), newUUID(), CheckForUpdate() (+30 more)

### Community 111 - "streams_api.go"
Cohesion: 0.18
Nodes (15): HistoryEntry, applyPagination(), T, campaignInfo, dropInfo, dropProgressShort, errorResponse, eventLogEntry (+7 more)

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

### Community 129 - "Streamer"
Cohesion: 0.10
Nodes (20): Priority, ParsePriority(), Streamer, anyStreakPending(), applyPriority(), applyPriorityDrops(), applyPriorityEndingSoonest(), applyPriorityLowAvailability() (+12 more)

### Community 130 - "Required Configuration"
Cohesion: 0.33
Nodes (6): Account Configuration, Authentication, Getting Browser Values, Getting TV Client ID, Required Configuration, Twitch Runtime Identifiers

### Community 131 - "Configuration (Fly.io)"
Cohesion: 0.33
Nodes (6): Configuration (Fly.io), Fly.io, Monitoring, Quick Start (Fly.io), Scaling, Secrets Management

### Community 132 - "newTestLogger"
Cohesion: 0.36
Nodes (8): AnalyticsServer, healthResponse(), TestHealthDegradedWhenNoMinersRunning(), TestHealthOKWhenMinerCountUnknown(), TestHealthOKWhenMinersRunning(), TestPprofNotRegisteredWithoutAuth(), TestPprofRequiresAuthWhenRegistered(), newTestLogger()

### Community 133 - "SelectWatchSet"
Cohesion: 0.44
Nodes (15): WatchSet, SelectWatchSet(), rotationOptions(), streakPending(), streakSettled(), TestPreferredChannelsTakeSlotsInListedOrder(), TestPreferredMatchingIsCaseInsensitive(), TestStreakSlotsGoToChannelsClosestToTheirStreak() (+7 more)

### Community 134 - "Linux Service (systemd / OpenRC)"
Cohesion: 0.40
Nodes (5): File Locations (defaults), Linux Service (systemd / OpenRC), Managing the Service, Quick Start, Uninstalling

### Community 135 - "sync.Mutex"
Cohesion: 0.17
Nodes (14): sync.Mutex, mockTransport, assertVarEquals(), assertVarPresent(), Client, newTestGQLClient(), TestGetChannelPointsContext_SendsChannelLogin(), TestGetPlaybackAccessToken_PlatformIsNonEmptyString() (+6 more)

### Community 136 - "AccountConfig"
Cohesion: 0.22
Nodes (7): FeaturesConfig, FollowersConfig, StreamerConfig, StreamerSettingsConfig, AccountConfig, AuthConfig, fakeMgr

### Community 137 - "DebugSnapshot"
Cohesion: 0.21
Nodes (7): Miner, Connection, ConnectionSnapshot, Pool, DebugPredictionEntry, DebugSnapshot, DebugWatchingEntry

### Community 138 - "clientWithStreaks"
Cohesion: 0.35
Nodes (9): clientWithStreaks(), Client, streamerWithBroadcast(), TestRestoreClearsFlagForKnownBroadcast(), TestRestoreDoesNotQueryWhenStreakAlreadySettled(), TestRestoreIsNoOpWithoutAStore(), TestRestoreKeepsChasingNewBroadcast(), TestRestoreKeepsChasingUnknownChannel() (+1 more)

### Community 139 - "TeamWatcher"
Cohesion: 0.18
Nodes (4): Miner, pollLoop(), TeamWatcher, NewTeamWatcher()

### Community 140 - "time.Time"
Cohesion: 0.17
Nodes (10): time.Time, serverTime(), EventPrediction, collectOnlineIndices(), Client, FloatRound(), formatFloat(), Millify() (+2 more)

### Community 142 - "testing.T"
Cohesion: 0.13
Nodes (25): testing.T, TestNewBetComputesDerivedOutcomeValues(), TestParseCondition(), TestParseStrategy(), TestStrategy_String(), TestNoopStore_ChangesIsNil(), TestNoopStore_DeleteIsNoop(), TestNoopStore_GetAccountReturnsNil() (+17 more)

### Community 143 - "1.6. Environment Variables"
Cohesion: 0.29
Nodes (7): 1.6.1. Global, 1.6.2. Per-Account Authentication, 1.6.3. Notification Secrets, 1.6.4.1. How To Obtain Twitch Runtime Identifiers, 1.6.4. `.env` File Support, 1.6.5. Cookie Encryption (optional), 1.6. Environment Variables

### Community 144 - "capturingNotifier"
Cohesion: 0.53
Nodes (3): newCapturingNotifier(), TestNotifyFuncDeliversChannelInBetBody(), capturingNotifier

### Community 145 - "1.7. Notifications"
Cohesion: 0.33
Nodes (6): 1.7.1. Supported Providers, 1.7.2. Example: Telegram, 1.7.3. Event Filtering, 1.7.4. Notification Batching, 1.7.5. Testing Notifications, 1.7. Notifications

### Community 146 - "autostart_unix.go"
Cohesion: 0.30
Nodes (12): autostartDirPath(), AutostartEnabled(), autostartFilePath(), ClearAutostart(), escapePlist(), launchAgentPath(), quote(), SetAutostart() (+4 more)

### Community 147 - "NewAuthenticator"
Cohesion: 0.21
Nodes (10): generateDeviceID(), GenerateHex(), NewAuthenticator(), envOrDefault(), Twitch, LoadTwitchFromEnv(), TestClientIDsForGQL_Dedup(), TestLoadTwitchFromEnv_Defaults() (+2 more)

### Community 148 - "1.12. Deploy to Fly.io"
Cohesion: 0.40
Nodes (5): 1.12.1. Setup, 1.12.2. CI/CD Auto-Deploy, 1.12.3. Manual Deploy, 1.12.4. Alternative Deployment, 1.12. Deploy to Fly.io

### Community 149 - "1.5. Configuration"
Cohesion: 0.50
Nodes (4): 1.5.1. Quick Start, 1.5.2. Config Editor, 1.5.3. Database Mode (optional), 1.5. Configuration

### Community 150 - "1.11. Windows Service"
Cohesion: 0.67
Nodes (3): 1.11.1. Managing the Service, 1.11.2. Uninstalling, 1.11. Windows Service

### Community 151 - "newTestConnection"
Cohesion: 0.43
Nodes (7): Connection, newTestConnection(), TestHandleResponse_ERR_BADAUTH_AlreadyRefreshedByAnother(), TestHandleResponse_ERR_BADAUTH_RefreshesAndResubscribes(), TestHandleResponse_ERR_BADAUTH_RefreshFailsNoResubscribe(), TestHandleResponse_OtherErrors_NoRefresh(), TestHandleResponse_ReconnectClosesConnection()

### Community 152 - "Field"
Cohesion: 0.71
Nodes (6): Field, buildBetBody(), buildCompactBody(), fieldValue(), formatChannel(), formatPick()

### Community 153 - "NewStream"
Cohesion: 0.33
Nodes (10): NewStream(), TestHealthyChannelIsNotStalled(), TestMarkMinuteWatchAttemptStamps(), TestMinuteWatchedAccumulatesAcrossSegments(), TestMinuteWatchedCountsContinuousWatching(), TestMinuteWatchedIgnoresUnwatchedGaps(), TestNeverAttemptedIsNotStalled(), TestNeverCreditedIsNotStalled() (+2 more)

### Community 154 - "NewServer"
Cohesion: 0.28
Nodes (4): main(), openBrowser(), net/http.HandlerFunc, NewServer()

### Community 155 - "NewStreamer"
Cohesion: 0.54
Nodes (7): DefaultStreamerSettings(), NewStreamer(), TestSetOnline_AlreadyOnline_NoOp(), TestSetOnline_Carryover_FirstOnline(), TestSetOnline_Carryover_StreakResolved_LongGap(), TestSetOnline_Carryover_StreakResolved_ShortGap(), TestSetOnline_Carryover_StreakUnresolved_ShortGap()

### Community 156 - "StreamerSettings"
Cohesion: 0.48
Nodes (4): StreamerSettings, ParseChatPresence(), ShouldJoinChat(), ChatPresence

## Knowledge Gaps
- **317 isolated node(s):** `_run-localdev.sh script`, `TWITCH_MINER_RUN_LOCALDEV`, `_run.sh script`, `github.com/Guliveer/twitch-miner-go`, `TokenErrorResponse` (+312 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 573 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **27 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Streamer` connect `Streamer` to `prediction.go`, `SelectWatchSet`, `clientWithStreaks`, `TeamWatcher`, `time.Time`, `context.Context`, `Client`, `Miner`, `Message`, `net/http.ResponseWriter`, `CommunityGoal`, `newMockTransport`, `SelectStreamersToWatch`, `Connection`, `Campaign`, `NewStreamer`, `StreamerSettings`, `AnalyticsServer`, `.handlePredictionCreated`, `Miner`, `newTestMiner`, `Stream`, `Raid`, `CategoryWatcher`, `streams_api.go`?**
  _High betweenness centrality (0.100) - this node is a cross-community bridge._
- **Why does `Logger` connect `Logger` to `newTestLogger`, `Client`, `Dispatcher`, `Manager`, `TeamWatcher`, `Client`, `Batcher`, `net/http.ResponseWriter`, `NewAuthenticator`, `newMockTransport`, `Connection`, `main`, `tray.go`, `middleware.go`, `AnalyticsServer`, `NewBatcher`, `AccountRow`, `Manager`, `Miner`, `Event`, `newTestFileWatcher`, `CategoryWatcher`, `Validate`, `Authenticator`, `checkWithURL`?**
  _High betweenness centrality (0.079) - this node is a cross-community bridge._
- **Why does `Miner` connect `Miner` to `Streamer`, `sync.Mutex`, `AccountConfig`, `Dispatcher`, `.pollForToken`, `Manager`, `time.Time`, `TeamWatcher`, `Client`, `CategoryWatcher`, `Manager`, `Logger`, `NewAuthenticator`, `Event`, `Connection`?**
  _High betweenness centrality (0.049) - this node is a cross-community bridge._
- **Are the 23 inferred relationships involving `newMockTransport()` (e.g. with `TestLogDropProgress_MixedPrintable()` and `TestLogDropProgress_NilPreconditions()`) actually correct?**
  _`newMockTransport()` has 23 INFERRED edges - model-reasoned connections that need verification._
- **What connects `_run-localdev.sh script`, `TWITCH_MINER_RUN_LOCALDEV`, `_run.sh script` to the rest of the system?**
  _317 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `web/app.js` be split into smaller, more focused modules?**
  _Cohesion score 0.0825508607198748 - nodes in this community are weakly interconnected._
- **Should `prediction.go` be split into smaller, more focused modules?**
  _Cohesion score 0.12903225806451613 - nodes in this community are weakly interconnected._