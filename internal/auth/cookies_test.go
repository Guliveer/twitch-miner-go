package auth

import (
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/encryption"
)

func testEncryptionKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generating test key: %v", err)
	}
	return key
}

func writeCookieFile(t *testing.T, path string, cookies []Cookie) {
	t.Helper()
	data, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		t.Fatalf("marshaling cookies: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating dir: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("writing file: %v", err)
	}
}

func TestLoadPlaintextCookies(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.json")

	writeCookieFile(t, path, []Cookie{
		{Name: "auth-token", Value: "tok123", Domain: ".twitch.tv", Path: "/"},
		{Name: "persistent", Value: "999", Domain: ".twitch.tv", Path: "/"},
	})

	jar := NewCookieJar()
	if err := jar.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if jar.Len() != 2 {
		t.Fatalf("expected 2 cookies, got %d", jar.Len())
	}
	if v := jar.Get("auth-token"); v != "tok123" {
		t.Fatalf("auth-token: got %q", v)
	}
	if v := jar.Get("persistent"); v != "999" {
		t.Fatalf("persistent: got %q", v)
	}
}

func TestLoadEncryptedCookies(t *testing.T) {
	key := testEncryptionKey(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.json")

	encrypted, err := encryption.Encrypt(key, "secret-token")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	writeCookieFile(t, path, []Cookie{
		{Name: "auth-token", Value: encrypted, Domain: ".twitch.tv", Path: "/"},
	})

	jar := NewCookieJarWithEncryption(key)
	if err := jar.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if v := jar.Get("auth-token"); v != "secret-token" {
		t.Fatalf("expected decrypted value, got %q", v)
	}
}

func TestSavePlaintextWhenEncryptionDisabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.json")

	jar := NewCookieJar()
	jar.Set("auth-token", "plaintext-value")
	if err := jar.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading saved file: %v", err)
	}
	if strings.Contains(string(data), encryption.Prefix) {
		t.Fatal("saved file should not contain encrypted values")
	}

	jar2 := NewCookieJar()
	if err := jar2.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if v := jar2.Get("auth-token"); v != "plaintext-value" {
		t.Fatalf("expected plaintext-value, got %q", v)
	}
}

func TestSaveEncryptedWhenEncryptionEnabled(t *testing.T) {
	key := testEncryptionKey(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.json")

	jar := NewCookieJarWithEncryption(key)
	jar.Set("auth-token", "secret-value")
	if err := jar.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading saved file: %v", err)
	}
	if !strings.Contains(string(data), encryption.Prefix+encryption.AlgAES256GCM+":") {
		t.Fatal("saved file should contain encrypted values")
	}
	// Plaintext should NOT appear in the file.
	if strings.Contains(string(data), "secret-value") {
		t.Fatal("plaintext secret-value should not appear in saved file")
	}
}

func TestAutoMigrationPlaintextToEncrypted(t *testing.T) {
	key := testEncryptionKey(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.json")

	// Start with a plaintext cookie file.
	writeCookieFile(t, path, []Cookie{
		{Name: "auth-token", Value: "old-plaintext", Domain: ".twitch.tv", Path: "/"},
	})

	// Load with encryption key — should decrypt (plaintext stays as-is).
	jar := NewCookieJarWithEncryption(key)
	if err := jar.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if v := jar.Get("auth-token"); v != "old-plaintext" {
		t.Fatalf("expected old-plaintext, got %q", v)
	}

	// Save — should now encrypt.
	if err := jar.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading: %v", err)
	}
	if !strings.Contains(string(data), encryption.Prefix) {
		t.Fatal("file should be encrypted after save")
	}

	// Reload — should still work.
	jar2 := NewCookieJarWithEncryption(key)
	if err := jar2.Load(path); err != nil {
		t.Fatalf("Load after migration: %v", err)
	}
	if v := jar2.Get("auth-token"); v != "old-plaintext" {
		t.Fatalf("expected old-plaintext after migration, got %q", v)
	}
}

func TestEncryptedCookieWithoutKey(t *testing.T) {
	key := testEncryptionKey(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.json")

	encrypted, err := encryption.Encrypt(key, "value")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	writeCookieFile(t, path, []Cookie{
		{Name: "auth-token", Value: encrypted, Domain: ".twitch.tv", Path: "/"},
	})

	jar := NewCookieJar() // no key
	err = jar.Load(path)
	if err == nil {
		t.Fatal("expected error when loading encrypted cookie without key")
	}
	if !strings.Contains(err.Error(), "encrypted cookie found but no encryption key is configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWrongDecryptionKey(t *testing.T) {
	key1 := testEncryptionKey(t)
	key2 := testEncryptionKey(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.json")

	encrypted, err := encryption.Encrypt(key1, "secret")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	writeCookieFile(t, path, []Cookie{
		{Name: "auth-token", Value: encrypted, Domain: ".twitch.tv", Path: "/"},
	})

	jar := NewCookieJarWithEncryption(key2)
	err = jar.Load(path)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
}

func TestMixedPlaintextAndEncryptedCookies(t *testing.T) {
	key := testEncryptionKey(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.json")

	encrypted, err := encryption.Encrypt(key, "enc-val")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	writeCookieFile(t, path, []Cookie{
		{Name: "auth-token", Value: encrypted, Domain: ".twitch.tv", Path: "/"},
		{Name: "persistent", Value: "12345", Domain: ".twitch.tv", Path: "/"},
	})

	jar := NewCookieJarWithEncryption(key)
	if err := jar.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if v := jar.Get("auth-token"); v != "enc-val" {
		t.Fatalf("auth-token: got %q", v)
	}
	if v := jar.Get("persistent"); v != "12345" {
		t.Fatalf("persistent: got %q", v)
	}
}

func TestHasEncryption(t *testing.T) {
	jar := NewCookieJar()
	if jar.HasEncryption() {
		t.Fatal("NewCookieJar should not have encryption")
	}

	key := testEncryptionKey(t)
	jar2 := NewCookieJarWithEncryption(key)
	if !jar2.HasEncryption() {
		t.Fatal("NewCookieJarWithEncryption should have encryption")
	}
}

func TestGetReturnsEmptyForMissingCookie(t *testing.T) {
	jar := NewCookieJar()
	if v := jar.Get("nonexistent"); v != "" {
		t.Fatalf("expected empty, got %q", v)
	}
}

func TestLoadThenSaveEncryptsPlaintextFile(t *testing.T) {
	key := testEncryptionKey(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.json")

	writeCookieFile(t, path, []Cookie{
		{Name: "auth-token", Value: "plain-token", Domain: ".twitch.tv", Path: "/"},
	})

	jar := NewCookieJarWithEncryption(key)
	if err := jar.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if v := jar.Get("auth-token"); v != "plain-token" {
		t.Fatalf("before save: got %q", v)
	}

	if err := jar.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading: %v", err)
	}
	if strings.Contains(string(data), "plain-token") {
		t.Fatal("plaintext should not remain after save with encryption")
	}
	if !strings.Contains(string(data), encryption.Prefix) {
		t.Fatal("file should contain encrypted values")
	}

	jar2 := NewCookieJarWithEncryption(key)
	if err := jar2.Load(path); err != nil {
		t.Fatalf("Load after migration: %v", err)
	}
	if v := jar2.Get("auth-token"); v != "plain-token" {
		t.Fatalf("after migration: got %q", v)
	}
}

func TestSaveReplacesFileAtomically(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.json")

	jar := NewCookieJar()
	jar.Set("auth-token", "first-value")
	if err := jar.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	jar.Set("auth-token", "second-value")
	if err := jar.Save(path); err != nil {
		t.Fatalf("Save (overwrite): %v", err)
	}

	// A leftover temp file means the rename step did not run, which is what
	// keeps a kill mid-write from truncating live credentials.
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("expected no leftover temp file, stat err = %v", err)
	}

	reloaded := NewCookieJar()
	if err := reloaded.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if v := reloaded.Get("auth-token"); v != "second-value" {
		t.Fatalf("expected second-value, got %q", v)
	}
}
