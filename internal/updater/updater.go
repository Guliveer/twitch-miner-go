package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/version"
)

const (
	releaseURL      = "https://api.github.com/repos/Guliveer/twitch-miner-go/releases/latest"
	repoURL         = "https://github.com/Guliveer/twitch-miner-go"
	timeout         = 5 * time.Second
	downloadTimeout = 5 * time.Minute
)

// UpdateInfo holds the result of an update check.
type UpdateInfo struct {
	Available bool
	Latest    string
	URL       string
	IsGitRepo bool
	AssetURL  string // direct download URL for the current platform's binary
}

type ghRelease struct {
	TagName string    `json:"tag_name"`
	HTMLURL string    `json:"html_url"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckForUpdate checks GitHub for a newer release.
func CheckForUpdate(ctx context.Context, currentVersion string) (*UpdateInfo, error) {
	return checkWithURL(ctx, currentVersion, releaseURL)
}

func checkWithURL(ctx context.Context, currentVersion, url string) (*UpdateInfo, error) {
	if currentVersion == "dev" {
		return &UpdateInfo{Available: false}, nil
	}

	current, err := version.Parse(currentVersion)
	if err != nil {
		return &UpdateInfo{Available: false}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	latestStr := strings.TrimPrefix(release.TagName, "v")
	latest, err := version.Parse(latestStr)
	if err != nil {
		return nil, fmt.Errorf("parse remote version %q: %w", release.TagName, err)
	}

	info := &UpdateInfo{
		Available: version.Compare(latest, current) > 0,
		Latest:    latestStr,
		URL:       repoURL,
		IsGitRepo: isGitRepo(),
	}

	if info.Available {
		info.AssetURL = findAssetURL(release.Assets)
	}

	return info, nil
}

// findAssetURL returns the download URL for the current platform's binary asset.
func findAssetURL(assets []ghAsset) string {
	name, err := platformAsset()
	if err != nil {
		return ""
	}
	for _, a := range assets {
		if a.Name == name {
			return a.BrowserDownloadURL
		}
	}
	return ""
}

// platformAsset returns the expected release asset filename for the current OS and architecture.
func platformAsset() (string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	switch goarch {
	case "amd64", "arm64":
	default:
		return "", fmt.Errorf("unsupported architecture: %s", goarch)
	}

	name := fmt.Sprintf("twitch-miner-go-%s-%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}
	return name, nil
}

// DownloadAsset downloads the binary at url to a temporary file and returns its path.
// The caller is responsible for removing the file on error.
func DownloadAsset(ctx context.Context, url string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "twitch-miner-go-update-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()

	if err := os.Chmod(tmp.Name(), 0o755); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("chmod temp file: %w", err)
	}

	return tmp.Name(), nil
}

// osExecutable is a variable so tests can override it.
var osExecutable = os.Executable

// ReplaceBinary atomically replaces the current executable with the file at tmpPath.
// tmpPath is removed on failure.
func ReplaceBinary(tmpPath string) error {
	exe, err := osExecutable()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("resolve executable path: %w", err)
	}
	// Follow symlinks to the real file.
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("eval symlinks: %w", err)
	}

	if err := os.Rename(tmpPath, exe); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replace binary: %w", err)
	}
	return nil
}

// ExitForRestart logs that the update has been applied and exits with code 0
// so the service manager (systemd, NSSM, OpenRC) restarts the process.
func ExitForRestart(log *slog.Logger) {
	log.Info("Update applied — restarting via service manager")
	os.Exit(0)
}

// FormatNotification returns the user-facing update message.
func FormatNotification(info *UpdateInfo, currentVersion string) string {
	if !info.Available {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n══════════════════════════════════════════════════════════\n")
	fmt.Fprintf(&b, "  🔔 New version available: v%s (current: v%s)\n", info.Latest, currentVersion)
	b.WriteString("\n")
	if info.IsGitRepo {
		b.WriteString("  Update:\n")
		b.WriteString("    git pull && ./run.sh\n")
		b.WriteString("\n")
		b.WriteString("  Or download manually:\n")
	} else {
		b.WriteString("  Download the latest version:\n")
	}
	fmt.Fprintf(&b, "    %s\n", info.URL)
	b.WriteString("══════════════════════════════════════════════════════════\n")
	return b.String()
}

func isGitRepo() bool {
	_, err := os.Stat(".git")
	return err == nil
}
