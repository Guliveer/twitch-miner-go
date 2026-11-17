# Graph Report - twitch-miner-go  (2026-11-17)

## Corpus Check
- 218 files · ~209,287 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2467 nodes · 5816 edges · 154 communities (112 shown, 26 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 467 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `34599fc6`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Commit Types Cheatsheet
- web/app.js
- newTestConfigServer
- Deployment Guide
- ParseMessage
- DefaultBetSettings
- twitch-miner-go // *Go to bed. It's mining.*
- Server
- Client
- Dispatcher
- CookieJar
- newTestAccountsServer
- What You Must Do When Invoked
- context.Context
- Client
- Batcher
- install-service.sh
- Message
- net/http.ResponseWriter
- EventPrediction
- newMockTransport
- mockAuthProvider
- prediction.go
- Troubleshooting
- SelectWatchSet
- Connection
- main
- NewDrop
- Configuration Reference
- testing.T
- autostart_unix.go
- Authenticator
- BadgeWatcher
- middleware.go
- time.Duration
- Manager
- static/app.js
- logs.js
- makeOnlineStreamer
- Getting Started
- configeditor/server.go
- .pollForToken
- NewStreamer
- AccountRow
- .handle2FA
- badge_catalog.go
- Provider setup
- Prediction Strategies
- tokenChangingMock
- Miner
- loadBadgeCampaigns
- sync.Mutex
- manager_test.go
- streams_api.go
- Event
- NewBatcher
- newTestMiner
- Modern Dark
- graphify reference: extra exports and benchmark
- Data flow
- Advanced Guide
- time.Time
- PULL_REQUEST_TEMPLATE.md
- Run
- 1.10. Linux Service (systemd / OpenRC)
- Authentication
- FAQ
- server_test.go
- NewAnalyticsServer
- Logger
- graphify reference: query, path, explain
- autostart_windows.go
- Twitch Channel Points Miner — Go Edition
- AGENTS.md
- os/exec.Cmd
- tui.go
- _run.sh
- operations_test.go
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
- Store
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
- updater.go
- pre-commit
- mockAuthProvider
- pre-push
- AnalyticsServer
- CategoryWatcher
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
- Field
- NewStream
- Open
- Raid
- ownsCampaignBadge
- Twitch
- Miner
- newTestRateLimiter
- 1.6. Environment Variables
- capturingNotifier
- 1.7. Notifications
- Manager
- 1.12. Deploy to Fly.io
- 1.5. Configuration
- 1.11. Windows Service
- Code of Conduct
- StreamerSettings
- NewServer
- newTestConnection
- docs/app.js
- Security Policy
- .RoundTrip

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
- `main()` --calls--> `RunTUI()`  [EXTRACTED]
  cmd/config-editor/main.go → internal/configeditor/tui.go
- `main()` --calls--> `OpenPostgres()`  [EXTRACTED]
  cmd/db-seed/main.go → internal/store/postgres.go
- `main()` --calls--> `ParseKey()`  [EXTRACTED]
  cmd/twitch-miner-go/main.go → internal/encryption/encryption.go
- `main()` --calls--> `ColorSupported()`  [EXTRACTED]
  cmd/twitch-miner-go/main.go → internal/logger/color_support.go
- `main()` --calls--> `Setup()`  [EXTRACTED]
  cmd/twitch-miner-go/main.go → internal/logger/logger.go

## Import Cycles
- None detected.

## Communities (154 total, 26 thin omitted)

### Community 0 - "Commit Types Cheatsheet"
Cohesion: 0.05
Nodes (40): Automated Versioning, `build` – Build Process or Dependencies, `chore` – Maintenance and Routine Tasks, `ci` – Continuous Integration, Commit Convention, Commit Types Cheatsheet, Contributing, `docs` – Documentation (+32 more)

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
Cohesion: 0.14
Nodes (22): ParseMessage(), serverTime(), splitTopic(), TestMessage_String(), TestParseMessage_ChannelIDFallbackToTopicUser(), TestParseMessage_ChannelIDFromDataChannelID(), TestParseMessage_ClaimAvailable(), TestParseMessage_EmptyObject() (+14 more)

### Community 5 - "DefaultBetSettings"
Cohesion: 0.14
Nodes (42): DefaultBetSettings(), Outcome, NewBet(), NewEventPrediction(), TestNewBetComputesDerivedOutcomeValues(), makeBet(), TestCalculate_AllOutcomesZeroUsers(), TestCalculate_AmountCappedByMaxPoints() (+34 more)

### Community 6 - "twitch-miner-go // *Go to bed. It's mining.*"
Cohesion: 0.11
Nodes (19): 1.13. Development, 1.14.1. Automatic updates, 1.14. Auto-Update, 1.15. License, 1.16. FAQ, 1.17. Advanced Guide, 1.1. Table of Contents, 1.2. Features (+11 more)

### Community 7 - "Server"
Cohesion: 0.31
Nodes (6): Server, io.Reader, net/http.ServeMux, readJSON(), sendError(), sendJSON()

### Community 8 - "Client"
Cohesion: 0.10
Nodes (17): net/http.Transport, circuitBreaker, gqlError, gqlExtensions, gqlRequest, gqlResponse, operationBehavior, persistedQuery (+9 more)

### Community 9 - "Dispatcher"
Cohesion: 0.17
Nodes (7): sync/atomic.Bool, ParseEvent(), Dispatcher, NewDispatcher(), parseEvents(), Notifier, notifierEntry

### Community 10 - "CookieJar"
Cohesion: 0.07
Nodes (51): Cookie, CookieJar, testing.B, CookieFileExists(), NewCookieJar(), NewCookieJarWithEncryption(), TestAutoMigrationPlaintextToEncrypted(), TestEncryptedCookieWithoutKey() (+43 more)

### Community 11 - "newTestAccountsServer"
Cohesion: 0.28
Nodes (22): accountsRequest(), AnalyticsServer, minimalAccountCfg(), newFakeStore(), newTestAccountsServer(), TestCreateAccount_InvalidJSON(), TestCreateAccount_MissingUsername(), TestCreateAccount_NoStore() (+14 more)

### Community 12 - "What You Must Do When Invoked"
Cohesion: 0.07
Nodes (26): For /graphify add and --watch, For /graphify query, For the commit hook and native CLAUDE.md integration, For --update and --cluster-only, /graphify, Honesty Rules, Interpreter guard for subcommands, Part A - Structural extraction for code files (+18 more)

### Community 13 - "context.Context"
Cohesion: 0.07
Nodes (10): context.Context, encoding/json.RawMessage, GameResp, GoalContribution, PlaybackAccessToken, TeamMember, Client, StreamInfoResponse (+2 more)

### Community 14 - "Client"
Cohesion: 0.09
Nodes (10): sync.Map, sync.RWMutex, LookupGameSlug(), RegisterGameSlug(), Client, isStaleUsernameError(), TestIsStaleUsernameError(), spadeCache (+2 more)

### Community 15 - "Batcher"
Cohesion: 0.36
Nodes (3): batchKey, batchEntry, Batcher

### Community 16 - "install-service.sh"
Cohesion: 0.18
Nodes (28): ask(), banner(), confirm(), DEFAULT_CONFIG_DIR, DEFAULT_DATA_DIR, DEFAULT_ENV_FILE, DEFAULT_INSTALL_DIR, DEFAULT_LOG_LEVEL (+20 more)

### Community 18 - "net/http.ResponseWriter"
Cohesion: 0.19
Nodes (11): net/http.Request, net/http.ResponseWriter, AnalyticsServer, AnalyticsServer, AnalyticsServer, filterStreamers(), AnalyticsServer, parsePagination() (+3 more)

### Community 19 - "EventPrediction"
Cohesion: 0.07
Nodes (27): ChannelPointsContext, BoolFromMap(), FloatFromAny(), IntFromAny(), IntFromMap(), StringFromAny(), StringFromMap(), TestBoolFromMap() (+19 more)

### Community 20 - "newMockTransport"
Cohesion: 0.09
Nodes (66): bytes.Buffer, NewForTest(), NewClientForTest(), boolPtr(), Client, makeStreamerWithDrop(), makeStreamerWithDropAtProgress(), milestoneEventCount() (+58 more)

### Community 22 - "prediction.go"
Cohesion: 0.07
Nodes (35): BuildDefaults(), BuildSchemaPayload(), T, stringValues(), TestBuildSchemaPayload_ContainsAllChatModes(), TestBuildSchemaPayload_ContainsAllDelayModes(), TestBuildSchemaPayload_ContainsAllEvents(), TestBuildSchemaPayload_ContainsAllFollowersOrders() (+27 more)

### Community 23 - "Troubleshooting"
Cohesion: 0.07
Nodes (27): `401 Unauthorized` errors in logs, All bets on one outcome keep losing, `authenticated as "X" but config expects "Y"`, Authentication errors, Config changes not taking effect, Config editor & tray issues, Config issues, Drop issues (+19 more)

### Community 24 - "SelectWatchSet"
Cohesion: 0.13
Nodes (32): Miner, Connection, ConnectionSnapshot, Pool, collectOnlineIndices(), WatchSet, SelectWatchSet(), selectWatchSetAt() (+24 more)

### Community 25 - "Connection"
Cohesion: 0.07
Nodes (16): github.com/coder/websocket.Conn, Provider, PubSubTopic, NewStreamerTopic(), NewUserTopic(), Connection, NewConnection(), Connection (+8 more)

### Community 26 - "main"
Cohesion: 0.16
Nodes (19): main(), playStartupAnimation(), resolveFileWatchInterval(), resolveLogDir(), resolveLogFormat(), resolveLogNoTime(), resolveNoBanner(), resolveNoTray() (+11 more)

### Community 27 - "NewDrop"
Cohesion: 0.06
Nodes (31): Campaign, NewCampaign(), Drop, NewDrop(), Percentage(), refTime(), TestDropUpdate_FieldUpdate_DropInstanceID(), TestDropUpdate_FieldUpdate_HasPreconditionsMet() (+23 more)

### Community 28 - "Configuration Reference"
Cohesion: 0.10
Nodes (20): `badge_watcher`, `batch` (global batching defaults), `bet.filter_condition` (optional), `bet` (nested under `streamer_defaults` and per-streamer `settings`), `blacklist`, `category_blacklist`, `category_watcher`, Configuration Reference (+12 more)

### Community 29 - "testing.T"
Cohesion: 0.13
Nodes (34): testing.T, TestParseCondition(), TestParseStrategy(), TestStrategy_String(), TestClientIDsForGQL_Dedup(), TestLoadTwitchFromEnv_Defaults(), TestLoadTwitchFromEnv_EnvOverride(), TestNoopStore_ChangesIsNil() (+26 more)

### Community 30 - "autostart_unix.go"
Cohesion: 0.05
Nodes (48): fyne.io/systray.MenuItem, net/url.Values, BugReportURL(), Expand(), FeatureRequestURL(), issueURL(), queryOf(), setVersion() (+40 more)

### Community 31 - "Authenticator"
Cohesion: 0.10
Nodes (5): generateDeviceID(), GenerateHex(), Authenticator, NewAuthenticator(), DeviceCodeStatus

### Community 32 - "BadgeWatcher"
Cohesion: 0.16
Nodes (14): TopStream, badgeCampaignAllowsChannel(), badgeCampaignGameMatches(), badgeStreamerValid(), findStreamer(), BadgeWatcher, NewBadgeWatcher(), unmarkBadge() (+6 more)

### Community 33 - "middleware.go"
Cohesion: 0.18
Nodes (14): net/http.Handler, checkCredentials(), generateRequestID(), RequestIDFromContext(), withAuth(), withLogging(), withRateLimit(), withRequestID() (+6 more)

### Community 34 - "time.Duration"
Cohesion: 0.05
Nodes (45): BetSettingsConfig, CategoryConfig, DiscordConfig, FeaturesConfig, FilterConditionConfig, FollowersConfig, GotifyConfig, jsonDuration (+37 more)

### Community 35 - "Manager"
Cohesion: 0.10
Nodes (9): Handler, Manager, twitch.Client, NewManager(), NewHandler(), twitch.PrivateMessage, twitch.UserJoinMessage, twitch.UserNoticeMessage (+1 more)

### Community 36 - "static/app.js"
Cohesion: 0.30
Nodes (14): buildFilterParams(), clearFilters(), debounce(), escapeHTML(), fetchJSON(), formatPoints(), initFilterListeners(), loadFilters() (+6 more)

### Community 37 - "logs.js"
Cohesion: 0.27
Nodes (13): buildFilterParams(), clearFilters(), escapeHTML(), fetchJSON(), formatPoints(), handleSort(), loadFilters(), populateSelect() (+5 more)

### Community 38 - "makeOnlineStreamer"
Cohesion: 0.22
Nodes (23): SelectStreamersToWatch(), makeOnlineStreamer(), TestSelectStreamersToWatch_DropsOnly_IncludedWhenHasCampaignIDs(), TestSelectStreamersToWatch_DropsOnly_SkippedWhenNoCampaignIDs(), TestSelectStreamersToWatch_EndingSoonest(), TestSelectStreamersToWatch_EndingSoonest_PicksFirstEnd(), TestSelectStreamersToWatch_EndingSoonest_SkipsNoCampaigns(), TestSelectStreamersToWatch_Freeze_AllFrozenReturnsNone() (+15 more)

### Community 39 - "Getting Started"
Cohesion: 0.15
Nodes (13): 1. Clone and configure, 2. Set required environment variables, 3. Run, 4. Authenticate, 5. Verify it's working, Automatic updates, Getting Started, Going deeper (+5 more)

### Community 40 - "configeditor/server.go"
Cohesion: 0.21
Nodes (14): accountMeta, cleanConfig(), isValidDuration(), mergeSecretsBack(), removeEmpty(), shallowCopy(), stripSecrets(), TestCleanConfig() (+6 more)

### Community 41 - ".pollForToken"
Cohesion: 0.31
Nodes (4): DeviceCodeResponse, TokenErrorResponse, TokenResponse, Authenticator

### Community 42 - "NewStreamer"
Cohesion: 0.21
Nodes (16): DefaultStreamerSettings(), NewStreamer(), TestSetOnline_AlreadyOnline_NoOp(), TestSetOnline_Carryover_FirstOnline(), TestSetOnline_Carryover_StreakResolved_LongGap(), TestSetOnline_Carryover_StreakResolved_ShortGap(), TestSetOnline_Carryover_StreakUnresolved_ShortGap(), clientWithStreaks() (+8 more)

### Community 44 - "AccountRow"
Cohesion: 0.07
Nodes (9): database/sql.DB, sync.Once, OpenPostgres(), scanRow(), AccountRow, fakePollerStore, fakeStore, NoopStore (+1 more)

### Community 45 - ".handle2FA"
Cohesion: 0.42
Nodes (3): loginResponse, Authenticator, promptLine()

### Community 46 - "badge_catalog.go"
Cohesion: 0.17
Nodes (21): appendUnique(), appendUniqueBadgeNames(), badgeRequiresPayment(), badgeTitlesForWatchReward(), canonicalGameSlug(), comparableWords(), equalWords(), isBadgeReward() (+13 more)

### Community 47 - "Provider setup"
Cohesion: 0.17
Nodes (12): Discord, Event reference, Gotify, Matrix, Notification batching, Notifications, Per-account vs global credentials, Provider setup (+4 more)

### Community 48 - "Prediction Strategies"
Cohesion: 0.17
Nodes (12): Betting amount calculation, Delay modes, Filter conditions, `HIGH_ODDS`, `MOST_VOTED`, `NUMBER_1` through `NUMBER_8`, `PERCENTAGE`, Prediction Strategies (+4 more)

### Community 50 - "Miner"
Cohesion: 0.13
Nodes (6): time.Timer, Miner, NewMiner(), oneTimeEventMessage(), streakDir(), API

### Community 51 - "loadBadgeCampaigns"
Cohesion: 0.15
Nodes (19): campaignCompleted(), completedCampaigns(), loadBadgeCampaigns(), TestBadgeRewardMatching(), TestCampaignCompletedUsesIDOrEndTimeTolerance(), TestCompletedCampaigns(), TestCompletedCampaignsRejectsMalformedInventory(), TestLoadBadgeCampaignsAcceptsSingleChannel() (+11 more)

### Community 52 - "sync.Mutex"
Cohesion: 0.17
Nodes (12): resolveLogLevel(), io.Writer, log/slog.Attr, log/slog.Handler, log/slog.Level, log/slog.Record, sync.Mutex, copyAttrs() (+4 more)

### Community 53 - "manager_test.go"
Cohesion: 0.31
Nodes (14): newLiveManager(), newTestManager(), testConfig(), TestManager_EntriesReturnsSnapshot(), TestManager_LiveCountCountsStartingMiners(), TestManager_LiveCountExcludesExitedMiners(), TestManager_LiveCountZeroWithoutEntries(), TestManager_RestartReplacesEntry() (+6 more)

### Community 54 - "streams_api.go"
Cohesion: 0.18
Nodes (15): HistoryEntry, applyPagination(), T, campaignInfo, dropInfo, dropProgressShort, errorResponse, eventLogEntry (+7 more)

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

### Community 62 - "time.Time"
Cohesion: 0.27
Nodes (11): time.Time, RewardDeadline, applyUrgentReward(), availableDropDeadline(), campaignRewardStart(), preferUrgentReward(), streamerRewardStart(), urgentStart() (+3 more)

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
Cohesion: 0.33
Nodes (17): apiRequest(), newTestServer(), TestCreateAccount(), TestCreateAccountConflict(), TestCreateAccountInvalidName(), TestCreateAccountValidationFail(), TestDeleteAccount(), TestDeleteAccountNotFound() (+9 more)

### Community 69 - "NewAnalyticsServer"
Cohesion: 0.25
Nodes (15): net/http/httptest.ResponseRecorder, decodeObject(), TestAuthStatusesEmptyWhenFuncReturnsNil(), TestAuthStatusesEmptyWithoutFunc(), TestAuthStatusesReturnsPendingFlows(), TestAuthStatusesRouteIsRegistered(), AnalyticsServer, healthResponse() (+7 more)

### Community 70 - "Logger"
Cohesion: 0.21
Nodes (11): os.File, sync/atomic.Value, NewClient(), DefaultConfig(), Logger, NotifyFunc, Setup(), TestLogFileNameEncodesFullStartupTimestamp() (+3 more)

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
Cohesion: 0.22
Nodes (20): editAccountFields, applyCategoryWatcherSection(), applyEditFields(), applyFeaturesSection(), applyStreamersSection(), applyTeamWatcherSection(), boolVal(), intVal() (+12 more)

### Community 78 - "operations_test.go"
Cohesion: 0.31
Nodes (12): mockTransport, assertVarEquals(), assertVarPresent(), Client, newTestGQLClient(), TestGetChannelPointsContext_SendsChannelLogin(), TestGetFollowedStreamers_FailsWhenDataMissing(), TestGetFollowedStreamers_UsesDataAlongsidePartialErrors() (+4 more)

### Community 79 - "Validate"
Cohesion: 0.08
Nodes (42): AccountConfig, fatalf(), main(), net/url.URL, AccountConfigFromJSON(), AccountConfigToJSON(), applyDefaults(), applyEnvOverrides() (+34 more)

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

### Community 89 - "Store"
Cohesion: 0.29
Nodes (3): Store, file, Record

### Community 98 - "buildBody"
Cohesion: 0.19
Nodes (13): sync/atomic.Int64, buildBody(), TestBuildBodyBetGeneralFallsBackToMessage(), TestBuildBodyBetKeepsUnknownFields(), TestBuildBodyBetResultNamesChannel(), TestBuildBodyBetStartShowsCountdown(), TestBuildBodyNonBetKeepsCompactForm(), TestBuildBodyNonBetWithoutFields() (+5 more)

### Community 104 - "findProc"
Cohesion: 0.23
Nodes (9): golang.org/x/sys/windows.LazyProc, findProc(), HideConsole(), showConsole(), TestConsoleToggleStateStartsVisible(), TestGetConsoleWindowResolves(), TestShowConsoleWithoutHideIsNoop(), TestShowWindowResolves() (+1 more)

### Community 107 - "updater.go"
Cohesion: 0.07
Nodes (45): runAutoUpdate(), log/slog.Logger, detectDeployment(), LoadConfigFromEnv(), loadOrGenerateInstanceID(), NewSender(), newUUID(), CheckForUpdate() (+37 more)

### Community 111 - "AnalyticsServer"
Cohesion: 0.21
Nodes (8): net/http.Server, AnalyticsServer, AuthStatusFunc, AuthStatusListFunc, DebugSnapshotFunc, MinerCountFunc, NotifyTestFunc, StreamerFunc

### Community 112 - "CategoryWatcher"
Cohesion: 0.50
Nodes (3): CategoryWatcher, NewCategoryWatcher(), categoryEntry

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
Cohesion: 0.09
Nodes (20): Miner, Streamer, anyStreakPending(), applyPriority(), applyPriorityBadges(), applyPriorityDrops(), applyPriorityEndingSoonest(), applyPriorityLowAvailability() (+12 more)

### Community 130 - "Required Configuration"
Cohesion: 0.33
Nodes (6): Account Configuration, Authentication, Getting Browser Values, Getting TV Client ID, Required Configuration, Twitch Runtime Identifiers

### Community 131 - "Configuration (Fly.io)"
Cohesion: 0.33
Nodes (6): Configuration (Fly.io), Fly.io, Monitoring, Quick Start (Fly.io), Scaling, Secrets Management

### Community 132 - "Field"
Cohesion: 0.71
Nodes (6): Field, buildBetBody(), buildCompactBody(), fieldValue(), formatChannel(), formatPick()

### Community 133 - "NewStream"
Cohesion: 0.33
Nodes (10): NewStream(), TestHealthyChannelIsNotStalled(), TestMarkMinuteWatchAttemptStamps(), TestMinuteWatchedAccumulatesAcrossSegments(), TestMinuteWatchedCountsContinuousWatching(), TestMinuteWatchedIgnoresUnwatchedGaps(), TestNeverAttemptedIsNotStalled(), TestNeverCreditedIsNotStalled() (+2 more)

### Community 134 - "Open"
Cohesion: 0.33
Nodes (10): Open(), TestAccountNameIsCaseInsensitive(), TestCorruptFileYieldsUsableEmptyStore(), TestEarnedIsScopedToTheBroadcast(), TestExpiredRecordsAreDroppedOnReopen(), TestMissingFileOpensEmpty(), TestRecordLeavesNoTempFileBehind(), TestRecordSurvivesReopen() (+2 more)

### Community 140 - "Twitch"
Cohesion: 0.60
Nodes (3): envOrDefault(), Twitch, LoadTwitchFromEnv()

### Community 142 - "newTestRateLimiter"
Cohesion: 0.31
Nodes (9): newRateLimiter(), newTestRateLimiter(), TestNewRateLimiterRejectsInvalidConfiguration(), TestRateLimiterDoesNotExceedLimitAfterIdle(), TestRateLimiterEvictsIdleClients(), TestRateLimiterKeepsActiveClients(), TestRateLimiterRefillsUnderSteadyTraffic(), trackedClients() (+1 more)

### Community 143 - "1.6. Environment Variables"
Cohesion: 0.29
Nodes (7): 1.6.1. Global, 1.6.2. Per-Account Authentication, 1.6.3. Notification Secrets, 1.6.4.1. How To Obtain Twitch Runtime Identifiers, 1.6.4. `.env` File Support, 1.6.5. Cookie Encryption (optional), 1.6. Environment Variables

### Community 144 - "capturingNotifier"
Cohesion: 0.53
Nodes (3): newCapturingNotifier(), TestNotifyFuncDeliversChannelInBetBody(), capturingNotifier

### Community 145 - "1.7. Notifications"
Cohesion: 0.33
Nodes (6): 1.7.1. Supported Providers, 1.7.2. Example: Telegram, 1.7.3. Event Filtering, 1.7.4. Notification Batching, 1.7.5. Testing Notifications, 1.7. Notifications

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

### Community 151 - "Code of Conduct"
Cohesion: 0.22
Nodes (8): Attribution, Code of Conduct, Corrective Action Guide, Enforcement, Our Pledge, Our Standards, Reporting, Scope

### Community 152 - "StreamerSettings"
Cohesion: 0.24
Nodes (7): AllChatModes(), StreamerSettings, ParseChatPresence(), ShouldJoinChat(), TeamWatcher, NewTeamWatcher(), ChatPresence

### Community 155 - "NewServer"
Cohesion: 0.28
Nodes (4): main(), openBrowser(), net/http.HandlerFunc, NewServer()

### Community 157 - "newTestConnection"
Cohesion: 0.43
Nodes (7): Connection, newTestConnection(), TestHandleResponse_ERR_BADAUTH_AlreadyRefreshedByAnother(), TestHandleResponse_ERR_BADAUTH_RefreshesAndResubscribes(), TestHandleResponse_ERR_BADAUTH_RefreshFailsNoResubscribe(), TestHandleResponse_OtherErrors_NoRefresh(), TestHandleResponse_ReconnectClosesConnection()

### Community 162 - "docs/app.js"
Cohesion: 0.21
Nodes (13): animateNumber(), frame(), announce(), copyText(), finalNumber(), formatNumber(), initCopyButtons(), initDashboard() (+5 more)

### Community 163 - "Security Policy"
Cohesion: 0.40
Nodes (4): Reporting a Vulnerability, Scope, Security Policy, Supported Versions

## Knowledge Gaps
- **356 isolated node(s):** `_run-localdev.sh script`, `TWITCH_MINER_RUN_LOCALDEV`, `_run.sh script`, `github.com/Guliveer/twitch-miner-go`, `TokenErrorResponse` (+351 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 619 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **26 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Streamer` connect `Streamer` to `DefaultBetSettings`, `Raid`, `Miner`, `Client`, `context.Context`, `Message`, `net/http.ResponseWriter`, `EventPrediction`, `newMockTransport`, `StreamerSettings`, `SelectWatchSet`, `Connection`, `NewDrop`, `autostart_unix.go`, `BadgeWatcher`, `makeOnlineStreamer`, `NewStreamer`, `Miner`, `streams_api.go`, `newTestMiner`, `time.Time`, `AnalyticsServer`, `CategoryWatcher`?**
  _High betweenness centrality (0.103) - this node is a cross-community bridge._
- **Why does `Logger` connect `Logger` to `Client`, `Dispatcher`, `Client`, `Batcher`, `net/http.ResponseWriter`, `Manager`, `newMockTransport`, `StreamerSettings`, `Connection`, `main`, `Authenticator`, `BadgeWatcher`, `middleware.go`, `time.Duration`, `Manager`, `Miner`, `Event`, `NewBatcher`, `NewAnalyticsServer`, `Validate`, `updater.go`, `AnalyticsServer`, `CategoryWatcher`?**
  _High betweenness centrality (0.056) - this node is a cross-community bridge._
- **Why does `Miner` connect `Miner` to `Store`, `Streamer`, `time.Duration`, `Manager`, `BadgeWatcher`, `Logger`, `Dispatcher`, `Twitch`, `Client`, `CategoryWatcher`, `EventPrediction`, `sync.Mutex`, `Manager`, `prediction.go`, `Event`, `StreamerSettings`, `Connection`, `Authenticator`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._
- **What connects `_run-localdev.sh script`, `TWITCH_MINER_RUN_LOCALDEV`, `_run.sh script` to the rest of the system?**
  _356 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Commit Types Cheatsheet` be split into smaller, more focused modules?**
  _Cohesion score 0.05 - nodes in this community are weakly interconnected._
- **Should `web/app.js` be split into smaller, more focused modules?**
  _Cohesion score 0.08143074581430745 - nodes in this community are weakly interconnected._
- **Should `Deployment Guide` be split into smaller, more focused modules?**
  _Cohesion score 0.125 - nodes in this community are weakly interconnected._