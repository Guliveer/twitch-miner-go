// Package auth handles Twitch authentication, cookie persistence, and
// credential management for the miner.
package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/encryption"
)

// Cookie represents a single HTTP cookie persisted to JSON.
type Cookie struct {
	Name    string    `json:"name"`
	Value   string    `json:"value"`
	Domain  string    `json:"domain,omitempty"`
	Path    string    `json:"path,omitempty"`
	Expires time.Time `json:"expires,omitempty"`
}

// CookieJar manages a collection of cookies with thread-safe access
// and JSON persistence. It replaces the Python pickle-based cookie storage.
type CookieJar struct {
	mu            sync.RWMutex
	cookies       []Cookie
	encryptionKey []byte // nil means encryption is disabled
}

// NewCookieJar creates an empty CookieJar without encryption.
func NewCookieJar() *CookieJar {
	return &CookieJar{
		cookies: make([]Cookie, 0),
	}
}

// NewCookieJarWithEncryption creates an empty CookieJar that encrypts
// cookie values on save and decrypts them on load.
func NewCookieJarWithEncryption(key []byte) *CookieJar {
	return &CookieJar{
		cookies:       make([]Cookie, 0),
		encryptionKey: key,
	}
}

// Load reads cookies from a JSON file at the given path.
// Encrypted values are transparently decrypted when an encryption key is set.
// Returns an error if the file does not exist or cannot be parsed.
func (cj *CookieJar) Load(path string) error {
	cj.mu.Lock()
	defer cj.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading cookie file %s: %w", path, err)
	}

	var cookies []Cookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return fmt.Errorf("parsing cookie file %s: %w", path, err)
	}

	for i, c := range cookies {
		decrypted, err := cj.maybeDecrypt(c.Value)
		if err != nil {
			return fmt.Errorf("decrypting cookie %q from %s: %w", c.Name, path, err)
		}
		cookies[i].Value = decrypted
	}

	cj.cookies = cookies
	return nil
}

// Save writes the current cookies to a JSON file at the given path.
// When an encryption key is configured, every cookie value is encrypted.
// It creates parent directories if they do not exist.
// Uses atomic write (write to temp file, then rename) to prevent corruption.
func (cj *CookieJar) Save(path string) error {
	cj.mu.RLock()
	defer cj.mu.RUnlock()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating cookie directory %s: %w", dir, err)
	}

	toSave := make([]Cookie, len(cj.cookies))
	for i, c := range cj.cookies {
		encoded, err := cj.maybeEncrypt(c.Value)
		if err != nil {
			return fmt.Errorf("encrypting cookie %q: %w", c.Name, err)
		}
		toSave[i] = c
		toSave[i].Value = encoded
	}

	data, err := json.MarshalIndent(toSave, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cookies: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("writing cookie file %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("replacing cookie file %s: %w", path, err)
	}

	return nil
}

// Get returns the value of a cookie by name, or empty string if not found.
func (cj *CookieJar) Get(name string) string {
	cj.mu.RLock()
	defer cj.mu.RUnlock()

	for _, cookie := range cj.cookies {
		if cookie.Name == name && cookie.Value != "" {
			return cookie.Value
		}
	}
	return ""
}

// Set adds or updates a cookie by name.
func (cj *CookieJar) Set(name, value string) {
	cj.mu.Lock()
	defer cj.mu.Unlock()

	for i, cookie := range cj.cookies {
		if cookie.Name == name {
			cj.cookies[i].Value = value
			return
		}
	}
	cj.cookies = append(cj.cookies, Cookie{
		Name:   name,
		Value:  value,
		Domain: ".twitch.tv",
		Path:   "/",
	})
}

// All returns a copy of all cookies.
func (cj *CookieJar) All() []Cookie {
	cj.mu.RLock()
	defer cj.mu.RUnlock()

	result := make([]Cookie, len(cj.cookies))
	copy(result, cj.cookies)
	return result
}

// Len returns the number of cookies in the jar.
func (cj *CookieJar) Len() int {
	cj.mu.RLock()
	defer cj.mu.RUnlock()
	return len(cj.cookies)
}

// Clear removes all cookies from the jar.
func (cj *CookieJar) Clear() {
	cj.mu.Lock()
	defer cj.mu.Unlock()
	cj.cookies = make([]Cookie, 0)
}

// HasEncryption reports whether the jar is configured to encrypt on save.
func (cj *CookieJar) HasEncryption() bool {
	cj.mu.RLock()
	defer cj.mu.RUnlock()
	return cj.encryptionKey != nil
}

// CookieFileExists checks if a cookie file exists at the given path.
func CookieFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// maybeEncrypt encrypts value when an encryption key is set.
// Returns the original value unchanged when encryption is disabled.
func (cj *CookieJar) maybeEncrypt(value string) (string, error) {
	if cj.encryptionKey == nil {
		return value, nil
	}
	return encryption.Encrypt(cj.encryptionKey, value)
}

// maybeDecrypt decrypts an encrypted value when an encryption key is set.
// Returns the original value unchanged when it is not encrypted.
// Returns an error when an encrypted value is found but no key is configured.
func (cj *CookieJar) maybeDecrypt(value string) (string, error) {
	if !encryption.IsEncrypted(value) {
		return value, nil
	}
	if cj.encryptionKey == nil {
		return "", fmt.Errorf("encrypted cookie found but no encryption key is configured")
	}
	return encryption.Decrypt(cj.encryptionKey, value)
}
