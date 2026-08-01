package encryption

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		keyID   string
		key     string
		wantErr bool
	}{
		{name: "valid", keyID: "primary", key: testKeyBase64(t), wantErr: false},
		{name: "missing key id", keyID: "", key: testKeyBase64(t), wantErr: true},
		{name: "missing key", keyID: "primary", key: "", wantErr: true},
		{name: "bad base64", keyID: "primary", key: "not-base64", wantErr: true},
		{name: "wrong length", keyID: "primary", key: base64.StdEncoding.EncodeToString([]byte("too short")), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := New(tt.keyID, tt.key)
			if (err != nil) != tt.wantErr {
				t.Fatalf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && mgr == nil {
				t.Fatal("New() returned nil manager")
			}
		})
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	mgr := mustManager(t, "primary", testKeyBase64(t))

	plaintext := "Kareelio encrypt me"
	encrypted, err := mgr.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if encrypted == plaintext {
		t.Fatal("Encrypt() returned plaintext unchanged")
	}
	if !strings.HasPrefix(encrypted, EnvelopeVersion+":primary:") {
		t.Fatalf("Encrypt() envelope = %q, want version/key id prefix", encrypted)
	}

	got, err := mgr.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if got != plaintext {
		t.Fatalf("Decrypt() = %q, want %q", got, plaintext)
	}
}

func TestEmptyStringPassthrough(t *testing.T) {
	mgr := mustManager(t, "primary", testKeyBase64(t))

	encrypted, err := mgr.Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if encrypted != "" {
		t.Fatalf("Encrypt(\"\") = %q, want empty string", encrypted)
	}

	decrypted, err := mgr.Decrypt("")
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if decrypted != "" {
		t.Fatalf("Decrypt(\"\") = %q, want empty string", decrypted)
	}
}

func TestDecryptWithWrongKeyFails(t *testing.T) {
	encryptor := mustManager(t, "primary", testKeyBase64(t))
	wrongKey := mustManager(t, "primary", otherKeyBase64(t))

	encrypted, err := encryptor.Encrypt("secret")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if _, err := wrongKey.Decrypt(encrypted); err == nil {
		t.Fatal("Decrypt() with wrong key unexpectedly succeeded")
	}
}

func TestDecryptWithCorruptCiphertextFails(t *testing.T) {
	mgr := mustManager(t, "primary", testKeyBase64(t))

	encrypted, err := mgr.Encrypt("secret")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	parts := strings.Split(encrypted, ":")
	if len(parts) != 3 {
		t.Fatalf("unexpected envelope %q", encrypted)
	}
	decoded, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}
	decoded[len(decoded)-1] ^= 0x01
	parts[2] = base64.StdEncoding.EncodeToString(decoded)
	corrupt := strings.Join(parts, ":")

	if _, err := mgr.Decrypt(corrupt); err == nil {
		t.Fatal("Decrypt() with corrupt ciphertext unexpectedly succeeded")
	}
}

func TestDecryptInvalidEnvelope(t *testing.T) {
	mgr := mustManager(t, "primary", testKeyBase64(t))

	for _, encrypted := range []string{"not-envelope", "v1", "v1:primary", "v1:other:"} {
		t.Run(encrypted, func(t *testing.T) {
			if _, err := mgr.Decrypt(encrypted); err == nil {
				t.Fatalf("Decrypt(%q) unexpectedly succeeded", encrypted)
			}
		})
	}
}

func testKeyBase64(t *testing.T) string {
	t.Helper()
	return base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
}

func otherKeyBase64(t *testing.T) string {
	t.Helper()
	return base64.StdEncoding.EncodeToString([]byte("fedcba9876543210fedcba9876543210"))
}

func mustManager(t *testing.T, keyID, key string) *Manager {
	t.Helper()
	mgr, err := New(keyID, key)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return mgr
}
