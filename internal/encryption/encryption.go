// Package encryption provides AES-256-GCM encryption and decryption for
// sensitive values stored on disk. Encrypted values use a self-describing
// format (enc:<algorithm>:<payload>) to allow future algorithm additions
// without breaking existing encrypted files.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

// Prefix marks a value as encrypted.
const Prefix = "enc:"

// Algorithm identifiers.
const (
	AlgAES256GCM = "aes-256-gcm"
)

// ParseKey decodes a Base64-encoded AES-256 key and validates its length.
func ParseKey(encoded string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid Base64 encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}
	return key, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns the
// self-describing encoded form: enc:aes-256-gcm:<base64(nonce||ciphertext)>.
func Encrypt(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	result := make([]byte, 0, gcm.NonceSize()+len(ciphertext))
	result = append(result, nonce...)
	result = append(result, ciphertext...)
	payload := base64.StdEncoding.EncodeToString(result)
	return Prefix + AlgAES256GCM + ":" + payload, nil
}

// Decrypt decodes an enc:<algorithm>:<payload> value and returns the plaintext.
func Decrypt(key []byte, encoded string) (string, error) {
	alg, payload, err := Parse(encoded)
	if err != nil {
		return "", err
	}

	switch alg {
	case AlgAES256GCM:
		return decryptAES256GCM(key, payload)
	default:
		return "", fmt.Errorf("unsupported encryption algorithm: %s", alg)
	}
}

// IsEncrypted reports whether value uses the enc: prefix format.
func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, Prefix)
}

// Parse splits an encoded value into algorithm and raw Base64 payload.
func Parse(encoded string) (algorithm, payload string, err error) {
	if !strings.HasPrefix(encoded, Prefix) {
		return "", "", fmt.Errorf("value does not use the enc: prefix format")
	}

	rest := strings.TrimPrefix(encoded, Prefix)
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("malformed encrypted value: expected enc:<algorithm>:<payload>")
	}
	return parts[0], parts[1], nil
}

func decryptAES256GCM(key []byte, payload string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("invalid Base64 payload: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("corrupted or truncated encrypted payload")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed — corrupted payload or wrong key: %w", err)
	}

	return string(plaintext), nil
}
