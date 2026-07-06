// Command miner is the entry point for the Twitch Channel Points Miner.
// It loads account configurations, starts one Miner per account, and
// manages graceful shutdown via OS signals.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/managedminer"
	"github.com/Guliveer/twitch-miner-go/internal/miner"
	"github.com/Guliveer/twitch-miner-go/internal/model"
	"github.com/Guliveer/twitch-miner-go/internal/runtimecfg"
	"github.com/Guliveer/twitch-miner-go/internal/server"
	"github.com/Guliveer/twitch-miner-go/internal/store"
	"github.com/Guliveer/twitch-miner-go/internal/telemetry"
	"github.com/Guliveer/twitch-miner-go/internal/updater"
	"github.com/Guliveer/twitch-miner-go/internal/utils"
	"github.com/Guliveer/twitch-miner-go/internal/version"
	"github.com/joho/godotenv"
)

var bannerPlain = []string{
	"  ______       _ __       __      __  ____               ",
	" /_  __/    __(_) /______/ /     /  |/  (_)___  ___  ____",
	"  / / | |/|/ / / __/ __/ __ \\   / /|_/ / / __ \\/ _ \\/ __/",
	" / /  |__,__/ / /_/ /_/ / / /  / /  / / / / / /  __/ /   ",
	"/_/        /_/\\__/\\__/_/ /_/  /_/  /_/_/_/ /_/\\___/_/    ",
}

var bannerColors = []string{
	"\033[38;5;129m", "\033[38;5;128m", "\033[38;5;128m",
	"\033[38;5;127m", "\033[38;5;127m",
}

var subtitle = "⛏  twitch-miner-go " + version.String()

func playStartupAnimation(colored bool) {
	if colored {
		spinFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		for i := range 10 {
			fmt.Fprintf(os.Stderr, "\r\033[38;5;129m%s Initializing...\033[0m", spinFrames[i%len(spinFrames)])
			time.Sleep(80 * time.Millisecond)
		}
		fmt.Fprint(os.Stderr, "\r\033[K")
	}

	fmt.Println()
	for i, line := range bannerPlain {
		if colored {
			fmt.Printf("%s%s\033[0m\n", bannerColors[i], line)
		} else {
			fmt.Println(line)
		}
		time.Sleep(60 * time.Millisecond)
	}
	fmt.Println()

	for i, r := range subtitle {
		fmt.Fprintf(os.Stderr, "%c", r)
		if i < 3 {
			time.Sleep(100 * time.Millisecond)
		} else {
			time.Sleep(25 * time.Millisecond)
		}
	}
	fmt.Fprintln(os.Stderr)

	sep := strings.Repeat("─", 56)
	if colored {
		fmt.Printf("\033[38;5;240m%s\033[0m\n\n", sep)
	} else {
		fmt.Printf("%s\n\n", sep)
	}
}

func main() {
	configDir := flag.String("config", "configs", "Path to the configuration directory")
	port := flag.String("port", "8080", "Port for the health/analytics HTTP server")
	logLevel := flag.String("log-level", "", "Log level: DEBUG, INFO, WARN, ERROR (overrides LOG_LEVEL env)")
	logFormat := flag.String("log-format", "", "Log format: text or json (overrides LOG_FORMAT env)")
	logDirFlag := flag.String("log-dir", "", "Enable file logging; directory for .log files named with startup timestamp (overrides LOG_DIR env)")
	healthcheckURL := flag.String("healthcheck-url", "", "Probe the given HTTP URL and exit with status 0 only on HTTP 200")
	showVersion := flag.Bool("version", false, "Print version and exit")
	autoUpdate := flag.Bool("auto-update", false, "Automatically download and apply updates on startup")
	noLifecycleNotify := flag.Bool("no-lifecycle-notify", false, "Suppress MINER_STARTED / MINER_STOPPED / MINER_CRASHED notifications for this run")
	logNoTime := flag.Bool("log-no-time", false, "Omit timestamps in console logs (useful when the platform adds its own, e.g. Fly.io); overrides LOG_NO_TIME env")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.String())
		os.Exit(0)
	}

	if *healthcheckURL != "" {
		if err := runHealthcheck(*healthcheckURL); err != nil {
			fmt.Fprintf(os.Stderr, "Healthcheck failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if err := godotenv.Load(); err != nil {
		if _, statErr := os.Stat(".env"); statErr == nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse .env file: %v\n", err)
		}
	}

	level := resolveLogLevel(*logLevel)
	httpPort := resolvePort(*port)
	colored := logger.ColorSupported()
	format := resolveLogFormat(*logFormat)
	logDir := resolveLogDir(*logDirFlag)
	noTime := resolveLogNoTime(*logNoTime)

	rootLog, err := logger.Setup(logger.Config{Level: level, Colored: colored, NoTime: noTime, Format: format, LogDir: logDir})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup logger: %v\n", err)
		os.Exit(1)
	}

	playStartupAnimation(colored)
	rootLog.Info("Starting twitch-miner-go", "version", version.String())

	twitchRuntime := runtimecfg.LoadTwitchFromEnv(rootLog.Logger)

	utils.SafeGo(func() { runAutoUpdate(rootLog, *autoUpdate) })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	utils.SafeGo(func() {
		sig := <-sigCh
		rootLog.Info("Received shutdown signal", "signal", sig.String())
		cancel()
		time.AfterFunc(30*time.Second, func() {
			rootLog.Error("Graceful shutdown timed out, forcing exit")
			os.Exit(1)
		})
	})

	mgr := managedminer.NewManager(ctx, rootLog, twitchRuntime)
	mgr.SetSuppressLifecycleNotify(*noLifecycleNotify)

	dbEnabled := os.Getenv("DB_ENABLED") == "true"
	var accountStore store.Store

	var initialSync <-chan struct{}

	if dbEnabled {
		dsn := os.Getenv("DB_DSN")
		if dsn == "" {
			rootLog.Error("DB_ENABLED=true but DB_DSN is not set")
			os.Exit(1)
		}
		pg, err := store.OpenPostgres(dsn)
		if err != nil {
			rootLog.Error("Failed to connect to database", "error", err)
			os.Exit(1)
		}
		defer pg.Close()
		accountStore = pg

		pollInterval := resolvePollInterval()
		poller := managedminer.NewPoller(pg, mgr, pollInterval, rootLog)
		initialSync = poller.InitialSyncDone
		utils.SafeGo(func() { poller.Run(ctx) })
		rootLog.Info("🗄️ DB mode active — account configs loaded from database", "poll_interval", pollInterval)
	} else {
		fileInterval := resolveFileWatchInterval()
		fw := managedminer.NewFileWatcher(*configDir, mgr, fileInterval, rootLog)
		initialSync = fw.InitialSyncDone
		utils.SafeGo(func() { fw.Run(ctx) })
		rootLog.Info("📁 File watcher active — hot-reload enabled", "dir", *configDir, "poll_interval", fileInterval)
	}

	analyticsServer := setupAnalyticsServer(":"+httpPort, rootLog, mgr, accountStore)
	utils.SafeGo(func() {
		if err := analyticsServer.Run(ctx); err != nil && ctx.Err() == nil {
			rootLog.Error("Analytics server failed", "error", err)
		}
	})
	rootLog.Info("🌐 Health/analytics server started", "addr", ":"+httpPort)

	telemetryCfg, err := telemetry.LoadConfigFromEnv(rootLog.Logger)
	if err != nil {
		rootLog.Warn("📡 Telemetry: failed to load config", "error", err)
	} else if telemetryCfg != nil {
		telemetryCfg.Version = version.Number
		sender := telemetry.NewSender(telemetryCfg, rootLog.Logger)
		sender.SetRunningAccountsFunc(func() int {
			return len(mgr.Entries())
		})
		if accountStore != nil {
			sender.SetTotalConfigsFunc(func() int {
				accounts, err := accountStore.ListAccounts()
				if err != nil {
					return 0
				}
				return len(accounts)
			})
		}
		utils.SafeGo(func() { sender.Run(ctx, initialSync) })
		rootLog.Info("📡 Anonymous telemetry enabled — sending instance_id, version, os, arch, running/total account count. To disable, set TELEMETRY_AGREE=false")
	} else {
		rootLog.Info("📡 Telemetry disabled",
			"help", "Set TELEMETRY_AGREE=false explicitly, or omit it to enable default telemetry",
		)
	}

	<-ctx.Done()
	mgr.StopAll()
	rootLog.Info("🛑 Shutdown complete")
	rootLog.Info("👋 All miners stopped. Goodbye!")
}

func resolveLogLevel(flag string) slog.Level {
	if flag != "" {
		return logger.ParseLevel(flag)
	}
	if env := os.Getenv("LOG_LEVEL"); env != "" {
		return logger.ParseLevel(env)
	}
	return slog.LevelInfo
}

func resolveLogFormat(flag string) string {
	if flag != "" {
		return strings.ToLower(flag)
	}
	if env := os.Getenv("LOG_FORMAT"); env != "" {
		return strings.ToLower(env)
	}
	return "text"
}

func resolveLogNoTime(flagVal bool) bool {
	if flagVal {
		return true
	}
	return os.Getenv("LOG_NO_TIME") == "true"
}

func resolveLogDir(flagVal string) string {
	if flagVal != "" {
		return flagVal
	}
	if env := os.Getenv("LOG_DIR"); env != "" {
		return env
	}
	return ""
}

func resolvePort(flag string) string {
	if env := os.Getenv("PORT"); env != "" {
		return env
	}
	return flag
}

func setupAnalyticsServer(addr string, rootLog *logger.Logger, mgr *managedminer.Manager, accountStore store.Store) *server.AnalyticsServer {
	var dashboardAuth *server.DashboardAuth
	if user := os.Getenv("DASHBOARD_USER"); user != "" {
		dashboardAuth = &server.DashboardAuth{
			Username:     user,
			PasswordHash: os.Getenv("DASHBOARD_PASSWORD_SHA256"),
		}
	}
	dashboardAPIKey := os.Getenv("DASHBOARD_API_KEY")
	srv := server.NewAnalyticsServer(addr, rootLog, dashboardAuth, dashboardAPIKey)

	srv.SetStreamerFunc(func() []*model.Streamer {
		var all []*model.Streamer
		for _, e := range mgr.Entries() {
			all = append(all, e.Miner.Streamers()...)
		}
		return all
	})

	srv.SetNotifyTestFunc(func(ctx context.Context) []error {
		return testNotifiers(ctx, mgr)
	})

	srv.SetDebugFunc(func() any {
		entries := mgr.Entries()
		snapshots := make([]miner.DebugSnapshot, 0, len(entries))
		for _, e := range entries {
			snapshots = append(snapshots, e.Miner.DebugSnapshot())
		}
		return map[string]any{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"miners":    snapshots,
		}
	})

	srv.SetAccountStore(accountStore)

	srv.SetAuthStatusFunc(func(username string) any {
		for _, e := range mgr.Entries() {
			if strings.EqualFold(e.Miner.Username(), username) {
				status := e.Miner.DeviceCodeStatus()
				if status == nil {
					return nil
				}
				return status
			}
		}
		return nil
	})

	return srv
}

func testNotifiers(ctx context.Context, mgr *managedminer.Manager) []error {
	entries := mgr.Entries()
	var allErrs []error
	for _, e := range entries {
		d := e.Miner.NotifyDispatcher()
		if d == nil || !d.HasNotifiers() {
			continue
		}
		errs := d.TestAll(ctx, "Twitch Miner", "🔔 Test notification — if you see this, notifications are working!")
		allErrs = append(allErrs, errs...)
	}
	if len(entries) > 0 && allErrs == nil {
		for _, e := range entries {
			d := e.Miner.NotifyDispatcher()
			if d != nil && d.HasNotifiers() {
				return nil
			}
		}
		return []error{fmt.Errorf("no notification providers configured in any miner")}
	}
	return allErrs
}

func resolvePollInterval() time.Duration {
	if s := os.Getenv("DB_POLL_INTERVAL"); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
	}
	return 30 * time.Second
}

func resolveFileWatchInterval() time.Duration {
	if s := os.Getenv("FILE_POLL_INTERVAL"); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
	}
	return 5 * time.Second
}

func runAutoUpdate(rootLog *logger.Logger, autoUpdate bool) {
	info, err := updater.CheckForUpdate(context.Background(), version.Number)
	if err != nil {
		rootLog.Debug("Update check failed", "error", err)
		return
	}
	if !info.Available {
		return
	}
	if autoUpdate && info.AssetURL != "" {
		rootLog.Info("New version available, downloading update", "version", info.Latest)
		tmp, err := updater.DownloadAsset(context.Background(), info.AssetURL)
		if err != nil {
			rootLog.Warn("Auto-update download failed, continuing with current version", "error", err)
			fmt.Print(updater.FormatNotification(info, version.Number))
			return
		}
		if err := updater.ReplaceBinary(tmp); err != nil {
			rootLog.Warn("Auto-update replace failed, continuing with current version", "error", err)
			fmt.Print(updater.FormatNotification(info, version.Number))
			return
		}
		updater.ExitForRestart(rootLog.Logger)
	} else {
		fmt.Print(updater.FormatNotification(info, version.Number))
	}
}

func runHealthcheck(target string) error {
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}
