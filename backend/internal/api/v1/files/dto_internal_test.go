package files

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileDto_ParseAndChecksums(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "sample.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to write sample file: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("failed to read temp dir entries: %v", err)
	}

	dto := FileDto{}
	if err := dto.ParseDirEntryToFileDto(entries[0]); err != nil {
		t.Fatalf("expected parse dir entry success, err=%v", err)
	}
	if dto.Name == "" {
		t.Fatalf("expected parsed entry name")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	dto2 := FileDto{}
	if err := dto2.ParseFileInfoToFileDto(info); err != nil {
		t.Fatalf("expected parse file info success, err=%v", err)
	}
	if dto2.Type != File || dto2.Format != ".txt" {
		t.Fatalf("expected file type and format, got type=%v format=%s", dto2.Type, dto2.Format)
	}
	if !dto2.LastInteraction.HasValue {
		t.Fatalf("expected last interaction to be set")
	}

	dto2.Path = filePath
	sum, err := dto2.GetCheckSumFromFile()
	if err != nil || sum == "" {
		t.Fatalf("expected file checksum, err=%v", err)
	}
	combined := dto2.GetCheckSumFromPath([]string{sum, "invalid"})
	if combined == "" {
		t.Fatalf("expected combined checksum")
	}
}

func TestRecentFileDtoModelConversions(t *testing.T) {
	now := time.Now()
	dto := RecentFileDto{
		ID:         1,
		IPAddress:  "127.0.0.1",
		FileID:     9,
		AccessedAt: now,
	}
	model := dto.ToModel()
	if model.FileID != 9 || model.IPAddress != "127.0.0.1" {
		t.Fatalf("unexpected model conversion result: %+v", model)
	}

	back := model.ToDto()
	if back.ID != dto.ID || back.FileID != dto.FileID {
		t.Fatalf("unexpected dto conversion result: %+v", back)
	}
}
