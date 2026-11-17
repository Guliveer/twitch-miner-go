package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCheckForUpdate_NewerAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ghRelease{TagName: "v1.2.0"})
	}))
	defer srv.Close()

	info, err := checkWithURL(context.Background(), "1.0.0", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Available {
		t.Error("expected update to be available")
	}
	if info.Latest != "1.2.0" {
		t.Errorf("Latest = %q, want %q", info.Latest, "1.2.0")
	}
}

func TestCheckForUpdate_UpToDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ghRelease{TagName: "v1.0.0"})
	}))
	defer srv.Close()

	info, err := checkWithURL(context.Background(), "1.0.0", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Available {
		t.Error("expected no update available")
	}
}

func TestCheckForUpdate_DevVersion(t *testing.T) {
	info, err := checkWithURL(context.Background(), "dev", "http://unused")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Available {
		t.Error("expected no update for dev version")
	}
}

func TestCheckForUpdate_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := checkWithURL(context.Background(), "1.0.0", srv.URL)
	if err == nil {
		t.Error("expected error on server failure")
	}
}

func TestCheckForUpdate_PopulatesAssetURL(t *testing.T) {
	assetName, err := platformAsset()
	if err != nil {
		t.Skipf("unsupported platform: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ghRelease{
			TagName: "v9.0.0",
			Assets: []ghAsset{
				{Name: assetName, BrowserDownloadURL: "https://example.com/" + assetName},
			},
		})
	}))
	defer srv.Close()

	info, err := checkWithURL(context.Background(), "1.0.0", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Available {
		t.Fatal("expected update to be available")
	}
	if info.AssetURL == "" {
		t.Error("expected AssetURL to be populated")
	}
}

func TestCheckForUpdate_NoMatchingAsset(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ghRelease{
			TagName: "v9.0.0",
			Assets:  []ghAsset{{Name: "twitch-miner-go-plan9-mips64", BrowserDownloadURL: "https://example.com/plan9"}},
		})
	}))
	defer srv.Close()

	info, err := checkWithURL(context.Background(), "1.0.0", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.AssetURL != "" {
		t.Errorf("expected empty AssetURL for unmatched platform, got %q", info.AssetURL)
	}
}

func TestPlatformAsset(t *testing.T) {
	supported := map[string]bool{
		"amd64": true,
		"arm64": true,
	}
	if !supported[runtime.GOARCH] {
		_, err := platformAsset()
		if err == nil {
			t.Errorf("expected error for unsupported arch %s", runtime.GOARCH)
		}
		return
	}

	name, err := platformAsset()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name == "" {
		t.Error("expected non-empty asset name")
	}

	expectedPrefix := "twitch-miner-go-" + runtime.GOOS + "-" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		expectedPrefix += ".exe"
	}
	if name != expectedPrefix {
		t.Errorf("platformAsset() = %q, want %q", name, expectedPrefix)
	}
}

func TestDownloadTo(t *testing.T) {
	content := []byte("fake binary content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer srv.Close()

	tmp, err := downloadTo(context.Background(), srv.URL, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(tmp)

	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("read temp file: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("file content = %q, want %q", data, content)
	}

	info, err := os.Stat(tmp)
	if err != nil {
		t.Fatalf("stat temp file: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0111 == 0 {
		t.Errorf("expected executable bit set, got %v", info.Mode().Perm())
	}
}

func TestDownloadTo_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := downloadTo(context.Background(), srv.URL, t.TempDir())
	if err == nil {
		t.Error("expected error on 404")
	}
}

func TestReplaceBinary(t *testing.T) {
	// Write a "current binary" to a temp file.
	exe, err := os.CreateTemp(t.TempDir(), "fake-exe-*")
	if err != nil {
		t.Fatal(err)
	}
	exe.WriteString("old content")
	exe.Close()
	exePath := exe.Name()

	// Write the "new binary" to another temp file.
	newBin, err := os.CreateTemp(t.TempDir(), "new-bin-*")
	if err != nil {
		t.Fatal(err)
	}
	newBin.WriteString("new content")
	newBin.Close()
	newPath := newBin.Name()

	// Patch os.Executable to return our fake path.
	origExecutable := osExecutable
	osExecutable = func() (string, error) { return exePath, nil }
	defer func() { osExecutable = origExecutable }()

	if err := ReplaceBinary(newPath); err != nil {
		t.Fatalf("ReplaceBinary: %v", err)
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatalf("read replaced file: %v", err)
	}
	if string(data) != "new content" {
		t.Errorf("replaced content = %q, want %q", data, "new content")
	}
}

func TestReplaceWindows(t *testing.T) {
	dir := t.TempDir()
	exePath := filepath.Join(dir, "twitch-miner-go.exe")
	if err := os.WriteFile(exePath, []byte("old content"), 0o755); err != nil {
		t.Fatal(err)
	}
	tmpPath := filepath.Join(dir, ".update")
	if err := os.WriteFile(tmpPath, []byte("new content"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := replaceWindows(exePath, tmpPath); err != nil {
		t.Fatalf("replaceWindows: %v", err)
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatalf("read replaced file: %v", err)
	}
	if string(data) != "new content" {
		t.Errorf("replaced content = %q, want %q", data, "new content")
	}

	old, err := os.ReadFile(exePath + oldBinarySuffix)
	if err != nil {
		t.Fatalf("read previous binary: %v", err)
	}
	if string(old) != "old content" {
		t.Errorf("previous content = %q, want %q", old, "old content")
	}
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("expected downloaded file to be moved, stat err = %v", err)
	}
}

func TestReplaceWindows_RestoresBinaryOnFailure(t *testing.T) {
	dir := t.TempDir()
	exePath := filepath.Join(dir, "twitch-miner-go.exe")
	if err := os.WriteFile(exePath, []byte("old content"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := replaceWindows(exePath, filepath.Join(dir, "missing")); err == nil {
		t.Fatal("expected error when the downloaded file is gone")
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if string(data) != "old content" {
		t.Errorf("restored content = %q, want %q", data, "old content")
	}
	if _, err := os.Stat(exePath + oldBinarySuffix); !os.IsNotExist(err) {
		t.Errorf("expected previous binary to be moved back, stat err = %v", err)
	}
}
