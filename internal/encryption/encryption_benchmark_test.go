package encryption

import (
	"encoding/base64"
	"testing"
)

func benchKey() []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return key
}

// BenchmarkEncrypt benchmarks AES-256-GCM encryption of a cookie value.
func BenchmarkEncrypt(b *testing.B) {
	key := benchKey()
	plaintext := "b42dm850fv8jok7ksuwlh81ozomh19"

	for b.Loop() {
		_, _ = Encrypt(key, plaintext)
	}
}

// BenchmarkDecrypt benchmarks AES-256-GCM decryption of a stored cookie value.
func BenchmarkDecrypt(b *testing.B) {
	key := benchKey()
	encrypted, _ := Encrypt(key, "b42dm850fv8jok7ksuwlh81ozomh19")

	for b.Loop() {
		_, _ = Decrypt(key, encrypted)
	}
}

// BenchmarkParseKey benchmarks parsing the Base64-encoded encryption key.
func BenchmarkParseKey(b *testing.B) {
	encoded := base64.StdEncoding.EncodeToString(benchKey())

	for b.Loop() {
		_, _ = ParseKey(encoded)
	}
}
