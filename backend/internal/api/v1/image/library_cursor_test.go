package image

import (
	"errors"
	"testing"
	"time"
)

func TestLibraryCursorRoundTripDated(t *testing.T) {
	takenAt := time.Date(2022, 3, 4, 5, 6, 7, 123456000, time.UTC)
	encodedCursor := LibraryCursor{TakenAt: &takenAt, FileID: 42}.Encode()

	decoded, err := DecodeLibraryCursor(encodedCursor)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.TakenAt == nil || !decoded.TakenAt.Equal(takenAt) || decoded.FileID != 42 {
		t.Fatalf("unexpected cursor %+v", decoded)
	}
}

func TestLibraryCursorRoundTripUndated(t *testing.T) {
	decoded, err := DecodeLibraryCursor(LibraryCursor{FileID: 7}.Encode())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.TakenAt != nil || decoded.FileID != 7 {
		t.Fatalf("unexpected cursor %+v", decoded)
	}
}

func TestDecodeLibraryCursorRejectsGarbage(t *testing.T) {
	for _, garbage := range []string{"!!!", "bm8tc2VwYXJhdG9y", "YXwx", "bnxhYmM", "MTIzfC0x", "bnwtMQ"} {
		if _, err := DecodeLibraryCursor(garbage); !errors.Is(err, ErrInvalidLibraryCursor) {
			t.Fatalf("cursor %q: expected ErrInvalidLibraryCursor, got %v", garbage, err)
		}
	}
}

func TestCursorFromItemUsesTakenAtAndFileID(t *testing.T) {
	takenAt := time.Now()
	cursor := CursorFromItem(LibraryItemModel{FileID: 9, TakenAt: &takenAt})
	if cursor.FileID != 9 || cursor.TakenAt != &takenAt {
		t.Fatalf("unexpected cursor %+v", cursor)
	}
}
