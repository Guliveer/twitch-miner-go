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
	"sync"
	"syscall"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/miner"
	"github.com/Guliveer/twitch-miner-go/internal/model"
	"github.com/Guliveer/twitch-miner-go/internal/runtimecfg"
	"github.com/Guliveer/twitch-miner-go/internal/server"
	"github.com/Guliveer/twitch-miner-go/internal/updater"
	"github.com/Guliveer/twitch-miner-go/internal/utils"
	"github.com/Guliveer/twitch-miner-go/internal/version"
	"github.com/joho/godotenv"
)

var bannerPlain = []string{
	"  ______       _ __       __       __  ____               ",
	" /_  __/    __(_) /______/ /_     /  |/  (_)___  ___  ____",
	"  / / | |/|/ / / __/ __/ __ \\   / /|_/ / / __ \\/ _ \\/ __/",
	" / /  |__,__/ / /_/ /_/ / / /  / /  / / / / / /  __/ /   ",
	"/_/        /_/\\__/\\__/_/ /_/  /_/  /_/_/_/ /_/\\___/_/    ",
}

var bannerColors = []string{
	"\033[38;5;129m", "\033[38;5;128m", "\033[38;5;127m",
	"\033[38;5;126m", "\033[38;5;125m",
}

var subtitle = "⛏  twitch-miner-go " + version.String()

func playStartupAnimation(colored bool) {
	if colored {
		spinFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		for i := 0; i < 10; i++ {
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
	healthcheckURL := flag.String("healthcheck-url", "", "Probe the given HTTP URL and exit with status 0 only on HTTP 200")
	showVersion := flag.Bool("version", false, "Print version and exit")
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

	// Load .env file if it exists (ignore error if file is missing)
	if err := godotenv.Load(); err != nil {
		// Only log if the file exists but couldn't be parsed
		if _, statErr := os.Stat(".env"); statErr == nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse .env file: %v\n", err)
		}
	}

	level := slog.LevelInfo
	if *logLevel != "" {
		level = logger.ParseLevel(*logLevel)
	} else if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		level = logger.ParseLevel(envLevel)
	}

	httpPort := *port
	if envPort := os.Getenv("PORT"); envPort != "" {
		httpPort = envPort
	}

	colored := logger.ColorSupported()

	rootLog, err := logger.Setup(logger.Config{
		Level:   level,
		Colored: colored,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup logger: %v\n", err)
		os.Exit(1)
	}

	playStartupAnimation(colored)
	rootLog.Info("🚀 Starting Twitch Channel Points Miner (Go)", "version", version.String())

	twitchRuntime := runtimecfg.LoadTwitchFromEnv(rootLog.Logger)

	// Check for updates in the background.
	utils.SafeGo(func() {
		info, err := updater.CheckForUpdate(context.Background(), version.Number)
		if err != nil {
			rootLog.Debug("Update check failed", "error", err)
			return
		}
		if msg := updater.FormatNotification(info, version.Number); msg != "" {
			fmt.Print(msg)
		}
	})

	configs, err := config.LoadAllAccountConfigs(*configDir)
	if err != nil {
		rootLog.Error("Failed to load account configs", "dir", *configDir, "error", err)
		os.Exit(1)
	}

	for _, cfg := range configs {
		if !cfg.IsEnabled() {
			continue
		}
		if err := config.Validate(cfg); err != nil {
			rootLog.Error("Invalid config", "account", cfg.Username, "error", err)
			os.Exit(1)
		}
	}

	rootLog.Info("📂 Loaded account configurations",
		"count", len(configs),
		"config_dir", *configDir,
	)

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

	type minerEntry struct {
		cfg   *config.AccountConfig
		miner *miner.Miner
	}

	miners := make([]minerEntry, 0, len(configs))
	for _, cfg := range configs {
		if !cfg.IsEnabled() {
			rootLog.Info("Account is disabled, skipping", "account", cfg.Username)
			continue
		}
		accountLog := rootLog.WithAccount(cfg.Username)
		minerInstance := miner.NewMiner(cfg, accountLog, twitchRuntime)
		miners = append(miners, minerEntry{cfg: cfg, miner: minerInstance})
	}

	addr := ":" + httpPort
	var dashboardAuth *server.DashboardAuth
	if user := os.Getenv("DASHBOARD_USER"); user != "" {
		dashboardAuth = &server.DashboardAuth{
			Username:     user,
			PasswordHash: os.Getenv("DASHBOARD_PASSWORD_SHA256"),
		}
	}
	analyticsServer := server.NewAnalyticsServer(addr, rootLog, dashboardAuth)

	analyticsServer.SetStreamerFunc(func() []*model.Streamer {
		var all []*model.Streamer
		for _, entry := range miners {
			all = append(all, entry.miner.Streamers()...)
		}
		return all
	})

	analyticsServer.SetNotifyTestFunc(func(ctx context.Context) []error {
		var allErrs []error
		for _, entry := range miners {
			d := entry.miner.NotifyDispatcher()
			if d == nil || !d.HasNotifiers() {
				continue
			}
			errs := d.TestAll(ctx, "Twitch Miner", "🔔 Test notification — if you see this, notifications are working!")
			allErrs = append(allErrs, errs...)
		}
		if len(miners) > 0 && allErrs == nil {
			// Check if any miner had notifiers at all.
			hasAny := false
			for _, entry := range miners {
				d := entry.miner.NotifyDispatcher()
				if d != nil && d.HasNotifiers() {
					hasAny = true
					break
				}
			}
			if !hasAny {
				return []error{fmt.Errorf("no notification providers configured in any miner")}
			}
		}
		return allErrs
	})

	analyticsServer.SetDebugFunc(func() any {
		snapshots := make([]miner.DebugSnapshot, 0, len(miners))
		for _, entry := range miners {
			snapshots = append(snapshots, entry.miner.DebugSnapshot())
		}
		return map[string]any{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"miners":    snapshots,
		}
	})

	utils.SafeGo(func() {
		if err := analyticsServer.Run(ctx); err != nil && ctx.Err() == nil {
			rootLog.Error("Analytics server failed", "error", err)
		}
	})

	rootLog.Info("🌐 Health/analytics server started", "addr", addr)

	var wg sync.WaitGroup
	for _, entry := range miners {
		cfg := entry.cfg
		minerInstance := entry.miner
		accountLog := rootLog.WithAccount(cfg.Username)

		wg.Add(1)
		go func(minerInstance *miner.Miner) {
			defer wg.Done()
			if err := minerInstance.Run(ctx); err != nil {
				if ctx.Err() != nil {
					accountLog.Info("Miner stopped due to shutdown", "account", cfg.Username)
				} else {
					accountLog.Error("Miner failed", "account", cfg.Username, "error", err)
				}
			}
		}(minerInstance)
	}

	wg.Wait()

	if ctx.Err() != nil {
		rootLog.Info("🛑 Shutdown complete")
	}

	rootLog.Info("👋 All miners stopped. Goodbye!")
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
