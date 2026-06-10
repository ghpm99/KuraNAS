package files

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"nas-go/api/pkg/utils"
	"nas-go/api/internal/testutil"
)

// fixtureScanDir resolves the real fixture folder shipped in the repo
// (tests/files_test/worker/testscan) regardless of the test's working dir.
func fixtureScanDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("cannot resolve caller path")
	}
	// this file: backend/internal/api/v1/files/ -> up 4 -> backend/
	backendRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
	dir := filepath.Join(backendRoot, "tests", "files_test", "worker", "testscan")
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("fixture dir not found at %s: %v", dir, err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("abs fixture dir: %v", err)
	}
	return abs
}

func truncateHomeFile(t *testing.T, repo *Repository) {
	t.Helper()
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, e := tx.Exec("TRUNCATE home_file RESTART IDENTITY CASCADE")
		return e
	})
	if err != nil {
		t.Fatalf("truncate home_file: %v", err)
	}
}

func insertFileRow(t *testing.T, repo *Repository, name, path, parent string, size int64, mod time.Time) {
	t.Helper()
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, e := repo.CreateFile(tx, FileModel{
			Name:       name,
			Path:       path,
			ParentPath: parent,
			Format:     filepath.Ext(name),
			Size:       size,
			UpdatedAt:  mod,
			CreatedAt:  mod,
			Type:       File,
		})
		return e
	})
	if err != nil {
		t.Fatalf("insert file row %q: %v", path, err)
	}
}

// TestPostgres_PathPrefixMatchesWindowsPaths covers the real root cause of
// "every file re-enqueued on every startup scan" and its fix. On a Windows
// server the stored path uses backslashes; PostgreSQL treats '\' as the LIKE
// escape character, so the original `path LIKE $prefix || '%'` matched ZERO
// rows for Windows paths. The fix switches the PathPrefix filter to a literal
// starts_with(), and the lookup used by the diff is exact equality.
//
// The test proves both, against a real database:
//   - the PathPrefix filter now finds the backslash-path row;
//   - the exact-match lookup (GetFileStatByPath) finds it with intact data.
func TestPostgres_PathPrefixMatchesWindowsPaths(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	truncateHomeFile(t, repo)

	winParent := `D:\Pasta`
	winPath := `D:\Pasta\72061450723014730295560719510667.pdf`
	mod := time.Date(2025, 4, 7, 9, 21, 18, 0, time.UTC)
	insertFileRow(t, repo, "72061450723014730295560719510667.pdf", winPath, winParent, 698566, mod)

	// PathPrefix (used by mark_deleted) must match the backslash path now that
	// the filter uses starts_with() instead of LIKE.
	prefixRes, err := repo.GetFiles(FileFilter{
		PathPrefix: utils.Optional[string]{HasValue: true, Value: winParent},
	}, 1, 500)
	if err != nil {
		t.Fatalf("GetFiles(PathPrefix) error: %v", err)
	}
	if len(prefixRes.Items) != 1 {
		t.Fatalf("expected PathPrefix to match the Windows-path row, got %d row(s)", len(prefixRes.Items))
	}
	if prefixRes.Items[0].Path != winPath {
		t.Fatalf("PathPrefix returned wrong row: %q", prefixRes.Items[0].Path)
	}

	// Exact match is immune to LIKE escaping and finds the row with intact data.
	stat, found, err := repo.GetFileStatByPath(winPath)
	if err != nil {
		t.Fatalf("GetFileStatByPath error: %v", err)
	}
	if !found {
		t.Fatalf("exact-match lookup must find the Windows-path row, got found=false")
	}
	if stat.Size != 698566 {
		t.Fatalf("size mismatch: got %d want 698566", stat.Size)
	}
	if !stat.UpdatedAt.Truncate(time.Second).Equal(mod.Truncate(time.Second)) {
		t.Fatalf("updated_at mismatch: got %v want %v", stat.UpdatedAt, mod)
	}
}

// TestPostgres_GetFileStatByPath_RecognizesUnchangedFixtureFiles indexes every
// real file under the fixture folder and then asserts the diff lookup sees each
// one as unchanged (found, with matching size + second-truncated mtime). This is
// exactly the signal executeDiffAgainstDBStep uses to SKIP a file, so it proves
// already-processed, untouched files are not re-sent to the pipeline.
func TestPostgres_GetFileStatByPath_RecognizesUnchangedFixtureFiles(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	truncateHomeFile(t, repo)

	root := fixtureScanDir(t)

	indexed := []string{}
	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		insertFileRow(t, repo, d.Name(), path, filepath.Dir(path), info.Size(), info.ModTime())
		indexed = append(indexed, path)
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk fixtures: %v", walkErr)
	}
	if len(indexed) == 0 {
		t.Fatalf("no fixture files found under %s", root)
	}

	for _, path := range indexed {
		info, statErr := os.Stat(path)
		if statErr != nil {
			t.Fatalf("stat %q: %v", path, statErr)
		}
		stat, found, lookupErr := repo.GetFileStatByPath(path)
		if lookupErr != nil {
			t.Fatalf("GetFileStatByPath(%q): %v", path, lookupErr)
		}
		if !found {
			t.Fatalf("indexed fixture not found by exact lookup: %s", path)
		}
		if stat.Size != info.Size() {
			t.Fatalf("size mismatch for %q: stored %d disk %d", path, stat.Size, info.Size())
		}
		if !stat.UpdatedAt.Truncate(time.Second).Equal(info.ModTime().Truncate(time.Second)) {
			t.Fatalf("mtime mismatch for %q: stored %v disk %v", path, stat.UpdatedAt, info.ModTime())
		}
	}
}
