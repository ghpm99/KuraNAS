package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFile(t *testing.T, path string, content string, mtime time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !mtime.IsZero() {
		if err := os.Chtimes(path, mtime, mtime); err != nil {
			t.Fatalf("chtimes: %v", err)
		}
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func setup(t *testing.T) (Root, string) {
	t.Helper()
	source := t.TempDir()
	dest := t.TempDir()
	return Root{Label: "Casa", Path: source}, dest
}

func TestRunCopiesNewFilesAndStamps(t *testing.T) {
	root, dest := setup(t)
	mtime := time.Now().Add(-time.Hour)
	writeFile(t, filepath.Join(root.Path, "docs", "a.txt"), "conteudo-a", mtime)
	writeFile(t, filepath.Join(root.Path, "b.txt"), "conteudo-b", mtime)

	var stamped []string
	stats, err := Run(Options{
		Roots:         []Root{root},
		Destination:   dest,
		RetentionDays: 30,
		Stamp: func(sourcePath string, at time.Time) error {
			stamped = append(stamped, sourcePath)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if stats.Copied != 2 || stats.Failures != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	copied := readFile(t, filepath.Join(dest, CurrentDirName, "Casa", "docs", "a.txt"))
	if copied != "conteudo-a" {
		t.Fatalf("unexpected copy content: %q", copied)
	}
	if len(stamped) != 2 {
		t.Fatalf("expected 2 stamps, got %v", stamped)
	}
}

func TestRunSkipsUnchangedFiles(t *testing.T) {
	root, dest := setup(t)
	mtime := time.Now().Add(-time.Hour)
	writeFile(t, filepath.Join(root.Path, "a.txt"), "same", mtime)

	if _, err := Run(Options{Roots: []Root{root}, Destination: dest, RetentionDays: 30}); err != nil {
		t.Fatalf("first run: %v", err)
	}

	stats, err := Run(Options{Roots: []Root{root}, Destination: dest, RetentionDays: 30})
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if stats.Copied != 0 || stats.Scanned != 1 {
		t.Fatalf("expected incremental skip, got %+v", stats)
	}
}

func TestRunVersionsChangedFiles(t *testing.T) {
	root, dest := setup(t)
	old := time.Now().Add(-2 * time.Hour)
	writeFile(t, filepath.Join(root.Path, "a.txt"), "versao-1", old)

	if _, err := Run(Options{Roots: []Root{root}, Destination: dest, RetentionDays: 30}); err != nil {
		t.Fatalf("first run: %v", err)
	}

	writeFile(t, filepath.Join(root.Path, "a.txt"), "versao-2-maior", time.Now())
	stats, err := Run(Options{Roots: []Root{root}, Destination: dest, RetentionDays: 30})
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if stats.Copied != 1 || stats.Versioned != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	current := readFile(t, filepath.Join(dest, CurrentDirName, "Casa", "a.txt"))
	if current != "versao-2-maior" {
		t.Fatalf("current copy not updated: %q", current)
	}

	versions := collectFiles(t, filepath.Join(dest, VersionsDirName))
	if len(versions) != 1 || readFile(t, versions[0]) != "versao-1" {
		t.Fatalf("previous version not preserved: %v", versions)
	}
}

func TestRunVersionsDeletedFiles(t *testing.T) {
	root, dest := setup(t)
	writeFile(t, filepath.Join(root.Path, "morto.txt"), "ainda-recuperavel", time.Now().Add(-time.Hour))

	if _, err := Run(Options{Roots: []Root{root}, Destination: dest, RetentionDays: 30}); err != nil {
		t.Fatalf("first run: %v", err)
	}

	if err := os.Remove(filepath.Join(root.Path, "morto.txt")); err != nil {
		t.Fatalf("remove source: %v", err)
	}

	stats, err := Run(Options{Roots: []Root{root}, Destination: dest, RetentionDays: 30})
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if stats.Versioned != 1 {
		t.Fatalf("expected deleted file versioned, got %+v", stats)
	}

	if _, statErr := os.Stat(filepath.Join(dest, CurrentDirName, "Casa", "morto.txt")); !os.IsNotExist(statErr) {
		t.Fatal("deleted file still present in current/")
	}
	versions := collectFiles(t, filepath.Join(dest, VersionsDirName))
	if len(versions) != 1 || readFile(t, versions[0]) != "ainda-recuperavel" {
		t.Fatalf("deleted file not recoverable: %v", versions)
	}
}

func TestRunPurgesExpiredVersions(t *testing.T) {
	root, dest := setup(t)
	versionsDir := filepath.Join(dest, VersionsDirName)

	now := time.Now().UTC()
	expired := now.Add(-40 * 24 * time.Hour).Format(versionStampFmt)
	fresh := now.Add(-5 * 24 * time.Hour).Format(versionStampFmt)
	writeFile(t, filepath.Join(versionsDir, expired, "Casa", "velho.txt"), "x", time.Time{})
	writeFile(t, filepath.Join(versionsDir, fresh, "Casa", "novo.txt"), "y", time.Time{})

	stats, err := Run(Options{Roots: []Root{root}, Destination: dest, RetentionDays: 30})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if stats.Purged != 1 {
		t.Fatalf("expected 1 purge, got %+v", stats)
	}
	if _, statErr := os.Stat(filepath.Join(versionsDir, expired)); !os.IsNotExist(statErr) {
		t.Fatal("expired version dir survived the purge")
	}
	if _, statErr := os.Stat(filepath.Join(versionsDir, fresh)); statErr != nil {
		t.Fatal("fresh version dir must survive the purge")
	}
}

func TestRunSkipsConfiguredDirNames(t *testing.T) {
	root, dest := setup(t)
	writeFile(t, filepath.Join(root.Path, ".kuranas-trash", "lixo.txt"), "x", time.Time{})
	writeFile(t, filepath.Join(root.Path, "vivo.txt"), "y", time.Time{})

	stats, err := Run(Options{
		Roots:         []Root{root},
		Destination:   dest,
		RetentionDays: 30,
		SkipDirNames:  []string{".kuranas-trash"},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if stats.Copied != 1 {
		t.Fatalf("expected only the live file copied, got %+v", stats)
	}
	if _, statErr := os.Stat(filepath.Join(dest, CurrentDirName, "Casa", ".kuranas-trash")); !os.IsNotExist(statErr) {
		t.Fatal("trash dir leaked into the backup")
	}
}

func TestRunCleansLeftoverTempFiles(t *testing.T) {
	root, dest := setup(t)
	orphan := filepath.Join(dest, CurrentDirName, "Casa", tmpPrefix+"123")
	writeFile(t, orphan, "lixo de um crash", time.Time{})

	if _, err := Run(Options{Roots: []Root{root}, Destination: dest, RetentionDays: 30}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, statErr := os.Stat(orphan); !os.IsNotExist(statErr) {
		t.Fatal("leftover temp file survived")
	}
}

func TestValidateDestination(t *testing.T) {
	root := Root{Label: "Casa", Path: "/mnt/dados"}

	if err := ValidateDestination("", []Root{root}); err == nil {
		t.Fatal("empty destination must be rejected")
	}
	if err := ValidateDestination("relativo/backup", []Root{root}); err == nil {
		t.Fatal("relative destination must be rejected")
	}
	if err := ValidateDestination("/mnt/dados/backup", []Root{root}); err == nil {
		t.Fatal("destination inside a root must be rejected")
	}
	if err := ValidateDestination("/mnt/dados", []Root{root}); err == nil {
		t.Fatal("destination equal to a root must be rejected")
	}
	if err := ValidateDestination("/mnt/backup", []Root{root}); err != nil {
		t.Fatalf("valid destination rejected: %v", err)
	}
	if err := ValidateDestination("/mnt/dadosX", []Root{root}); err != nil {
		t.Fatalf("sibling path sharing the prefix must be accepted: %v", err)
	}
}

func TestRunCountsUnreadableFilesAsFailures(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root: permission bits are not enforced")
	}
	root, dest := setup(t)
	blocked := filepath.Join(root.Path, "sem-leitura.txt")
	writeFile(t, blocked, "segredo", time.Time{})
	if err := os.Chmod(blocked, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	defer os.Chmod(blocked, 0o644)

	stats, err := Run(Options{Roots: []Root{root}, Destination: dest, RetentionDays: 30})
	if err != nil {
		t.Fatalf("Run must not abort on one bad file: %v", err)
	}
	if stats.Failures != 1 || stats.Copied != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func collectFiles(t *testing.T, dir string) []string {
	t.Helper()
	var found []string
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && !strings.HasPrefix(d.Name(), tmpPrefix) {
			found = append(found, path)
		}
		return nil
	})
	return found
}
