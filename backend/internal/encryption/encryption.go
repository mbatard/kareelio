package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	EnvelopeVersion = "v1"
	keySizeBytes    = 32
)

var (
	ErrEmptyKeyID       = errors.New("data encryption key id is required")
	ErrEmptyKey         = errors.New("data encryption key is required")
	ErrInvalidEnvelope  = errors.New("invalid encrypted payload")
	ErrKeyIDMismatch    = errors.New("encrypted payload key id does not match active key id")
	ErrInvalidKeyLength = errors.New("data encryption key must decode to 32 bytes")
)

type Manager struct {
	keyID string
	aead  cipher.AEAD
}

func New(keyID, base64Key string) (*Manager, error) {
	if strings.TrimSpace(keyID) == "" {
		return nil, ErrEmptyKeyID
	}
	if strings.TrimSpace(base64Key) == "" {
		return nil, ErrEmptyKey
	}

	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("decode data encryption key: %w", err)
	}
	if len(key) != keySizeBytes {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidKeyLength, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("initialize cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("initialize AEAD: %w", err)
	}

	return &Manager{keyID: keyID, aead: aead}, nil
}

func (m *Manager) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	nonce := make([]byte, m.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := m.aead.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append(nonce, ciphertext...)

	return fmt.Sprintf("%s:%s:%s", EnvelopeVersion, m.keyID, base64.StdEncoding.EncodeToString(payload)), nil
}

func (m *Manager) Decrypt(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	version, rest, ok := strings.Cut(encrypted, ":")
	if !ok || version != EnvelopeVersion {
		return "", ErrInvalidEnvelope
	}

	keyID, encodedPayload, ok := strings.Cut(rest, ":")
	if !ok || strings.TrimSpace(keyID) == "" || strings.TrimSpace(encodedPayload) == "" {
		return "", ErrInvalidEnvelope
	}
	if keyID != m.keyID {
		return "", ErrKeyIDMismatch
	}

	payload, err := base64.StdEncoding.DecodeString(encodedPayload)
	if err != nil {
		return "", fmt.Errorf("decode encrypted payload: %w", err)
	}
	if len(payload) < m.aead.NonceSize() {
		return "", ErrInvalidEnvelope
	}

	nonce := payload[:m.aead.NonceSize()]
	ciphertext := payload[m.aead.NonceSize():]
	plaintext, err := m.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt payload: %w", err)
	}

	return string(plaintext), nil
}
