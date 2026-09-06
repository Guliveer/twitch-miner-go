//go:build !ios

// Package tray provides a system tray icon for the miner binary so users
// can open the dashboard and config editor and trigger a graceful exit
// without needing the terminal/console.
//
// fyne.io/systray requires cgo on macOS (Objective-C/Cocoa), so darwin builds
// need CGO_ENABLED=1. The Linux release pipeline uses CGO_ENABLED=0 and the
// library is pure Go there (DBus StatusNotifier). On Windows the library is
// pure Go as well.
//
// On Windows the tray is only meaningful in an interactive session; when the
// process runs as a service (session 0) there is no desktop and the tray is
// skipped. See session_windows.go.
package tray

import (
	"embed"
	"runtime"

	"fyne.io/systray"

	"github.com/Guliveer/twitch-miner-go/internal/utils"
)

//go:embed icon.png icon.ico
var iconFS embed.FS

// Links holds the local URLs the tray menu opens.
type Links struct {
	// DashboardURL is the analytics dashboard URL (e.g. http://localhost:8080).
	DashboardURL string
	// ConfigEditorURL is the config editor URL (e.g. http://localhost:8070).
	ConfigEditorURL string
}

// Options control tray behaviour.
type Options struct {
	// Links to open from the tray menu.
	Links Links
	// Title and tooltip shown for the tray icon.
	Title string
	// ExePath is the absolute path to the running binary, used to build the
	// logon autostart entry. When empty, autostart is disabled.
	ExePath string
	// OnExit is invoked on a graceful exit request (Exit menu item).
	// It runs on its own goroutine.
	OnExit func()
	// OnLog receives diagnostic lines from the tray menu (service/autostart
	// actions) so the host can surface them. Optional.
	OnLog func(format string, args ...any)
}

// Available reports whether a tray icon can be shown in the current session.
// It returns false when running as a Windows service (session 0), where no
// desktop is available to host the icon.
func Available() bool {
	return sessionAvailable()
}

// Run starts the tray icon event loop on a dedicated OS thread and returns
// immediately. It is safe to call even when Available() is false; in that case
// it does nothing. The provided OnExit callback is invoked when the user
// selects the Exit menu item.
func Run(opts Options) {
	if !sessionAvailable() {
		return
	}

	// systray.Run blocks on its platform event loop and must run on a
	// dedicated OS thread (macOS requirement; harmless elsewhere). We start it
	// on a locked goroutine so the miner's own main flow is unaffected.
	utils.SafeGo(func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		systray.Run(func() { onReady(opts) }, func() {})
	})
}

// Quit stops the tray icon and its event loop.
func Quit() {
	if sessionAvailable() {
		systray.Quit()
	}
}

func onReady(opts Options) {
	// fyne.io/systray requires .ico bytes on Windows and accepts .png elsewhere.
	iconName := "icon.png"
	if runtime.GOOS == "windows" {
		iconName = "icon.ico"
	}
	icon, err := iconFS.ReadFile(iconName)
	if err != nil {
		// Non-fatal: the tray still works without a custom icon.
		icon = nil
	}
	if icon != nil {
		systray.SetIcon(icon)
	}
	if opts.Title != "" {
		systray.SetTitle(opts.Title)
		systray.SetTooltip(opts.Title)
	} else {
		systray.SetTooltip("twitch-miner-go")
	}

	// Left-click on the icon opens the dashboard.
	systray.SetOnTapped(func() {
		if opts.Links.DashboardURL != "" {
			_ = utils.OpenBrowser(opts.Links.DashboardURL)
		}
	})

	mDashboard := systray.AddMenuItem("Dashboard", "Open the analytics dashboard")

	// An empty ConfigEditorURL means the config editor is not running (DB mode),
	// so the menu must not offer an item that points at a dead editor.
	var mEditor *systray.MenuItem
	if opts.Links.ConfigEditorURL != "" {
		mEditor = systray.AddMenuItem("Config Editor", "Open the config editor")
	}

	// Service actions delegate to the platform installer script; the submenu is
	// only built when the script is discoverable next to the binary.
	var (
		mServiceInstall   *systray.MenuItem
		mServiceStart     *systray.MenuItem
		mServiceStop      *systray.MenuItem
		mServiceRestart   *systray.MenuItem
		mServiceStatus    *systray.MenuItem
		mServiceUninstall *systray.MenuItem
	)
	if ServiceScriptAvailable() {
		mService := systray.AddMenuItem("Service", "Manage the background service")
		mServiceInstall = mService.AddSubMenuItem("Install", "Install and configure the service")
		mServiceStart = mService.AddSubMenuItem("Start", "Start the service")
		mServiceStop = mService.AddSubMenuItem("Stop", "Stop the service")
		mServiceRestart = mService.AddSubMenuItem("Restart", "Restart the service")
		mServiceStatus = mService.AddSubMenuItem("Status", "Show service status")
		mService.AddSeparator()
		mServiceUninstall = mService.AddSubMenuItem("Uninstall", "Remove the service")
	}

	// Startup controls logon autostart (checkbox) and boot autostart (delegates
	// to the service installer, which registers a system service).
	mStartup := systray.AddMenuItem("Startup", "Start twitch-miner-go automatically")
	mAutostart := mStartup.AddSubMenuItemCheckbox("Start on logon", "Launch when you log on", AutostartEnabled())
	mBoot := mStartup.AddSubMenuItem("Start at boot", "Install as a service that starts at boot")

	// The console toggle switches between hiding and showing the terminal
	// window on Windows. The initial title reflects the process state: when
	// started with -no-console the window is already hidden, so the item must
	// offer to bring it back.
	var mHideConsole *systray.MenuItem
	if runtime.GOOS == "windows" {
		if isConsoleHidden() {
			mHideConsole = systray.AddMenuItem("Show Terminal", "Show the terminal window")
		} else {
			mHideConsole = systray.AddMenuItem("Hide Terminal", "Hide the terminal window")
		}
	}

	systray.AddSeparator()
	mExit := systray.AddMenuItem("Exit", "Stop twitch-miner-go")

	logf := opts.OnLog
	if logf == nil {
		logf = func(string, ...any) {}
	}

	go func() {
		for {
			select {
			case <-mDashboard.ClickedCh:
				_ = utils.OpenBrowser(opts.Links.DashboardURL)
			case <-clickCh(mEditor):
				_ = utils.OpenBrowser(opts.Links.ConfigEditorURL)
			case <-clickCh(mServiceInstall):
				RunServiceAction(ServiceInstall, logf)
			case <-clickCh(mServiceStart):
				RunServiceAction(ServiceStart, logf)
			case <-clickCh(mServiceStop):
				RunServiceAction(ServiceStop, logf)
			case <-clickCh(mServiceRestart):
				RunServiceAction(ServiceRestart, logf)
			case <-clickCh(mServiceStatus):
				RunServiceAction(ServiceStatus, logf)
			case <-clickCh(mServiceUninstall):
				RunServiceAction(ServiceUninstall, logf)
			case <-mAutostart.ClickedCh:
				if opts.ExePath == "" {
					logf("autostart: no executable path known, cannot toggle")
					continue
				}
				enable := !mAutostart.Checked()
				if err := SyncAutostart(enable, opts.ExePath); err != nil {
					logf("autostart: %v", err)
					continue
				}
				if enable {
					mAutostart.Check()
					logf("autostart: enabled (starts on logon)")
				} else {
					mAutostart.Uncheck()
					logf("autostart: disabled")
				}
			case <-mBoot.ClickedCh:
				RunServiceAction(ServiceInstall, logf)
			case <-clickCh(mHideConsole):
				if isConsoleHidden() {
					showConsole()
					mHideConsole.SetTitle("Hide Terminal")
					mHideConsole.SetTooltip("Hide the terminal window")
				} else {
					HideConsole()
					mHideConsole.SetTitle("Show Terminal")
					mHideConsole.SetTooltip("Show the terminal window")
				}
			case <-mExit.ClickedCh:
				systray.Quit()
				if opts.OnExit != nil {
					opts.OnExit()
				}
				return
			}
		}
	}()
}

// clickCh returns the item's click channel, or a nil channel (which never
// fires) when item is nil, so actions for unavailable menu items stay inert.
func clickCh(item *systray.MenuItem) <-chan struct{} {
	if item == nil {
		return nil
	}
	return item.ClickedCh
}
