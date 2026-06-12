package crypto

import (
	"bytes"
	"encoding/base64"
	"errors"
	"testing"
)

func testKey() string {
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x42}, 32))
}

func TestNewAESGCMRejectsBadKeys(t *testing.T) {
	cases := []string{
		"",
		"not-base64!!!",
		base64.StdEncoding.EncodeToString([]byte("short")),
		base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 16)),
	}
	for _, key := range cases {
		if _, err := NewAESGCM(key); !errors.Is(err, ErrInvalidKey) {
			t.Fatalf("key %q: expected ErrInvalidKey, got %v", key, err)
		}
	}
}

func TestSealOpenRoundTrip(t *testing.T) {
	c, err := NewAESGCM(testKey())
	if err != nil {
		t.Fatalf("NewAESGCM: %v", err)
	}

	plaintexts := [][]byte{
		[]byte(""),
		[]byte("a"),
		[]byte(`{"access_token":"x","refresh_token":"y"}`),
		bytes.Repeat([]byte{0xff}, 4096),
	}

	for _, plaintext := range plaintexts {
		sealed, err := c.Seal(plaintext)
		if err != nil {
			t.Fatalf("Seal: %v", err)
		}
		if len(plaintext) > 8 && bytes.Contains(sealed, plaintext) {
			t.Fatal("sealed blob contains the plaintext")
		}

		opened, err := c.Open(sealed)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		if !bytes.Equal(opened, plaintext) {
			t.Fatalf("round trip mismatch: got %q want %q", opened, plaintext)
		}
	}
}

func TestSealProducesDistinctNonces(t *testing.T) {
	c, _ := NewAESGCM(testKey())
	a, _ := c.Seal([]byte("same input"))
	b, _ := c.Seal([]byte("same input"))
	if bytes.Equal(a, b) {
		t.Fatal("two Seal calls produced identical ciphertexts (nonce reuse)")
	}
}

func TestOpenRejectsTamperedAndShortBlobs(t *testing.T) {
	c, _ := NewAESGCM(testKey())

	if _, err := c.Open([]byte{1, 2, 3}); !errors.Is(err, ErrInvalidCiphertext) {
		t.Fatalf("short blob: expected ErrInvalidCiphertext, got %v", err)
	}

	sealed, _ := c.Seal([]byte("secret"))
	sealed[len(sealed)-1] ^= 0x01
	if _, err := c.Open(sealed); !errors.Is(err, ErrInvalidCiphertext) {
		t.Fatalf("tampered blob: expected ErrInvalidCiphertext, got %v", err)
	}
}

func TestOpenRejectsWrongKey(t *testing.T) {
	c1, _ := NewAESGCM(testKey())
	otherKey := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x24}, 32))
	c2, _ := NewAESGCM(otherKey)

	sealed, _ := c1.Seal([]byte("secret"))
	if _, err := c2.Open(sealed); !errors.Is(err, ErrInvalidCiphertext) {
		t.Fatalf("wrong key: expected ErrInvalidCiphertext, got %v", err)
	}
}
