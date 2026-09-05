// Package logger provides structured logging with colored console output,
// optional file output, and per-account logger prefixing using log/slog.
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/model"
)

var startupTime = time.Now()

var eventEmoji = map[string]string{
	"GAIN_FOR_WATCH":        "💵",
	"GAIN_FOR_WATCH_STREAK": "💵",
	"GAIN_FOR_CLAIM":        "💵",
	"GAIN_FOR_RAID":         "💵",
	"BONUS_CLAIM":           "💵",
	"BET_START":             "🎰",
	"BET_WIN":               "🏆",
	"BET_LOSE":              "💸",
	"BET_REFUND":            "↩️",
	"BET_FILTERS":           "🎰",
	"BET_GENERAL":           "🎰",
	"BET_FAILED":            "🎰",
	"DROP_CLAIM":              "📦",
	"DROP_CLAIM_AVAILABLE":    "📦",
	"DROP_STATUS":             "📦",
	"DROP_MILESTONE":          "📊",
	"STREAMER_ONLINE":       "🟢",
	"STREAMER_OFFLINE":      "⚫",
	"JOIN_RAID":             "⚔️",
	"CHAT_MENTION":          "💬",
	"MOMENT_CLAIM":            "🎉",
	"GIFTED_SUB":              "🎁",
	"ACCOUNT_CONFIG_RELOADED": "🔄",
}

// ANSI color codes for terminal output.
const (
	colorReset     = "\033[0m"
	colorRed       = "\033[31m"
	colorGreen     = "\033[32m"
	colorYellow    = "\033[33m"
	colorLightBlue = "\033[94m"
	colorMagenta   = "\033[35m"
	colorCyan      = "\033[36m"
	colorGray      = "\033[90m"
)

// coloredAttrKeys maps slog attribute keys to ANSI color codes for value highlighting.
var coloredAttrKeys = map[string]string{
	"streamer": colorMagenta,
	"channel":  colorMagenta,
	"category": colorLightBlue,
	"target":   colorMagenta,
}

// Field is a single structured key-value pair extracted from an event's log
// args. Fields keep the order they were logged in, which lets notification
// providers lay out a message deterministically.
type Field struct {
	Key   string
	Value string
}

// NotifyFunc is a callback invoked when a log event matches notification criteria.
// It receives the bare event message (emoji included, no key-value suffix) plus
// the ordered fields, so the notify package owns all message formatting.
// Implementations should be non-blocking.
type NotifyFunc func(ctx context.Context, event model.Event, message string, fields []Field)

// Config holds logger configuration options.
type Config struct {
	Level       slog.Level
	FileLevel   slog.Level
	Colored     bool
	NoTime      bool
	LogDir      string
	AccountName string
	NotifyFn    NotifyFunc
	Format      string
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Level:     slog.LevelInfo,
		FileLevel: slog.LevelDebug,
		Colored:   true,
	}
}

// Logger wraps slog.Logger with account-scoped context and notification dispatch.
type Logger struct {
	*slog.Logger
	cfg      Config
	notifyFn atomic.Value // stores NotifyFunc
	file     *os.File    // log file when LogDir is set, nil otherwise
}

// Setup creates a new Logger based on the provided configuration.
// It sets up console and optional file handlers.
func Setup(cfg Config) (*Logger, error) {
	var handlers []slog.Handler
	var (
		logFile *os.File
		err     error
	)

	var consoleHandler slog.Handler
	if cfg.Format == "json" {
		consoleHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Level})
		if cfg.AccountName != "" {
			consoleHandler = consoleHandler.WithAttrs([]slog.Attr{slog.String("account", cfg.AccountName)})
		}
	} else {
		consoleHandler = newColorHandler(os.Stdout, cfg.Level, cfg.Colored, cfg.NoTime, cfg.AccountName)
	}
	handlers = append(handlers, consoleHandler)

	if cfg.LogDir != "" {
		if err := os.MkdirAll(cfg.LogDir, 0o755); err != nil {
			return nil, fmt.Errorf("creating log directory %s: %w", cfg.LogDir, err)
		}

		timestamp := startupTime.Format("2003-12-31_13-10-00")
		filename := timestamp + ".log"
		if cfg.AccountName != "" {
			filename = timestamp + "_" + cfg.AccountName + ".log"
		}

		logFile, err = os.OpenFile(
			filepath.Join(cfg.LogDir, filename),
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0o644,
		)
		if err != nil {
			return nil, fmt.Errorf("opening log file: %w", err)
		}

		opts := &slog.HandlerOptions{Level: cfg.FileLevel}
		var fileHandler slog.Handler
		if cfg.Format == "json" {
			fileHandler = slog.NewJSONHandler(logFile, opts)
		} else {
			fileHandler = slog.NewTextHandler(logFile, opts)
		}
		if cfg.AccountName != "" {
			fileHandler = fileHandler.WithAttrs([]slog.Attr{slog.String("account", cfg.AccountName)})
		}
		handlers = append(handlers, fileHandler)
	}

	var handler slog.Handler
	if len(handlers) == 1 {
		handler = handlers[0]
	} else {
		handler = &multiHandler{handlers: handlers}
	}

	logger := &Logger{
		Logger: slog.New(handler),
		cfg:    cfg,
		file:   logFile,
	}

	if cfg.NotifyFn != nil {
		logger.notifyFn.Store(cfg.NotifyFn)
	}

	return logger, nil
}

// WithAccount returns a new Logger with the account name set.
func (l *Logger) WithAccount(name string) *Logger {
	newCfg := l.cfg
	newCfg.AccountName = name
	newLogger, _ := Setup(newCfg)
	return newLogger
}

// Close releases the underlying log file, if any. Safe to call multiple times.
func (l *Logger) Close() error {
	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}

// Event logs a message at INFO level and dispatches a notification if configured.
// If the event has a mapped emoji, it is prepended to the log message.
// The alternating key-value args are forwarded to the NotifyFunc as ordered
// Fields; the notify package decides how to render them.
func (l *Logger) Event(ctx context.Context, event model.Event, msg string, args ...any) {
	if emoji, ok := eventEmoji[string(event)]; ok {
		msg = emoji + " " + msg
	}
	l.Info(msg, append(args, "event", string(event))...)

	if fn, ok := l.notifyFn.Load().(NotifyFunc); ok && fn != nil {
		fields := make([]Field, 0, len(args)/2)
		for i := 0; i+1 < len(args); i += 2 {
			key, ok := args[i].(string)
			if !ok {
				continue
			}
			fields = append(fields, Field{Key: key, Value: fmt.Sprintf("%v", args[i+1])})
		}

		fn(ctx, event, msg, fields)
	}
}

// SetNotifyFunc sets the notification callback function. Thread-safe.
func (l *Logger) SetNotifyFunc(fn NotifyFunc) {
	l.notifyFn.Store(fn)
}

// ParseLevel converts a string log level to slog.Level.
func ParseLevel(s string) slog.Level {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type colorHandler struct {
	mu          sync.Mutex
	writer      io.Writer
	level       slog.Level
	colored     bool
	noTime      bool
	accountName string
	attrs       []slog.Attr
}

func newColorHandler(w io.Writer, level slog.Level, colored bool, noTime bool, accountName string) *colorHandler {
	return &colorHandler{
		writer:      w,
		level:       level,
		colored:     colored,
		noTime:      noTime,
		accountName: accountName,
	}
}

func (h *colorHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *colorHandler) Handle(_ context.Context, record slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	levelStr := record.Level.String()
	msg := record.Message

	prefix := ""
	if h.accountName != "" {
		prefix = fmt.Sprintf("[%s] ", h.accountName)
	}

	if h.colored {
		levelColor := h.levelColor(record.Level)
		if h.noTime {
			fmt.Fprintf(h.writer, "%s%s%s - %s%s",
				levelColor, levelStr, colorReset,
				prefix, msg,
			)
		} else {
			timeStr := record.Time.Format("02/01/06 15:04:05")
			fmt.Fprintf(h.writer, "%s%s - %s%s%s - %s%s",
				colorGray, timeStr,
				levelColor, levelStr, colorReset,
				prefix, msg,
			)
		}
	} else {
		if h.noTime {
			fmt.Fprintf(h.writer, "%s - %s%s", levelStr, prefix, msg)
		} else {
			timeStr := record.Time.Format("02/01/06 15:04:05")
			fmt.Fprintf(h.writer, "%s - %s - %s%s", timeStr, levelStr, prefix, msg)
		}
	}

	for _, a := range h.attrs {
		if h.colored {
			if color, ok := coloredAttrKeys[a.Key]; ok {
				fmt.Fprintf(h.writer, " %s=%s{%v}%s", a.Key, color, a.Value, colorReset)
				continue
			}
		}
		fmt.Fprintf(h.writer, " %s={%v}", a.Key, a.Value)
	}

	record.Attrs(func(a slog.Attr) bool {
		if h.colored {
			if color, ok := coloredAttrKeys[a.Key]; ok {
				fmt.Fprintf(h.writer, " %s=%s{%v}%s", a.Key, color, a.Value, colorReset)
				return true
			}
		}
		fmt.Fprintf(h.writer, " %s={%v}", a.Key, a.Value)
		return true
	})

	fmt.Fprintln(h.writer)
	return nil
}

func (h *colorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &colorHandler{
		writer:      h.writer,
		level:       h.level,
		colored:     h.colored,
		noTime:      h.noTime,
		accountName: h.accountName,
		attrs:       append(copyAttrs(h.attrs), attrs...),
	}
}

func (h *colorHandler) WithGroup(name string) slog.Handler {
	return &colorHandler{
		writer:      h.writer,
		level:       h.level,
		colored:     h.colored,
		noTime:      h.noTime,
		accountName: h.accountName,
		attrs:       copyAttrs(h.attrs),
	}
}

func copyAttrs(attrs []slog.Attr) []slog.Attr {
	if len(attrs) == 0 {
		return nil
	}
	cp := make([]slog.Attr, len(attrs))
	copy(cp, attrs)
	return cp
}

func (h *colorHandler) levelColor(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return colorRed
	case level >= slog.LevelWarn:
		return colorYellow
	case level >= slog.LevelInfo:
		return colorGreen
	default:
		return colorCyan
	}
}

type multiHandler struct {
	handlers []slog.Handler
}

func (handler *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range handler.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (handler *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range handler.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (handler *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(handler.handlers))
	for i, h := range handler.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

func (handler *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(handler.handlers))
	for i, h := range handler.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}
