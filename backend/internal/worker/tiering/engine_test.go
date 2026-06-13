package tiering

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestRun_DemotesIdleFile(t *testing.T) {
	dir := t.TempDir()
	hot := filepath.Join(dir, "hot", "Documentos", "a.txt")
	cold := filepath.Join(dir, "cold", "root", "Documentos", "a.txt")
	writeFile(t, hot, "hello cold tier")

	var recorded string
	setPhysical := func(fileID int, physicalPath string) error {
		if fileID != 7 {
			t.Fatalf("unexpected fileID %d", fileID)
		}
		recorded = physicalPath
		return nil
	}

	stats := Run(filepath.Join(dir, "cold"), nil, []Demotion{
		{FileID: 7, HotPath: hot, ColdPath: cold},
	}, setPhysical)

	if stats.Demoted != 1 || stats.Failures != 0 {
		t.Fatalf("unexpected stats %+v", stats)
	}
	if recorded != cold {
		t.Fatalf("physical_path recorded as %q, want %q", recorded, cold)
	}
	if _, err := os.Stat(hot); !os.IsNotExist(err) {
		t.Fatalf("hot copy should be gone, stat err = %v", err)
	}
	got, err := os.ReadFile(cold)
	if err != nil || string(got) != "hello cold tier" {
		t.Fatalf("cold copy content = %q, err = %v", got, err)
	}
}

func TestRun_PromotesFileBackToHot(t *testing.T) {
	dir := t.TempDir()
	hot := filepath.Join(dir, "hot", "Documentos", "a.txt")
	cold := filepath.Join(dir, "cold", "root", "Documentos", "a.txt")
	writeFile(t, cold, "warm me up")

	cleared := false
	setPhysical := func(fileID int, physicalPath string) error {
		if physicalPath != "" {
			t.Fatalf("promotion should clear physical_path, got %q", physicalPath)
		}
		cleared = true
		return nil
	}

	stats := Run(filepath.Join(dir, "cold"), []Promotion{
		{FileID: 3, HotPath: hot, ColdPath: cold},
	}, nil, setPhysical)

	if stats.Promoted != 1 || stats.Failures != 0 {
		t.Fatalf("unexpected stats %+v", stats)
	}
	if !cleared {
		t.Fatal("physical_path was not cleared")
	}
	if _, err := os.Stat(cold); !os.IsNotExist(err) {
		t.Fatalf("cold copy should be gone, stat err = %v", err)
	}
	got, err := os.ReadFile(hot)
	if err != nil || string(got) != "warm me up" {
		t.Fatalf("hot copy content = %q, err = %v", got, err)
	}
}

// A DB failure mid-demotion must keep the hot copy intact (the source of
// truth): the cold copy may linger, but the file is never lost.
func TestRun_DemotionDbFailureKeepsHotCopy(t *testing.T) {
	dir := t.TempDir()
	hot := filepath.Join(dir, "hot", "a.txt")
	cold := filepath.Join(dir, "cold", "a.txt")
	writeFile(t, hot, "do not lose me")

	setPhysical := func(int, string) error { return errors.New("db down") }

	stats := Run(filepath.Join(dir, "cold"), nil, []Demotion{
		{FileID: 1, HotPath: hot, ColdPath: cold},
	}, setPhysical)

	if stats.Demoted != 0 || stats.Failures != 1 {
		t.Fatalf("unexpected stats %+v", stats)
	}
	if got, err := os.ReadFile(hot); err != nil || string(got) != "do not lose me" {
		t.Fatalf("hot copy must survive a db failure, content = %q, err = %v", got, err)
	}
}

func TestRun_MissingSourceCountsAsFailure(t *testing.T) {
	dir := t.TempDir()
	stats := Run(filepath.Join(dir, "cold"), nil, []Demotion{
		{FileID: 1, HotPath: filepath.Join(dir, "nope.txt"), ColdPath: filepath.Join(dir, "cold", "nope.txt")},
	}, func(int, string) error { return nil })

	if stats.Failures != 1 || stats.Demoted != 0 {
		t.Fatalf("unexpected stats %+v", stats)
	}
}

func TestRun_RemovesLeftoverTempFiles(t *testing.T) {
	dir := t.TempDir()
	coldDir := filepath.Join(dir, "cold")
	leftover := filepath.Join(coldDir, "sub", tmpPrefix+"orphan")
	writeFile(t, leftover, "junk")

	Run(coldDir, nil, nil, func(int, string) error { return nil })

	if _, err := os.Stat(leftover); !os.IsNotExist(err) {
		t.Fatalf("leftover temp should be removed, stat err = %v", err)
	}
}
