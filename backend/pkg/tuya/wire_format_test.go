package tuya

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

const expectedReadFrameHex = "000055aa000000010000000a00000058d652f952e5585d3b065bf07e1e6e0045ea20e4fe63da2c4037ab6fb2d200ea2b49b9c47a3d78df0d20268e27f855ae020d6a62524157b9d5cabcfef954189fb20efea48a0091daa595a3a780fe13bd0e910edf7e0000aa55"

func TestReadFrameMatchesFixedWireVector(t *testing.T) {
	device := Device{Address: "x", DeviceID: "bf1234", LocalKey: testKey}

	encoded, err := buildReadFrame(device, 1, time.Unix(1700000000, 0))

	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(encoded) != expectedReadFrameHex {
		t.Fatalf("wire format changed:\n got %x\nwant %s", encoded, expectedReadFrameHex)
	}
}

func TestFixedVectorIsReadBackByFrameReader(t *testing.T) {
	encoded, _ := hex.DecodeString(expectedReadFrameHex)

	received, err := readFrame(bytes.NewReader(encoded))

	if err != nil || received.command != 0x0a || received.sequence != 1 {
		t.Fatalf("unexpected outcome: %+v %v", received, err)
	}
	plaintext, err := received.decrypt(testKey)
	if err != nil || !bytes.Contains(plaintext, []byte(`"gwId":"bf1234"`)) {
		t.Fatalf("unexpected plaintext: %s %v", plaintext, err)
	}
}

func TestReadFrameRejectsCorruptedFrames(t *testing.T) {
	valid, _ := hex.DecodeString(expectedReadFrameHex)
	corruptions := map[string]func([]byte){
		"prefix":   func(frameBytes []byte) { frameBytes[0] = 0xff },
		"checksum": func(frameBytes []byte) { frameBytes[20] ^= 0xff },
		"suffix":   func(frameBytes []byte) { frameBytes[len(frameBytes)-1] = 0 },
		"size":     func(frameBytes []byte) { frameBytes[12] = 0xff },
		"tiny":     func(frameBytes []byte) { frameBytes[15] = 2 },
	}
	for name, corrupt := range corruptions {
		corrupted := append([]byte{}, valid...)
		corrupt(corrupted)
		if _, err := readFrame(bytes.NewReader(corrupted)); !errors.Is(err, errInvalidFrame) {
			t.Errorf("%s: expected errInvalidFrame, got %v", name, err)
		}
	}
}

func TestReadFrameFailsOnTruncatedInput(t *testing.T) {
	valid, _ := hex.DecodeString(expectedReadFrameHex)

	if _, err := readFrame(bytes.NewReader(valid[:30])); err == nil {
		t.Fatal("expected an error for truncated frame")
	}
}

func TestWriteFramePlacesVersionHeaderBeforeCiphertext(t *testing.T) {
	device := Device{Address: "x", DeviceID: "bf1234", LocalKey: testKey}

	encoded, err := buildWriteFrame(device, 2, time.Unix(1700000000, 0), map[string]any{"115": "gettime"})

	if err != nil {
		t.Fatal(err)
	}
	payload := encoded[16 : len(encoded)-8]
	if !bytes.HasPrefix(payload, append([]byte("3.3"), make([]byte, 12)...)) || (len(payload)-15)%16 != 0 {
		t.Fatalf("unexpected payload layout: %x", payload)
	}
}

func TestDecryptRejectsBadCiphertext(t *testing.T) {
	if _, err := decryptECB([]byte(testKey), []byte("short")); !errors.Is(err, errInvalidCiphertext) {
		t.Fatalf("expected errInvalidCiphertext, got %v", err)
	}
	if _, err := decryptECB([]byte(testKey), bytes.Repeat([]byte{1}, 16)); !errors.Is(err, errInvalidPadding) {
		t.Fatalf("expected errInvalidPadding, got %v", err)
	}
	if _, err := encryptECB([]byte("bad"), []byte("x")); err == nil {
		t.Fatal("expected error for invalid key size")
	}
	if _, err := decryptECB([]byte("bad"), bytes.Repeat([]byte{1}, 16)); err == nil {
		t.Fatal("expected error for invalid key size")
	}
}
