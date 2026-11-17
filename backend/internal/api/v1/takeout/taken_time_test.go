package takeout

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTimedFile(t *testing.T, initialModTime time.Time) string {
	t.Helper()
	filePath := filepath.Join(t.TempDir(), "photo.jpg")
	if err := os.WriteFile(filePath, []byte("photo"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.Chtimes(filePath, initialModTime, initialModTime); err != nil {
		t.Fatalf("set initial times: %v", err)
	}
	return filePath
}

func modTimeOf(t *testing.T, filePath string) time.Time {
	t.Helper()
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	return info.ModTime()
}

func TestApplyTakenTimeToFilePrefersPhotoTakenTime(t *testing.T) {
	initialModTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	filePath := writeTimedFile(t, initialModTime)

	applyTakenTimeToFile(filePath, &TakeoutMetadata{
		PhotoTakenTime: takeoutTimestamp{Timestamp: "1500000000"},
		CreationTime:   takeoutTimestamp{Timestamp: "1600000000"},
	})

	if got := modTimeOf(t, filePath); !got.Equal(time.Unix(1500000000, 0)) {
		t.Fatalf("expected photoTakenTime, got %v", got)
	}
}

func TestApplyTakenTimeToFileFallsBackToCreationTime(t *testing.T) {
	filePath := writeTimedFile(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))

	applyTakenTimeToFile(filePath, &TakeoutMetadata{CreationTime: takeoutTimestamp{Timestamp: "1600000000"}})

	if got := modTimeOf(t, filePath); !got.Equal(time.Unix(1600000000, 0)) {
		t.Fatalf("expected creationTime, got %v", got)
	}
}

func TestApplyTakenTimeToFileLeavesFileUntouchedWithoutValidTime(t *testing.T) {
	initialModTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	invalidTimestamps := []string{"", "abc", "0", "-5", "1.5e9"}

	for _, invalidTimestamp := range invalidTimestamps {
		filePath := writeTimedFile(t, initialModTime)
		applyTakenTimeToFile(filePath, &TakeoutMetadata{
			PhotoTakenTime: takeoutTimestamp{Timestamp: invalidTimestamp},
			CreationTime:   takeoutTimestamp{Timestamp: invalidTimestamp},
		})
		if got := modTimeOf(t, filePath); !got.Equal(initialModTime) {
			t.Fatalf("timestamp %q must not change the file time, got %v", invalidTimestamp, got)
		}
	}

	filePath := writeTimedFile(t, initialModTime)
	applyTakenTimeToFile(filePath, nil)
	if got := modTimeOf(t, filePath); !got.Equal(initialModTime) {
		t.Fatalf("nil metadata must not change the file time, got %v", got)
	}
}
