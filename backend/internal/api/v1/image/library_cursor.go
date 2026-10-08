package image

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	cursorSeparator   = "|"
	undatedCursorMark = "n"
)

var ErrInvalidLibraryCursor = errors.New("invalid library cursor")

// LibraryCursor is the keyset position (taken_at, file_id) after which the
// next gallery page starts. TakenAt is nil for images without a capture date.
type LibraryCursor struct {
	TakenAt *time.Time
	FileID  int
}

const undatedSortKey = "-infinity"

func (cursor LibraryCursor) sortKey() any {
	if cursor.TakenAt == nil {
		return undatedSortKey
	}
	return *cursor.TakenAt
}

func CursorFromItem(item LibraryItemModel) LibraryCursor {
	return LibraryCursor{TakenAt: item.TakenAt, FileID: item.FileID}
}

func (cursor LibraryCursor) Encode() string {
	takenAtPart := undatedCursorMark
	if cursor.TakenAt != nil {
		takenAtPart = strconv.FormatInt(cursor.TakenAt.UnixMicro(), 10)
	}
	rawCursor := takenAtPart + cursorSeparator + strconv.Itoa(cursor.FileID)
	return base64.RawURLEncoding.EncodeToString([]byte(rawCursor))
}

func DecodeLibraryCursor(encodedCursor string) (LibraryCursor, error) {
	decodedBytes, err := base64.RawURLEncoding.DecodeString(encodedCursor)
	if err != nil {
		return LibraryCursor{}, fmt.Errorf("%w: %v", ErrInvalidLibraryCursor, err)
	}

	takenAtPart, fileIDPart, hasSeparator := strings.Cut(string(decodedBytes), cursorSeparator)
	if !hasSeparator {
		return LibraryCursor{}, ErrInvalidLibraryCursor
	}

	fileID, err := strconv.Atoi(fileIDPart)
	if err != nil || fileID < 0 {
		return LibraryCursor{}, ErrInvalidLibraryCursor
	}

	if takenAtPart == undatedCursorMark {
		return LibraryCursor{FileID: fileID}, nil
	}

	takenAtMicros, err := strconv.ParseInt(takenAtPart, 10, 64)
	if err != nil {
		return LibraryCursor{}, ErrInvalidLibraryCursor
	}
	takenAt := time.UnixMicro(takenAtMicros).UTC()
	return LibraryCursor{TakenAt: &takenAt, FileID: fileID}, nil
}
