# Graph Report - twitch-miner-go  (2026-10-12)

## Corpus Check
- 206 files · ~188,457 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2346 nodes · 5481 edges · 148 communities (107 shown, 24 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 413 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `0c33f348`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- NewDrop
- web/app.js
- prediction.go
- Deployment Guide
- testing.T
- DefaultBetSettings
- twitch-miner-go - Efficient Auto Drops & Points Claim for Twitch
- badge_catalog.go
- Client
- Dispatcher
- CookieJar
- sync.Mutex
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
- makeOnlineStreamer
- Connection
- main
- Campaign
- Configuration Reference
- testBadgeWatcher
- autostart_unix.go
- Sender
- poller_test.go
- AnalyticsServer
- AccountConfig
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
- poller_test.go
- Provider setup
- Prediction Strategies
- Logger
- Miner
- Poller
- cookies_test.go
- time.Duration
- Commit Types Cheatsheet
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
- 1.10. Linux Service (systemd / OpenRC)
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
- checkWithURL
- pre-commit
- mockAuthProvider
- pre-push
- Code of Conduct
- _run-localdev.sh
- edit-config.sh
- gen-cookie-key.sh
- gen-dashboard-auth.sh
- install-hooks.sh
- CI/CD Pipelines
- Configuration (Docker)
- Windows Service
- Streamer
- Required Configuration
- Configuration (Fly.io)
- SelectWatchSet
- operations_test.go
- DebugSnapshot
- testBadgeWatcher
- TeamWatcher
- time.Time
- Miner
- Open
- 1.6. Environment Variables
- capturingNotifier
- 1.7. Notifications
- Twitch
- 1.12. Deploy to Fly.io
- 1.5. Configuration
- 1.11. Windows Service
- Poller
- NewStream
- NewStreamer

## God Nodes (most connected - your core abstractions)
1. `Streamer` - 118 edges
2. `Logger` - 55 edges
3. `newMockTransport()` - 44 edges
4. `DefaultBetSettings()` - 41 edges
5. `Miner` - 40 edges
6. `makeBet()` - 36 edges
7. `Event` - 31 edges
8. `Connection` - 31 edges
9. `newTestClientWithCapture()` - 31 edges
10. `main()` - 29 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `NewServer()`  [EXTRACTED]
  cmd/config-editor/main.go → internal/configeditor/server.go
- `main()` --calls--> `OpenPostgres()`  [EXTRACTED]
  cmd/db-seed/main.go → internal/store/postgres.go
- `main()` --calls--> `applyNoConsole()`  [INFERRED]
  cmd/twitch-miner-go/main.go → cmd/twitch-miner-go/tray_enabled.go
- `main()` --calls--> `ParseKey()`  [EXTRACTED]
  cmd/twitch-miner-go/main.go → internal/encryption/encryption.go
- `main()` --calls--> `ColorSupported()`  [EXTRACTED]
  cmd/twitch-miner-go/main.go → internal/logger/color_support.go

## Import Cycles
- None detected.

## Communities (148 total, 24 thin omitted)

### Community 0 - "NewDrop"
Cohesion: 0.27
Nodes (19): NewDrop(), refTime(), TestDropUpdate_FieldUpdate_DropInstanceID(), TestDropUpdate_FieldUpdate_HasPreconditionsMet(), TestDropUpdate_FieldUpdate_IsClaimed(), TestDropUpdate_IsClaimable_Claimed(), TestDropUpdate_IsClaimable_EmptyInstanceID(), TestDropUpdate_IsClaimable_WithInstanceID() (+11 more)

### Community 1 - "web/app.js"
Cohesion: 0.08
Nodes (69): addCategoryItem(), addStreamerItem(), addTag(), addTeamItem(), api(), assignFloatFromEl(), assignNum(), assignNumFromEl() (+61 more)

### Community 2 - "prediction.go"
Cohesion: 0.07
Nodes (36): BuildDefaults(), BuildSchemaPayload(), T, stringValues(), TestBuildSchemaPayload_ContainsAllChatModes(), TestBuildSchemaPayload_ContainsAllDelayModes(), TestBuildSchemaPayload_ContainsAllEvents(), TestBuildSchemaPayload_ContainsAllFollowersOrders() (+28 more)

### Community 3 - "Deployment Guide"
Cohesion: 0.12
Nodes (14): Comparison Matrix, Deployment Guide, Deployment Options, File Locations (defaults), Linux Service (systemd / OpenRC), Managing the Service, Next Steps, Quick Start (+6 more)

### Community 4 - "testing.T"
Cohesion: 0.14
Nodes (29): testing.T, ParseMessage(), splitTopic(), TestMessage_String(), TestParseMessage_ChannelIDFallbackToTopicUser(), TestParseMessage_ChannelIDFromDataChannelID(), TestParseMessage_ClaimAvailable(), TestParseMessage_EmptyObject() (+21 more)

### Community 5 - "DefaultBetSettings"
Cohesion: 0.14
Nodes (42): DefaultBetSettings(), Outcome, NewBet(), NewEventPrediction(), TestNewBetComputesDerivedOutcomeValues(), makeBet(), TestCalculate_AllOutcomesZeroUsers(), TestCalculate_AmountCappedByMaxPoints() (+34 more)

### Community 6 - "twitch-miner-go - Efficient Auto Drops & Points Claim for Twitch"
Cohesion: 0.11
Nodes (19): 1.13. Development, 1.14.1. Automatic updates, 1.14. Auto-Update, 1.15. License, 1.16. FAQ, 1.17. Advanced Guide, 1.1. Table of Contents, 1.2. Features (+11 more)

### Community 7 - "badge_catalog.go"
Cohesion: 0.06
Nodes (54): TopStream, badgeCampaignAllowsChannel(), badgeCampaignGameMatches(), badgeStreamerValid(), appendUnique(), appendUniqueBadgeNames(), badgeRequiresPayment(), badgeTitlesForWatchReward() (+46 more)

### Community 8 - "Client"
Cohesion: 0.11
Nodes (16): net/http.Transport, circuitBreaker, gqlError, gqlExtensions, gqlRequest, gqlResponse, operationBehavior, persistedQuery (+8 more)

### Community 9 - "Dispatcher"
Cohesion: 0.19
Nodes (7): sync/atomic.Bool, ParseEvent(), Dispatcher, NewDispatcher(), parseEvents(), Notifier, notifierEntry

### Community 10 - "CookieJar"
Cohesion: 0.07
Nodes (47): Cookie, CookieJar, generateDeviceID(), GenerateHex(), NewAuthenticator(), NewCookieJar(), NewCookieJarWithEncryption(), TestAutoMigrationPlaintextToEncrypted() (+39 more)

### Community 11 - "sync.Mutex"
Cohesion: 0.10
Nodes (9): Handler, Manager, twitch.Client, NewManager(), NewHandler(), twitch.PrivateMessage, twitch.UserJoinMessage, twitch.UserNoticeMessage (+1 more)

### Community 12 - "What You Must Do When Invoked"
Cohesion: 0.07
Nodes (26): For /graphify add and --watch, For /graphify query, For the commit hook and native CLAUDE.md integration, For --update and --cluster-only, /graphify, Honesty Rules, Interpreter guard for subcommands, Part A - Structural extraction for code files (+18 more)

### Community 13 - "context.Context"
Cohesion: 0.07
Nodes (10): context.Context, encoding/json.RawMessage, GameResp, GoalContribution, PlaybackAccessToken, TeamMember, Client, StreamInfoResponse (+2 more)

### Community 14 - "Client"
Cohesion: 0.09
Nodes (9): sync.Map, sync.RWMutex, LookupGameSlug(), RegisterGameSlug(), Client, isStaleUsernameError(), spadeCache, spadeCacheEntry (+1 more)

### Community 15 - "Batcher"
Cohesion: 0.36
Nodes (3): batchKey, batchEntry, Batcher

### Community 16 - "install-service.sh"
Cohesion: 0.18
Nodes (28): ask(), banner(), confirm(), DEFAULT_CONFIG_DIR, DEFAULT_DATA_DIR, DEFAULT_ENV_FILE, DEFAULT_INSTALL_DIR, DEFAULT_LOG_LEVEL (+20 more)

### Community 17 - "Message"
Cohesion: 0.17
Nodes (6): extractNestedInt(), Miner, mapReasonToEvent(), parseCommunityGoal(), Message, MessageType

### Community 18 - "net/http.ResponseWriter"
Cohesion: 0.06
Nodes (48): accountMeta, Server, io.Reader, net/http.HandlerFunc, net/http.Request, net/http.Response, net/http.ResponseWriter, net/http.ServeMux (+40 more)

### Community 19 - "CommunityGoal"
Cohesion: 0.13
Nodes (17): ChannelPointsContext, BoolFromMap(), FloatFromAny(), IntFromAny(), IntFromMap(), StringFromMap(), TestBoolFromMap(), TestFloatFromAny() (+9 more)

### Community 20 - "newMockTransport"
Cohesion: 0.09
Nodes (67): bytes.Buffer, sync.Mutex, NewForTest(), NewClientForTest(), NewCampaign(), boolPtr(), Client, makeStreamerWithDrop() (+59 more)

### Community 22 - "colorHandler"
Cohesion: 0.18
Nodes (11): resolveLogLevel(), io.Writer, log/slog.Attr, log/slog.Handler, log/slog.Level, log/slog.Record, copyAttrs(), newColorHandler() (+3 more)

### Community 23 - "Troubleshooting"
Cohesion: 0.07
Nodes (27): `401 Unauthorized` errors in logs, All bets on one outcome keep losing, `authenticated as "X" but config expects "Y"`, Authentication errors, Config changes not taking effect, Config editor & tray issues, Config issues, Drop issues (+19 more)

### Community 24 - "makeOnlineStreamer"
Cohesion: 0.24
Nodes (22): SelectStreamersToWatch(), makeOnlineStreamer(), TestSelectStreamersToWatch_DropsOnly_IncludedWhenHasCampaignIDs(), TestSelectStreamersToWatch_DropsOnly_SkippedWhenNoCampaignIDs(), TestSelectStreamersToWatch_EndingSoonest(), TestSelectStreamersToWatch_EndingSoonest_PicksFirstEnd(), TestSelectStreamersToWatch_EndingSoonest_SkipsNoCampaigns(), TestSelectStreamersToWatch_Freeze_AllFrozenReturnsNone() (+14 more)

### Community 25 - "Connection"
Cohesion: 0.06
Nodes (23): github.com/coder/websocket.Conn, Provider, PubSubTopic, NewStreamerTopic(), NewUserTopic(), Connection, NewConnection(), Connection (+15 more)

### Community 26 - "main"
Cohesion: 0.18
Nodes (17): main(), playStartupAnimation(), resolveFileWatchInterval(), resolveLogDir(), resolveLogFormat(), resolveLogNoTime(), resolveNoBanner(), resolveNoTray() (+9 more)

### Community 27 - "Campaign"
Cohesion: 0.16
Nodes (6): Campaign, BenchmarkParseCampaign(), campaignMatchesStreamer(), Client, isQuarterMilestone(), parseCampaign()

### Community 28 - "Configuration Reference"
Cohesion: 0.11
Nodes (19): `badge_watcher`, `batch` (global batching defaults), `bet.filter_condition` (optional), `bet` (nested under `streamer_defaults` and per-streamer `settings`), `blacklist`, `category_blacklist`, `category_watcher`, Configuration Reference (+11 more)

### Community 29 - "testBadgeWatcher"
Cohesion: 0.24
Nodes (19): activeBadgeCampaign(), TestBadgeStreamerUsesPreloadedCampaignIDs(), TestBadgeStreamerValidRequiresActiveDropsForAllChannels(), testBadgeWatcher(), TestBadgeWatcherAllChannelsFallsBackAndSelectsEligibleCampaign(), TestBadgeWatcherAllChannelsSearchDoesNotRequireDropsTag(), TestBadgeWatcherAllChannelsSkipsLookupErrorsAndIneligibleStreams(), TestBadgeWatcherAllChannelsStopsAtFirstEligibleTaggedStream() (+11 more)

### Community 30 - "autostart_unix.go"
Cohesion: 0.06
Nodes (37): applyNoConsole(), startTray(), fyne.io/systray.MenuItem, Percentage(), autostartDirPath(), AutostartEnabled(), autostartFilePath(), ClearAutostart() (+29 more)

### Community 31 - "Sender"
Cohesion: 0.26
Nodes (9): log/slog.Logger, detectDeployment(), LoadConfigFromEnv(), loadOrGenerateInstanceID(), NewSender(), newUUID(), Config, heartbeatPayload (+1 more)

### Community 32 - "poller_test.go"
Cohesion: 0.30
Nodes (15): net/http/httptest.ResponseRecorder, configRequest(), AnalyticsServer, newTestConfigServer(), TestConfigGenerate_InvalidConfig(), TestConfigGenerate_InvalidJSON(), TestConfigGenerate_MissingUsername(), TestConfigGenerate_Valid() (+7 more)

### Community 33 - "AnalyticsServer"
Cohesion: 0.09
Nodes (30): net/http.Handler, net/http.Server, AnalyticsServer, healthResponse(), TestHealthDegradedWhenNoMinersRunning(), TestHealthOKWhenMinerCountUnknown(), TestHealthOKWhenMinersRunning(), TestPprofNotRegisteredWithoutAuth() (+22 more)

### Community 34 - "AccountConfig"
Cohesion: 0.05
Nodes (47): BetSettingsConfig, CategoryConfig, DiscordConfig, FeaturesConfig, FilterConditionConfig, FollowersConfig, GotifyConfig, jsonDuration (+39 more)

### Community 35 - ".handlePredictionCreated"
Cohesion: 0.23
Nodes (5): StringFromAny(), parseOutcomes(), Miner, isTerminalPredictionStatus(), isTransientPredictionError()

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
Cohesion: 0.07
Nodes (9): database/sql.DB, sync.Once, OpenPostgres(), scanRow(), AccountRow, fakePollerStore, fakeStore, NoopStore (+1 more)

### Community 45 - ".handle2FA"
Cohesion: 0.42
Nodes (3): loginResponse, Authenticator, promptLine()

### Community 46 - "poller_test.go"
Cohesion: 0.44
Nodes (10): minimalConfigJSON(), newTestPoller(), TestPoller_DisabledAccountStopsMiner(), TestPoller_InvalidConfigSkipped(), TestPoller_MultipleAccounts(), TestPoller_NewAccountStartsMiner(), TestPoller_RemovedAccountStopsMiner(), TestPoller_SameTimestampNoRestart() (+2 more)

### Community 47 - "Provider setup"
Cohesion: 0.17
Nodes (12): Discord, Event reference, Gotify, Matrix, Notification batching, Notifications, Per-account vs global credentials, Provider setup (+4 more)

### Community 48 - "Prediction Strategies"
Cohesion: 0.17
Nodes (12): Betting amount calculation, Delay modes, Filter conditions, `HIGH_ODDS`, `MOST_VOTED`, `NUMBER_1` through `NUMBER_8`, `PERCENTAGE`, Prediction Strategies (+4 more)

### Community 49 - "Logger"
Cohesion: 0.18
Nodes (11): os.File, sync/atomic.Value, NewClient(), DefaultConfig(), Logger, NotifyFunc, Setup(), TestLogFileNameEncodesFullStartupTimestamp() (+3 more)

### Community 50 - "Miner"
Cohesion: 0.13
Nodes (7): time.Timer, Miner, NewMiner(), oneTimeEventMessage(), streakDir(), ShouldJoinChat(), API

### Community 51 - "Poller"
Cohesion: 0.33
Nodes (4): NewPoller(), Store, minerManager, Poller

### Community 52 - "cookies_test.go"
Cohesion: 0.28
Nodes (22): accountsRequest(), AnalyticsServer, minimalAccountCfg(), newFakeStore(), newTestAccountsServer(), TestCreateAccount_InvalidJSON(), TestCreateAccount_MissingUsername(), TestCreateAccount_NoStore() (+14 more)

### Community 53 - "time.Duration"
Cohesion: 0.46
Nodes (4): time.Duration, NewFileWatcher(), fileManager, FileWatcher

### Community 54 - "Commit Types Cheatsheet"
Cohesion: 0.05
Nodes (40): Automated Versioning, `build` – Build Process or Dependencies, `chore` – Maintenance and Routine Tasks, `ci` – Continuous Integration, Commit Convention, Commit Types Cheatsheet, Contributing, `docs` – Documentation (+32 more)

### Community 55 - "Event"
Cohesion: 0.10
Nodes (10): net/http.Client, Event, isBetEvent(), fetchJSON(), baseNotifier, Discord, Gotify, Pushover (+2 more)

### Community 56 - "newTestFileWatcher"
Cohesion: 0.44
Nodes (10): newTestFileWatcher(), TestFileWatcher_ChangedMtimeRestartsMiner(), TestFileWatcher_EmptyDirNoStarts(), TestFileWatcher_InvalidYAMLSkipped(), TestFileWatcher_MinimalYAMLValid(), TestFileWatcher_MultipleAccounts(), TestFileWatcher_NewFileStartsMiner(), TestFileWatcher_RemovedFileStopsMiner() (+2 more)

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

### Community 65 - "1.10. Linux Service (systemd / OpenRC)"
Cohesion: 0.50
Nodes (4): 1.10.1. Managing the Service, 1.10.2. Uninstalling, 1.10.3. Default File Locations, 1.10. Linux Service (systemd / OpenRC)

### Community 66 - "Authentication"
Cohesion: 0.29
Nodes (7): Authentication, Cookie persistence, Obtaining an OAuth token, Priority chain, Recommended approach per environment, Token scopes required, Variable naming convention

### Community 67 - "FAQ"
Cohesion: 0.29
Nodes (7): Accounts & configuration, Behaviour & resources, Config editor & tray, Docker & Fly.io, FAQ, Running, Telemetry & data

### Community 68 - "server_test.go"
Cohesion: 0.30
Nodes (18): apiRequest(), newTestServer(), TestCleanConfig(), TestCreateAccount(), TestCreateAccountConflict(), TestCreateAccountInvalidName(), TestCreateAccountValidationFail(), TestDeleteAccount() (+10 more)

### Community 69 - "Stream"
Cohesion: 0.16
Nodes (3): GameInfo, Stream, Tag

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
Cohesion: 0.23
Nodes (4): Miner, CategoryWatcher, NewCategoryWatcher(), categoryEntry

### Community 79 - "Validate"
Cohesion: 0.10
Nodes (35): AccountConfig, fatalf(), main(), net/url.URL, testing.B, AccountConfigFromJSON(), AccountConfigToJSON(), applyDefaults() (+27 more)

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

### Community 89 - "Authenticator"
Cohesion: 0.12
Nodes (3): Authenticator, CookieFileExists(), AuthConfig

### Community 104 - "findProc"
Cohesion: 0.23
Nodes (9): golang.org/x/sys/windows.LazyProc, findProc(), HideConsole(), showConsole(), TestConsoleToggleStateStartsVisible(), TestGetConsoleWindowResolves(), TestShowConsoleWithoutHideIsNoop(), TestShowWindowResolves() (+1 more)

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

### Community 129 - "Streamer"
Cohesion: 0.11
Nodes (19): Streamer, anyStreakPending(), applyPriority(), applyPriorityBadges(), applyPriorityDrops(), applyPriorityEndingSoonest(), applyPriorityLowAvailability(), applyPriorityOrder() (+11 more)

### Community 130 - "Required Configuration"
Cohesion: 0.33
Nodes (6): Account Configuration, Authentication, Getting Browser Values, Getting TV Client ID, Required Configuration, Twitch Runtime Identifiers

### Community 131 - "Configuration (Fly.io)"
Cohesion: 0.33
Nodes (6): Configuration (Fly.io), Fly.io, Monitoring, Quick Start (Fly.io), Scaling, Secrets Management

### Community 133 - "SelectWatchSet"
Cohesion: 0.42
Nodes (16): WatchSet, SelectWatchSet(), rotationOptions(), streakPending(), streakSettled(), TestBadgePriorityWinsWatchSlot(), TestPreferredChannelsTakeSlotsInListedOrder(), TestPreferredMatchingIsCaseInsensitive() (+8 more)

### Community 135 - "operations_test.go"
Cohesion: 0.40
Nodes (10): mockTransport, assertVarEquals(), assertVarPresent(), Client, newTestGQLClient(), TestGetChannelPointsContext_SendsChannelLogin(), TestGetPlaybackAccessToken_PlatformIsNonEmptyString(), TestGetPlaybackAccessToken_SendAllRequiredVariables() (+2 more)

### Community 137 - "DebugSnapshot"
Cohesion: 0.21
Nodes (7): Miner, Connection, ConnectionSnapshot, Pool, DebugPredictionEntry, DebugSnapshot, DebugWatchingEntry

### Community 138 - "testBadgeWatcher"
Cohesion: 0.35
Nodes (9): clientWithStreaks(), Client, streamerWithBroadcast(), TestRestoreClearsFlagForKnownBroadcast(), TestRestoreDoesNotQueryWhenStreakAlreadySettled(), TestRestoreIsNoOpWithoutAStore(), TestRestoreKeepsChasingNewBroadcast(), TestRestoreKeepsChasingUnknownChannel() (+1 more)

### Community 140 - "time.Time"
Cohesion: 0.13
Nodes (10): time.Time, serverTime(), EventPrediction, Store, collectOnlineIndices(), Client, FloatRound(), PredictionResult (+2 more)

### Community 142 - "Open"
Cohesion: 0.33
Nodes (10): Open(), TestAccountNameIsCaseInsensitive(), TestCorruptFileYieldsUsableEmptyStore(), TestEarnedIsScopedToTheBroadcast(), TestExpiredRecordsAreDroppedOnReopen(), TestMissingFileOpensEmpty(), TestRecordLeavesNoTempFileBehind(), TestRecordSurvivesReopen() (+2 more)

### Community 143 - "1.6. Environment Variables"
Cohesion: 0.29
Nodes (7): 1.6.1. Global, 1.6.2. Per-Account Authentication, 1.6.3. Notification Secrets, 1.6.4.1. How To Obtain Twitch Runtime Identifiers, 1.6.4. `.env` File Support, 1.6.5. Cookie Encryption (optional), 1.6. Environment Variables

### Community 144 - "capturingNotifier"
Cohesion: 0.53
Nodes (3): newCapturingNotifier(), TestNotifyFuncDeliversChannelInBetBody(), capturingNotifier

### Community 145 - "1.7. Notifications"
Cohesion: 0.33
Nodes (6): 1.7.1. Supported Providers, 1.7.2. Example: Telegram, 1.7.3. Event Filtering, 1.7.4. Notification Batching, 1.7.5. Testing Notifications, 1.7. Notifications

### Community 147 - "Twitch"
Cohesion: 0.31
Nodes (6): envOrDefault(), Twitch, LoadTwitchFromEnv(), TestClientIDsForGQL_Dedup(), TestLoadTwitchFromEnv_Defaults(), TestLoadTwitchFromEnv_EnvOverride()

### Community 148 - "1.12. Deploy to Fly.io"
Cohesion: 0.40
Nodes (5): 1.12.1. Setup, 1.12.2. CI/CD Auto-Deploy, 1.12.3. Manual Deploy, 1.12.4. Alternative Deployment, 1.12. Deploy to Fly.io

### Community 149 - "1.5. Configuration"
Cohesion: 0.50
Nodes (4): 1.5.1. Quick Start, 1.5.2. Config Editor, 1.5.3. Database Mode (optional), 1.5. Configuration

### Community 150 - "1.11. Windows Service"
Cohesion: 0.67
Nodes (3): 1.11.1. Managing the Service, 1.11.2. Uninstalling, 1.11. Windows Service

### Community 152 - "Poller"
Cohesion: 0.71
Nodes (6): Field, buildBetBody(), buildCompactBody(), fieldValue(), formatChannel(), formatPick()

### Community 153 - "NewStream"
Cohesion: 0.33
Nodes (10): NewStream(), TestHealthyChannelIsNotStalled(), TestMarkMinuteWatchAttemptStamps(), TestMinuteWatchedAccumulatesAcrossSegments(), TestMinuteWatchedCountsContinuousWatching(), TestMinuteWatchedIgnoresUnwatchedGaps(), TestNeverAttemptedIsNotStalled(), TestNeverCreditedIsNotStalled() (+2 more)

### Community 155 - "NewStreamer"
Cohesion: 0.27
Nodes (10): DefaultStreamerSettings(), StreamerSettings, NewStreamer(), ParseChatPresence(), TestSetOnline_AlreadyOnline_NoOp(), TestSetOnline_Carryover_FirstOnline(), TestSetOnline_Carryover_StreakResolved_LongGap(), TestSetOnline_Carryover_StreakResolved_ShortGap() (+2 more)

## Knowledge Gaps
- **342 isolated node(s):** `_run-localdev.sh script`, `TWITCH_MINER_RUN_LOCALDEV`, `_run.sh script`, `github.com/Guliveer/twitch-miner-go`, `TokenErrorResponse` (+337 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 601 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **24 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Logger` connect `Logger` to `badge_catalog.go`, `Client`, `Dispatcher`, `CookieJar`, `sync.Mutex`, `TeamWatcher`, `Client`, `Batcher`, `net/http.ResponseWriter`, `newMockTransport`, `Connection`, `main`, `autostart_unix.go`, `Sender`, `AnalyticsServer`, `AccountConfig`, `NewBatcher`, `Miner`, `Poller`, `time.Duration`, `Event`, `CategoryWatcher`, `Authenticator`, `checkWithURL`?**
  _High betweenness centrality (0.074) - this node is a cross-community bridge._
- **Why does `Streamer` connect `Streamer` to `DefaultBetSettings`, `SelectWatchSet`, `badge_catalog.go`, `testBadgeWatcher`, `TeamWatcher`, `time.Time`, `Miner`, `Client`, `context.Context`, `Message`, `net/http.ResponseWriter`, `CommunityGoal`, `newMockTransport`, `makeOnlineStreamer`, `Connection`, `Campaign`, `NewStreamer`, `AnalyticsServer`, `.handlePredictionCreated`, `Miner`, `newTestMiner`, `Stream`, `Raid`, `CategoryWatcher`?**
  _High betweenness centrality (0.061) - this node is a cross-community bridge._
- **Why does `Connection` connect `Connection` to `time.Time`, `AccountRow`, `Message`, `Logger`, `newMockTransport`?**
  _High betweenness centrality (0.032) - this node is a cross-community bridge._
- **Are the 23 inferred relationships involving `newMockTransport()` (e.g. with `TestLogDropProgress_MixedPrintable()` and `TestLogDropProgress_NilPreconditions()`) actually correct?**
  _`newMockTransport()` has 23 INFERRED edges - model-reasoned connections that need verification._
- **What connects `_run-localdev.sh script`, `TWITCH_MINER_RUN_LOCALDEV`, `_run.sh script` to the rest of the system?**
  _342 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `web/app.js` be split into smaller, more focused modules?**
  _Cohesion score 0.0825508607198748 - nodes in this community are weakly interconnected._
- **Should `prediction.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07272727272727272 - nodes in this community are weakly interconnected._