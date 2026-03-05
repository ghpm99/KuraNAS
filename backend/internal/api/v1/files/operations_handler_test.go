package files

import (
	"path/filepath"
	"testing"

	"nas-go/api/internal/config"
)

func TestResolvePathInEntryPoint(t *testing.T) {
	tempDir := t.TempDir()
	config.AppConfig.EntryPoint = tempDir

	path, err := resolvePathInEntryPoint("")
	if err != nil {
		t.Fatalf("expected entry point path, got error: %v", err)
	}
	if path != filepath.Clean(tempDir) {
		t.Fatalf("expected %s, got %s", filepath.Clean(tempDir), path)
	}

	relativePath := "docs/file.txt"
	resolvedRelative, err := resolvePathInEntryPoint(relativePath)
	if err != nil {
		t.Fatalf("expected valid relative path, got error: %v", err)
	}
	expectedRelative := filepath.Join(filepath.Clean(tempDir), filepath.FromSlash(relativePath))
	if resolvedRelative != expectedRelative {
		t.Fatalf("expected %s, got %s", expectedRelative, resolvedRelative)
	}

	if _, err := resolvePathInEntryPoint("../outside"); err == nil {
		t.Fatalf("expected error for path outside entry point")
	}
}
