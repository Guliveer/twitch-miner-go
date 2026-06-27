package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/Guliveer/twitch-miner-go/internal/configeditor"
)

func main() {
	var (
		configDir string
		port      int
		tui       bool
		noBrowser bool
	)

	flag.StringVar(&configDir, "config", "configs", "Path to the configs directory")
	flag.IntVar(&port, "port", 3000, "HTTP port for web mode")
	flag.BoolVar(&tui, "tui", false, "Launch TUI mode instead of the web server")
	flag.BoolVar(&noBrowser, "no-browser", false, "Web mode: don't auto-open the browser")
	flag.Parse()

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot create config directory %q: %v\n", configDir, err)
		os.Exit(1)
	}

	if tui {
		if err := configeditor.RunTUI(configDir); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	srv := configeditor.NewServer(configDir)
	addr := fmt.Sprintf(":%d", port)
	url := fmt.Sprintf("http://localhost:%d", port)

	fmt.Printf("\n  Config Editor running at %s\n", url)
	fmt.Printf("  Config directory: %s\n\n", configDir)

	if !noBrowser {
		go openBrowser(url)
	}

	if err := http.ListenAndServe(addr, srv); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	_ = exec.Command(cmd, args...).Start()
}
