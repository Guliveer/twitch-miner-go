# Graph Report - twitch-miner-go  (2026-11-16)

## Corpus Check
- 217 files · ~197,496 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2445 nodes · 5773 edges · 160 communities (120 shown, 24 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 459 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `fba7f59b`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- jsonDuration
- web/app.js
- newTestConfigServer
- Deployment Guide
- ParseMessage
- DefaultBetSettings
- twitch-miner-go - Efficient Auto Drops & Points Claim for Twitch
- BadgeWatcher
- Client
- Dispatcher
- encryption_test.go
- colorHandler
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
- prediction.go
- Troubleshooting
- makeOnlineStreamer
- Connection
- main
- time.Duration
- Configuration Reference
- testing.T
- autostart_unix.go
- clientWithStreaks
- poller_test.go
- middleware.go
- account.go
- operations_test.go
- static/app.js
- logs.js
- NewDrop
- Getting Started
- newTestAccountsServer
- .pollForToken
- NewStreamer
- AccountRow
- .handle2FA
- Authenticator
- Provider setup
- Prediction Strategies
- tokenChangingMock
- Miner
- Miner
- Poller
- manager_test.go
- Commit Types Cheatsheet
- Event
- NewBatcher
- newTestMiner
- Modern Dark
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
- BuildSchemaPayload
- cookies_test.go
- graphify reference: query, path, explain
- autostart_windows.go
- Twitch Channel Points Miner — Go Edition
- AGENTS.md
- os/exec.Cmd
- tui.go
- _run.sh
- SafeGo
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
- ResolveBatchConfig
- graphify.js
- rules/graphify.md
- workflows/graphify.md
- extraction-spec.md
- Operations
- github.com/Guliveer/twitch-miner-go
- buildBody
- findProc
- commit-msg
- .handlePredictionCreated
- checkWithURL
- pre-commit
- mockAuthProvider
- pre-push
- CookieJar
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
- AnalyticsServer
- badge_catalog.go
- Open
- Raid
- Logger
- time.Time
- testing.B
- NewAnalyticsServer
- Twitch
- Miner
- sync.Mutex
- 1.6. Environment Variables
- capturingNotifier
- 1.7. Notifications
- Field
- Manager
- 1.12. Deploy to Fly.io
- 1.5. Configuration
- 1.11. Windows Service
- AccountConfig
- loadBadgeCampaigns
- Handler
- newTestRateLimiter
- Manager
- Store
- ParseKey
- newTestConnection
- campaignCompleted

## God Nodes (most connected - your core abstractions)
1. `Streamer` - 126 edges
2. `Logger` - 55 edges
3. `newMockTransport()` - 47 edges
4. `DefaultBetSettings()` - 41 edges
5. `Miner` - 40 edges
6. `makeBet()` - 36 edges
7. `makeOnlineStreamer()` - 32 edges
8. `Event` - 31 edges
9. `Connection` - 31 edges
10. `newTestClientWithCapture()` - 31 edges

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

## Communities (160 total, 24 thin omitted)

### Community 0 - "jsonDuration"
Cohesion: 0.15
Nodes (5): CategoryConfig, jsonDuration, TeamConfig, CategoryWatcherConfig, TeamWatcherConfig

### Community 1 - "web/app.js"
Cohesion: 0.08
Nodes (70): addCategoryItem(), addStreamerItem(), addTag(), addTeamItem(), api(), assignFloatFromEl(), assignNum(), assignNumFromEl() (+62 more)

### Community 2 - "newTestConfigServer"
Cohesion: 0.33
Nodes (14): configRequest(), AnalyticsServer, newTestConfigServer(), TestConfigGenerate_InvalidConfig(), TestConfigGenerate_InvalidJSON(), TestConfigGenerate_MissingUsername(), TestConfigGenerate_Valid(), TestConfigGenerate_YAMLContainsAllFields() (+6 more)

### Community 3 - "Deployment Guide"
Cohesion: 0.12
Nodes (14): Comparison Matrix, Deployment Guide, Deployment Options, File Locations (defaults), Linux Service (systemd / OpenRC), Managing the Service, Next Steps, Quick Start (+6 more)

### Community 4 - "ParseMessage"
Cohesion: 0.15
Nodes (21): ParseMessage(), splitTopic(), TestMessage_String(), TestParseMessage_ChannelIDFallbackToTopicUser(), TestParseMessage_ChannelIDFromDataChannelID(), TestParseMessage_ClaimAvailable(), TestParseMessage_EmptyObject(), TestParseMessage_Identifier() (+13 more)

### Community 5 - "DefaultBetSettings"
Cohesion: 0.17
Nodes (38): DefaultBetSettings(), makeBet(), TestCalculate_AllOutcomesZeroUsers(), TestCalculate_AmountCappedByMaxPoints(), TestCalculate_AmountFormula(), TestCalculate_EqualOdds(), TestCalculate_HighOdds(), TestCalculate_MostVoted() (+30 more)

### Community 6 - "twitch-miner-go - Efficient Auto Drops & Points Claim for Twitch"
Cohesion: 0.11
Nodes (19): 1.13. Development, 1.14.1. Automatic updates, 1.14. Auto-Update, 1.15. License, 1.16. FAQ, 1.17. Advanced Guide, 1.1. Table of Contents, 1.2. Features (+11 more)

### Community 7 - "BadgeWatcher"
Cohesion: 0.17
Nodes (14): BadgeWatcherConfig, TopStream, badgeCampaignAllowsChannel(), badgeCampaignGameMatches(), badgeStreamerValid(), findStreamer(), BadgeWatcher, NewBadgeWatcher() (+6 more)

### Community 8 - "Client"
Cohesion: 0.10
Nodes (17): net/http.Transport, circuitBreaker, gqlError, gqlExtensions, gqlRequest, gqlResponse, operationBehavior, persistedQuery (+9 more)

### Community 9 - "Dispatcher"
Cohesion: 0.19
Nodes (7): sync/atomic.Bool, ParseEvent(), Dispatcher, NewDispatcher(), parseEvents(), Notifier, notifierEntry

### Community 10 - "encryption_test.go"
Cohesion: 0.24
Nodes (18): Decrypt(), decryptAES256GCM(), Encrypt(), IsEncrypted(), Parse(), TestDecryptCorruptedPayload(), TestDecryptInvalidBase64(), TestDecryptMalformedFormat() (+10 more)

### Community 11 - "colorHandler"
Cohesion: 0.21
Nodes (9): io.Writer, log/slog.Attr, log/slog.Handler, log/slog.Level, log/slog.Record, copyAttrs(), newColorHandler(), colorHandler (+1 more)

### Community 12 - "What You Must Do When Invoked"
Cohesion: 0.07
Nodes (26): For /graphify add and --watch, For /graphify query, For the commit hook and native CLAUDE.md integration, For --update and --cluster-only, /graphify, Honesty Rules, Interpreter guard for subcommands, Part A - Structural extraction for code files (+18 more)

### Community 13 - "context.Context"
Cohesion: 0.07
Nodes (10): context.Context, encoding/json.RawMessage, GameResp, GoalContribution, PlaybackAccessToken, TeamMember, Client, StreamInfoResponse (+2 more)

### Community 14 - "Client"
Cohesion: 0.09
Nodes (11): sync.Map, sync.RWMutex, LookupGameSlug(), RegisterGameSlug(), Client, isStaleUsernameError(), NewClient(), TestIsStaleUsernameError() (+3 more)

### Community 15 - "Batcher"
Cohesion: 0.31
Nodes (4): batchKey, sync.Once, batchEntry, Batcher

### Community 16 - "install-service.sh"
Cohesion: 0.18
Nodes (28): ask(), banner(), confirm(), DEFAULT_CONFIG_DIR, DEFAULT_DATA_DIR, DEFAULT_ENV_FILE, DEFAULT_INSTALL_DIR, DEFAULT_LOG_LEVEL (+20 more)

### Community 18 - "net/http.ResponseWriter"
Cohesion: 0.06
Nodes (48): accountMeta, Server, io.Reader, net/http.HandlerFunc, net/http.Request, net/http.Response, net/http.ResponseWriter, net/http.ServeMux (+40 more)

### Community 19 - "CommunityGoal"
Cohesion: 0.11
Nodes (19): ChannelPointsContext, BoolFromMap(), FloatFromAny(), IntFromAny(), IntFromMap(), StringFromMap(), TestBoolFromMap(), TestFloatFromAny() (+11 more)

### Community 20 - "newMockTransport"
Cohesion: 0.09
Nodes (66): bytes.Buffer, NewForTest(), NewClientForTest(), boolPtr(), Client, makeStreamerWithDrop(), makeStreamerWithDropAtProgress(), milestoneEventCount() (+58 more)

### Community 22 - "prediction.go"
Cohesion: 0.10
Nodes (22): StringFromAny(), TestStringFromAny(), parseOutcomes(), GetPredictionWindow(), BetSettings, Outcome, NewBet(), NewEventPrediction() (+14 more)

### Community 23 - "Troubleshooting"
Cohesion: 0.07
Nodes (27): `401 Unauthorized` errors in logs, All bets on one outcome keep losing, `authenticated as "X" but config expects "Y"`, Authentication errors, Config changes not taking effect, Config editor & tray issues, Config issues, Drop issues (+19 more)

### Community 24 - "makeOnlineStreamer"
Cohesion: 0.09
Nodes (52): Miner, Connection, ConnectionSnapshot, Pool, SelectStreamersToWatch(), SelectWatchSet(), makeOnlineStreamer(), TestSelectStreamersToWatch_DropsOnly_IncludedWhenHasCampaignIDs() (+44 more)

### Community 25 - "Connection"
Cohesion: 0.07
Nodes (16): github.com/coder/websocket.Conn, Provider, PubSubTopic, NewStreamerTopic(), NewUserTopic(), Connection, NewConnection(), Connection (+8 more)

### Community 26 - "main"
Cohesion: 0.19
Nodes (17): main(), playStartupAnimation(), resolveFileWatchInterval(), resolveLogDir(), resolveLogFormat(), resolveLogLevel(), resolveLogNoTime(), resolveNoBanner() (+9 more)

### Community 27 - "time.Duration"
Cohesion: 0.05
Nodes (30): log/slog.Logger, time.Duration, Campaign, NewCampaign(), GameInfo, Stream, Tag, NewStream() (+22 more)

### Community 28 - "Configuration Reference"
Cohesion: 0.10
Nodes (20): `badge_watcher`, `batch` (global batching defaults), `bet.filter_condition` (optional), `bet` (nested under `streamer_defaults` and per-streamer `settings`), `blacklist`, `category_blacklist`, `category_watcher`, Configuration Reference (+12 more)

### Community 29 - "testing.T"
Cohesion: 0.18
Nodes (28): testing.T, TestNoopStore_ChangesIsNil(), TestNoopStore_DeleteIsNoop(), TestNoopStore_GetAccountReturnsNil(), TestNoopStore_ListAccountsReturnsEmpty(), TestNoopStore_UpsertIsNoop(), TestBadgeCampaignNextRewardUsesTightestSlack(), TestBadgeWatcherRefreshesDeadlineFromInventory() (+20 more)

### Community 30 - "autostart_unix.go"
Cohesion: 0.08
Nodes (38): applyNoConsole(), startTray(), fyne.io/systray.MenuItem, net/url.Values, BugReportURL(), Expand(), FeatureRequestURL(), issueURL() (+30 more)

### Community 31 - "clientWithStreaks"
Cohesion: 0.35
Nodes (9): clientWithStreaks(), Client, streamerWithBroadcast(), TestRestoreClearsFlagForKnownBroadcast(), TestRestoreDoesNotQueryWhenStreakAlreadySettled(), TestRestoreIsNoOpWithoutAStore(), TestRestoreKeepsChasingNewBroadcast(), TestRestoreKeepsChasingUnknownChannel() (+1 more)

### Community 32 - "poller_test.go"
Cohesion: 0.39
Nodes (11): discardLogger(), minimalConfigJSON(), newTestPoller(), TestPoller_DisabledAccountStopsMiner(), TestPoller_InvalidConfigSkipped(), TestPoller_MultipleAccounts(), TestPoller_NewAccountStartsMiner(), TestPoller_RemovedAccountStopsMiner() (+3 more)

### Community 33 - "middleware.go"
Cohesion: 0.18
Nodes (14): net/http.Handler, checkCredentials(), generateRequestID(), RequestIDFromContext(), withAuth(), withLogging(), withRateLimit(), withRequestID() (+6 more)

### Community 34 - "account.go"
Cohesion: 0.29
Nodes (12): BetSettingsConfig, DiscordConfig, FilterConditionConfig, GotifyConfig, MatrixConfig, PushoverConfig, StreamerConfig, StreamerSettingsConfig (+4 more)

### Community 35 - "operations_test.go"
Cohesion: 0.31
Nodes (12): mockTransport, assertVarEquals(), assertVarPresent(), Client, newTestGQLClient(), TestGetChannelPointsContext_SendsChannelLogin(), TestGetFollowedStreamers_FailsWhenDataMissing(), TestGetFollowedStreamers_UsesDataAlongsidePartialErrors() (+4 more)

### Community 36 - "static/app.js"
Cohesion: 0.30
Nodes (14): buildFilterParams(), clearFilters(), debounce(), escapeHTML(), fetchJSON(), formatPoints(), initFilterListeners(), loadFilters() (+6 more)

### Community 37 - "logs.js"
Cohesion: 0.27
Nodes (13): buildFilterParams(), clearFilters(), escapeHTML(), fetchJSON(), formatPoints(), handleSort(), loadFilters(), populateSelect() (+5 more)

### Community 38 - "NewDrop"
Cohesion: 0.17
Nodes (21): Drop, NewDrop(), Percentage(), refTime(), TestDropUpdate_FieldUpdate_DropInstanceID(), TestDropUpdate_FieldUpdate_HasPreconditionsMet(), TestDropUpdate_FieldUpdate_IsClaimed(), TestDropUpdate_IsClaimable_Claimed() (+13 more)

### Community 39 - "Getting Started"
Cohesion: 0.15
Nodes (13): 1. Clone and configure, 2. Set required environment variables, 3. Run, 4. Authenticate, 5. Verify it's working, Automatic updates, Getting Started, Going deeper (+5 more)

### Community 40 - "newTestAccountsServer"
Cohesion: 0.28
Nodes (22): accountsRequest(), AnalyticsServer, minimalAccountCfg(), newFakeStore(), newTestAccountsServer(), TestCreateAccount_InvalidJSON(), TestCreateAccount_MissingUsername(), TestCreateAccount_NoStore() (+14 more)

### Community 41 - ".pollForToken"
Cohesion: 0.31
Nodes (4): DeviceCodeResponse, TokenErrorResponse, TokenResponse, Authenticator

### Community 42 - "NewStreamer"
Cohesion: 0.10
Nodes (19): Miner, AllChatModes(), DefaultStreamerSettings(), StreamerSettings, NewStreamer(), ParseChatPresence(), ShouldJoinChat(), TestSetOnline_AlreadyOnline_NoOp() (+11 more)

### Community 44 - "AccountRow"
Cohesion: 0.07
Nodes (8): database/sql.DB, OpenPostgres(), scanRow(), AccountRow, fakePollerStore, fakeStore, NoopStore, PostgresStore

### Community 45 - ".handle2FA"
Cohesion: 0.42
Nodes (3): loginResponse, Authenticator, promptLine()

### Community 46 - "Authenticator"
Cohesion: 0.10
Nodes (5): generateDeviceID(), GenerateHex(), Authenticator, NewAuthenticator(), DeviceCodeStatus

### Community 47 - "Provider setup"
Cohesion: 0.17
Nodes (12): Discord, Event reference, Gotify, Matrix, Notification batching, Notifications, Per-account vs global credentials, Provider setup (+4 more)

### Community 48 - "Prediction Strategies"
Cohesion: 0.17
Nodes (12): Betting amount calculation, Delay modes, Filter conditions, `HIGH_ODDS`, `MOST_VOTED`, `NUMBER_1` through `NUMBER_8`, `PERCENTAGE`, Prediction Strategies (+4 more)

### Community 50 - "Miner"
Cohesion: 0.12
Nodes (6): time.Timer, Miner, NewMiner(), oneTimeEventMessage(), streakDir(), API

### Community 51 - "Miner"
Cohesion: 0.26
Nodes (3): Miner, isTerminalPredictionStatus(), isTransientPredictionError()

### Community 52 - "Poller"
Cohesion: 0.33
Nodes (4): NewPoller(), Store, minerManager, Poller

### Community 53 - "manager_test.go"
Cohesion: 0.31
Nodes (14): newLiveManager(), newTestManager(), testConfig(), TestManager_EntriesReturnsSnapshot(), TestManager_LiveCountCountsStartingMiners(), TestManager_LiveCountExcludesExitedMiners(), TestManager_LiveCountZeroWithoutEntries(), TestManager_RestartReplacesEntry() (+6 more)

### Community 54 - "Commit Types Cheatsheet"
Cohesion: 0.05
Nodes (40): Automated Versioning, `build` – Build Process or Dependencies, `chore` – Maintenance and Routine Tasks, `ci` – Continuous Integration, Commit Convention, Commit Types Cheatsheet, Contributing, `docs` – Documentation (+32 more)

### Community 55 - "Event"
Cohesion: 0.10
Nodes (10): net/http.Client, Event, isBetEvent(), fetchJSON(), baseNotifier, Discord, Gotify, Pushover (+2 more)

### Community 56 - "NewBatcher"
Cohesion: 0.49
Nodes (10): NewBatcher(), boolPtr(), newTestLogger(), TestBatcher_BuffersAndFlushes(), TestBatcher_ImmediateEventsBypassBuffer(), TestBatcher_MaxEntriesSplitsMessages(), TestBatcher_SingleEntryNotBatched(), TestBatcher_StopFlushesRemaining() (+2 more)

### Community 57 - "newTestMiner"
Cohesion: 0.13
Nodes (30): Miner, newTestMiner(), raidMessage(), TestCheckAllStreamersOnline_AppliesChatPresence(), TestHandleRaid_Dedup(), TestHandleRaid_EmptyRaidID(), TestHandleRaid_FollowRaidDisabled(), TestHandleRaid_MissingRaidData() (+22 more)

### Community 58 - "Modern Dark"
Cohesion: 0.17
Nodes (11): Authoritative tokens, Color and surfaces, Components, Design limits, Identity, Images, Layout, Modern Dark (+3 more)

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

### Community 69 - "BuildSchemaPayload"
Cohesion: 0.16
Nodes (21): BuildDefaults(), BuildSchemaPayload(), T, stringValues(), TestBuildSchemaPayload_ContainsAllChatModes(), TestBuildSchemaPayload_ContainsAllDelayModes(), TestBuildSchemaPayload_ContainsAllEvents(), TestBuildSchemaPayload_ContainsAllFollowersOrders() (+13 more)

### Community 70 - "cookies_test.go"
Cohesion: 0.31
Nodes (16): NewCookieJar(), NewCookieJarWithEncryption(), TestAutoMigrationPlaintextToEncrypted(), TestEncryptedCookieWithoutKey(), testEncryptionKey(), TestGetReturnsEmptyForMissingCookie(), TestHasEncryption(), TestLoadEncryptedCookies() (+8 more)

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
Cohesion: 0.14
Nodes (13): Before writing code, CI/CD verification, Code quality, Comments, Completion checklist, Dependencies, Design, Documentation (+5 more)

### Community 75 - "os/exec.Cmd"
Cohesion: 0.40
Nodes (3): os/exec.Cmd, detach(), detach()

### Community 76 - "tui.go"
Cohesion: 0.18
Nodes (22): main(), openBrowser(), editAccountFields, applyCategoryWatcherSection(), applyEditFields(), applyFeaturesSection(), applyStreamersSection(), applyTeamWatcherSection() (+14 more)

### Community 78 - "SafeGo"
Cohesion: 0.17
Nodes (11): formatFloat(), Millify(), OpenBrowser(), Percentage(), SafeGo(), Slugify(), TestFloatRound(), TestMillify() (+3 more)

### Community 79 - "Validate"
Cohesion: 0.08
Nodes (43): AccountConfig, fatalf(), main(), net/url.URL, AccountConfigFromJSON(), AccountConfigToJSON(), applyDefaults(), applyEnvOverrides() (+35 more)

### Community 80 - "CLAUDE.md"
Cohesion: 0.14
Nodes (13): Before writing code, CI/CD verification, Code quality, Comments, Completion checklist, Dependencies, Design, Documentation (+5 more)

### Community 81 - ".claude/CLAUDE.md"
Cohesion: 0.14
Nodes (13): Before writing code, CI/CD verification, Code quality, Comments, Completion checklist, Dependencies, Design, Documentation (+5 more)

### Community 82 - "GEMINI.md"
Cohesion: 0.14
Nodes (13): Before writing code, CI/CD verification, Code quality, Comments, Completion checklist, Dependencies, Design, Documentation (+5 more)

### Community 83 - "copilot-instructions.md"
Cohesion: 0.14
Nodes (13): Before writing code, CI/CD verification, Code quality, Comments, Completion checklist, Dependencies, Design, Documentation (+5 more)

### Community 84 - "graphify reference: add a URL and watch a folder"
Cohesion: 0.50
Nodes (3): For /graphify add, For --watch, graphify reference: add a URL and watch a folder

### Community 85 - "graphify reference: commit hook and native CLAUDE.md integration"
Cohesion: 0.50
Nodes (3): For git commit hook, For native CLAUDE.md integration, graphify reference: commit hook and native CLAUDE.md integration

### Community 86 - "graphify reference: incremental update and cluster-only"
Cohesion: 0.50
Nodes (3): For --cluster-only, For --update (incremental re-extraction), graphify reference: incremental update and cluster-only

### Community 89 - "ResolveBatchConfig"
Cohesion: 0.46
Nodes (7): ResolveBatchConfig(), boolPtr(), TestIsBatchEnabled_Nil(), TestResolveBatchConfig_BothNil(), TestResolveBatchConfig_GlobalOnly(), TestResolveBatchConfig_ProviderDisablesGlobal(), TestResolveBatchConfig_ProviderOverridesGlobal()

### Community 98 - "buildBody"
Cohesion: 0.19
Nodes (13): sync/atomic.Int64, buildBody(), TestBuildBodyBetGeneralFallsBackToMessage(), TestBuildBodyBetKeepsUnknownFields(), TestBuildBodyBetResultNamesChannel(), TestBuildBodyBetStartShowsCountdown(), TestBuildBodyNonBetKeepsCompactForm(), TestBuildBodyNonBetWithoutFields() (+5 more)

### Community 104 - "findProc"
Cohesion: 0.23
Nodes (9): golang.org/x/sys/windows.LazyProc, findProc(), HideConsole(), showConsole(), TestConsoleToggleStateStartsVisible(), TestGetConsoleWindowResolves(), TestShowConsoleWithoutHideIsNoop(), TestShowWindowResolves() (+1 more)

### Community 107 - "checkWithURL"
Cohesion: 0.10
Nodes (30): runAutoUpdate(), CheckForUpdate(), checkWithURL(), DownloadAsset(), ExitForRestart(), findAssetURL(), FormatNotification(), isGitRepo() (+22 more)

### Community 111 - "CookieJar"
Cohesion: 0.19
Nodes (3): Cookie, CookieJar, CookieFileExists()

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
Nodes (24): Priority, ParsePriority(), Streamer, anyStreakPending(), applyPriority(), applyPriorityBadges(), applyPriorityDrops(), applyPriorityEndingSoonest() (+16 more)

### Community 130 - "Required Configuration"
Cohesion: 0.33
Nodes (6): Account Configuration, Authentication, Getting Browser Values, Getting TV Client ID, Required Configuration, Twitch Runtime Identifiers

### Community 131 - "Configuration (Fly.io)"
Cohesion: 0.33
Nodes (6): Configuration (Fly.io), Fly.io, Monitoring, Quick Start (Fly.io), Scaling, Secrets Management

### Community 132 - "AnalyticsServer"
Cohesion: 0.21
Nodes (8): net/http.Server, AnalyticsServer, AuthStatusFunc, AuthStatusListFunc, DebugSnapshotFunc, MinerCountFunc, NotifyTestFunc, StreamerFunc

### Community 133 - "badge_catalog.go"
Cohesion: 0.18
Nodes (20): appendUnique(), appendUniqueBadgeNames(), badgeRequiresPayment(), badgeTitlesForWatchReward(), canonicalGameSlug(), comparableWords(), equalWords(), isBadgeReward() (+12 more)

### Community 134 - "Open"
Cohesion: 0.33
Nodes (10): Open(), TestAccountNameIsCaseInsensitive(), TestCorruptFileYieldsUsableEmptyStore(), TestEarnedIsScopedToTheBroadcast(), TestExpiredRecordsAreDroppedOnReopen(), TestMissingFileOpensEmpty(), TestRecordLeavesNoTempFileBehind(), TestRecordSurvivesReopen() (+2 more)

### Community 136 - "Logger"
Cohesion: 0.18
Nodes (9): os.File, sync/atomic.Value, DefaultConfig(), Logger, NotifyFunc, ParseLevel(), Setup(), TestLogFileNameEncodesFullStartupTimestamp() (+1 more)

### Community 137 - "time.Time"
Cohesion: 0.25
Nodes (12): time.Time, serverTime(), RewardDeadline, applyUrgentReward(), availableDropDeadline(), campaignRewardStart(), preferUrgentReward(), streamerRewardStart() (+4 more)

### Community 138 - "testing.B"
Cohesion: 0.48
Nodes (6): testing.B, BenchmarkBetCalculate(), BenchmarkFilterConditionSkip(), BenchmarkParseMessage(), BenchmarkUpdateHistory(), BenchmarkUpdateOutcomes()

### Community 139 - "NewAnalyticsServer"
Cohesion: 0.25
Nodes (15): net/http/httptest.ResponseRecorder, decodeObject(), TestAuthStatusesEmptyWhenFuncReturnsNil(), TestAuthStatusesEmptyWithoutFunc(), TestAuthStatusesReturnsPendingFlows(), TestAuthStatusesRouteIsRegistered(), AnalyticsServer, healthResponse() (+7 more)

### Community 140 - "Twitch"
Cohesion: 0.31
Nodes (6): envOrDefault(), Twitch, LoadTwitchFromEnv(), TestClientIDsForGQL_Dedup(), TestLoadTwitchFromEnv_Defaults(), TestLoadTwitchFromEnv_EnvOverride()

### Community 142 - "sync.Mutex"
Cohesion: 0.24
Nodes (5): sync.Mutex, EventPrediction, Client, FloatRound(), PredictionResult

### Community 143 - "1.6. Environment Variables"
Cohesion: 0.29
Nodes (7): 1.6.1. Global, 1.6.2. Per-Account Authentication, 1.6.3. Notification Secrets, 1.6.4.1. How To Obtain Twitch Runtime Identifiers, 1.6.4. `.env` File Support, 1.6.5. Cookie Encryption (optional), 1.6. Environment Variables

### Community 144 - "capturingNotifier"
Cohesion: 0.53
Nodes (3): newCapturingNotifier(), TestNotifyFuncDeliversChannelInBetBody(), capturingNotifier

### Community 145 - "1.7. Notifications"
Cohesion: 0.33
Nodes (6): 1.7.1. Supported Providers, 1.7.2. Example: Telegram, 1.7.3. Event Filtering, 1.7.4. Notification Batching, 1.7.5. Testing Notifications, 1.7. Notifications

### Community 146 - "Field"
Cohesion: 0.71
Nodes (6): Field, buildBetBody(), buildCompactBody(), fieldValue(), formatChannel(), formatPick()

### Community 147 - "Manager"
Cohesion: 0.26
Nodes (4): context.CancelFunc, Manager, NewManager(), entry

### Community 148 - "1.12. Deploy to Fly.io"
Cohesion: 0.40
Nodes (5): 1.12.1. Setup, 1.12.2. CI/CD Auto-Deploy, 1.12.3. Manual Deploy, 1.12.4. Alternative Deployment, 1.12. Deploy to Fly.io

### Community 149 - "1.5. Configuration"
Cohesion: 0.50
Nodes (4): 1.5.1. Quick Start, 1.5.2. Config Editor, 1.5.3. Database Mode (optional), 1.5. Configuration

### Community 150 - "1.11. Windows Service"
Cohesion: 0.67
Nodes (3): 1.11.1. Managing the Service, 1.11.2. Uninstalling, 1.11. Windows Service

### Community 151 - "AccountConfig"
Cohesion: 0.25
Nodes (5): FeaturesConfig, FollowersConfig, AccountConfig, AuthConfig, fakeMgr

### Community 152 - "loadBadgeCampaigns"
Cohesion: 0.16
Nodes (16): loadBadgeCampaigns(), ownsCampaignBadge(), TestBadgeRequiresPaymentUsesWholeWords(), TestBadgeRewardMatching(), TestLoadBadgeCampaignsAcceptsSingleChannel(), TestLoadBadgeCampaignsDoesNotConfusePaidBadgeWithWatchCampaign(), TestLoadBadgeCampaignsKeepsCleanMatchWhenAnotherDropIsAmbiguous(), TestLoadBadgeCampaignsMatchesCampaignDescription() (+8 more)

### Community 153 - "Handler"
Cohesion: 0.17
Nodes (6): Handler, NewHandler(), twitch.PrivateMessage, twitch.UserJoinMessage, twitch.UserNoticeMessage, twitch.UserPartMessage

### Community 154 - "newTestRateLimiter"
Cohesion: 0.31
Nodes (9): newRateLimiter(), newTestRateLimiter(), TestNewRateLimiterRejectsInvalidConfiguration(), TestRateLimiterDoesNotExceedLimitAfterIdle(), TestRateLimiterEvictsIdleClients(), TestRateLimiterKeepsActiveClients(), TestRateLimiterRefillsUnderSteadyTraffic(), trackedClients() (+1 more)

### Community 155 - "Manager"
Cohesion: 0.24
Nodes (3): Manager, twitch.Client, NewManager()

### Community 156 - "Store"
Cohesion: 0.29
Nodes (3): Store, file, Record

### Community 157 - "ParseKey"
Cohesion: 0.31
Nodes (8): benchKey(), BenchmarkDecrypt(), BenchmarkEncrypt(), BenchmarkParseKey(), ParseKey(), TestParseKey(), TestParseKeyInvalidBase64(), TestParseKeyWrongLength()

### Community 158 - "newTestConnection"
Cohesion: 0.43
Nodes (7): Connection, newTestConnection(), TestHandleResponse_ERR_BADAUTH_AlreadyRefreshedByAnother(), TestHandleResponse_ERR_BADAUTH_RefreshesAndResubscribes(), TestHandleResponse_ERR_BADAUTH_RefreshFailsNoResubscribe(), TestHandleResponse_OtherErrors_NoRefresh(), TestHandleResponse_ReconnectClosesConnection()

### Community 159 - "campaignCompleted"
Cohesion: 0.33
Nodes (6): campaignCompleted(), completedCampaigns(), TestCampaignCompletedUsesIDOrEndTimeTolerance(), TestCompletedCampaigns(), TestCompletedCampaignsRejectsMalformedInventory(), completedCampaignSignature

## Knowledge Gaps
- **356 isolated node(s):** `_run-localdev.sh script`, `TWITCH_MINER_RUN_LOCALDEV`, `_run.sh script`, `github.com/Guliveer/twitch-miner-go`, `TokenErrorResponse` (+351 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 617 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **24 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Streamer` connect `Streamer` to `AnalyticsServer`, `Raid`, `BadgeWatcher`, `time.Time`, `Miner`, `sync.Mutex`, `Client`, `context.Context`, `Message`, `net/http.ResponseWriter`, `CommunityGoal`, `newMockTransport`, `prediction.go`, `makeOnlineStreamer`, `Connection`, `time.Duration`, `clientWithStreaks`, `NewStreamer`, `Miner`, `Miner`, `newTestMiner`?**
  _High betweenness centrality (0.081) - this node is a cross-community bridge._
- **Why does `Logger` connect `Logger` to `AnalyticsServer`, `BadgeWatcher`, `Client`, `Dispatcher`, `NewAnalyticsServer`, `Client`, `Batcher`, `net/http.ResponseWriter`, `Manager`, `newMockTransport`, `Connection`, `Handler`, `time.Duration`, `main`, `Manager`, `autostart_unix.go`, `poller_test.go`, `middleware.go`, `NewStreamer`, `Authenticator`, `Miner`, `Poller`, `Event`, `NewBatcher`, `Validate`, `checkWithURL`?**
  _High betweenness centrality (0.062) - this node is a cross-community bridge._
- **Why does `onReady()` connect `autostart_unix.go` to `SafeGo`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **What connects `_run-localdev.sh script`, `TWITCH_MINER_RUN_LOCALDEV`, `_run.sh script` to the rest of the system?**
  _356 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `web/app.js` be split into smaller, more focused modules?**
  _Cohesion score 0.08143074581430745 - nodes in this community are weakly interconnected._
- **Should `Deployment Guide` be split into smaller, more focused modules?**
  _Cohesion score 0.125 - nodes in this community are weakly interconnected._
- **Should `ParseMessage` be split into smaller, more focused modules?**
  _Cohesion score 0.14855072463768115 - nodes in this community are weakly interconnected._