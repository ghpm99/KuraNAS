// Package crypto provides AES-256-GCM sealing for secrets at rest (e-mail
// OAuth tokens). The key is a 32-byte value supplied base64-encoded via env;
// the random nonce is prefixed to the ciphertext so each blob is
// self-contained.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidKey        = errors.New("crypto: key must be 32 bytes base64-encoded")
	ErrInvalidCiphertext = errors.New("crypto: ciphertext is malformed or was tampered with")
)

// AESGCM seals and opens byte blobs with a fixed 256-bit key.
type AESGCM struct {
	aead cipher.AEAD
}

// NewAESGCM builds a cipher from a base64-encoded 32-byte key.
func NewAESGCM(base64Key string) (*AESGCM, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil || len(key) != 32 {
		return nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: %w", err)
	}

	return &AESGCM{aead: aead}, nil
}

// Seal encrypts plaintext and returns nonce||ciphertext.
func (a *AESGCM) Seal(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, a.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: %w", err)
	}
	return a.aead.Seal(nonce, nonce, plaintext, nil), nil
}

// Open decrypts a blob produced by Seal.
func (a *AESGCM) Open(sealed []byte) ([]byte, error) {
	nonceSize := a.aead.NonceSize()
	if len(sealed) < nonceSize {
		return nil, ErrInvalidCiphertext
	}

	plaintext, err := a.aead.Open(nil, sealed[:nonceSize], sealed[nonceSize:], nil)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}
	return plaintext, nil
}
