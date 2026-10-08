package files

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

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
	prefixRes, err := repo.GetFilesByPathPrefix(winParent, 1, 500)
	if err != nil {
		t.Fatalf("GetFilesByPathPrefix error: %v", err)
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

// TestPostgres_DeletedSemanticsOfDecomposedQueries proves each decomposed
// query carries the soft-delete intent it declares: the tree listing hides
// soft-deleted rows (the bug was that they leaked into the tree), while the
// prefix walk used by mark_deleted sees every state.
func TestPostgres_DeletedSemanticsOfDecomposedQueries(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	truncateHomeFile(t, repo)

	parent := "/srv/dados"
	activePath := parent + "/ativo.txt"
	deletedPath := parent + "/deletado.txt"
	mod := time.Date(2026, 6, 11, 10, 0, 0, 0, time.UTC)
	insertFileRow(t, repo, "ativo.txt", activePath, parent, 10, mod)
	insertFileRow(t, repo, "deletado.txt", deletedPath, parent, 20, mod)

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, e := tx.Exec("UPDATE home_file SET deleted_at = now() WHERE path = $1", deletedPath)
		return e
	})
	if err != nil {
		t.Fatalf("soft-delete row: %v", err)
	}

	children, err := repo.GetActiveChildrenByParentPath(parent, AllCategory, DefaultChildrenSort, 1, 50)
	if err != nil {
		t.Fatalf("GetActiveChildrenByParentPath: %v", err)
	}
	if len(children.Items) != 1 || children.Items[0].Path != activePath {
		t.Fatalf("tree listing must hide soft-deleted rows, got %+v", children.Items)
	}

	byPath, err := repo.GetActiveFilesByPath(deletedPath, 1, 10)
	if err != nil {
		t.Fatalf("GetActiveFilesByPath: %v", err)
	}
	if len(byPath.Items) != 0 {
		t.Fatalf("path lookup must hide soft-deleted rows, got %+v", byPath.Items)
	}

	countsByParentPath, err := repo.GetDirectoryContentCounts([]string{parent, "/srv/vazio"})
	if err != nil {
		t.Fatalf("GetDirectoryContentCounts: %v", err)
	}
	if countsByParentPath[parent] != 1 || countsByParentPath["/srv/vazio"] != 0 {
		t.Fatalf("content counts must ignore soft-deleted rows, got %v", countsByParentPath)
	}

	walk, err := repo.GetFilesByPathPrefix(parent, 1, 50)
	if err != nil {
		t.Fatalf("GetFilesByPathPrefix: %v", err)
	}
	if len(walk.Items) != 2 {
		t.Fatalf("prefix walk must see every state, got %+v", walk.Items)
	}

	byNamePath, err := repo.GetFilesByNameAndPath("deletado.txt", deletedPath, 5)
	if err != nil {
		t.Fatalf("GetFilesByNameAndPath: %v", err)
	}
	if len(byNamePath) != 1 || !byNamePath[0].DeletedAt.Valid {
		t.Fatalf("name+path lookup must see the soft-deleted row, got %+v", byNamePath)
	}
}

func TestPostgres_ChildrenSortKeepsDirectoriesFirstAndOrdersByKey(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	truncateHomeFile(t, repo)

	parent := "/srv/ordenacao"
	baseTime := time.Date(2026, 6, 11, 10, 0, 0, 0, time.UTC)
	insertFileRow(t, repo, "b.txt", parent+"/b.txt", parent, 30, baseTime.Add(2*time.Hour))
	insertFileRow(t, repo, "a.txt", parent+"/a.txt", parent, 20, baseTime.Add(3*time.Hour))
	insertFileRow(t, repo, "c.txt", parent+"/c.txt", parent, 10, baseTime.Add(1*time.Hour))
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, createErr := repo.CreateFile(tx, FileModel{
			Name: "z-dir", Path: parent + "/z-dir", ParentPath: parent,
			Size: 0, UpdatedAt: baseTime, CreatedAt: baseTime, Type: Directory,
		})
		return createErr
	})
	if err != nil {
		t.Fatalf("insert directory: %v", err)
	}

	tests := []struct {
		childrenSort  ChildrenSort
		expectedNames []string
	}{
		{DefaultChildrenSort, []string{"z-dir", "a.txt", "b.txt", "c.txt"}},
		{ChildrenSort{SortByName, SortDescending}, []string{"z-dir", "c.txt", "b.txt", "a.txt"}},
		{ChildrenSort{SortBySize, SortAscending}, []string{"z-dir", "c.txt", "a.txt", "b.txt"}},
		{ChildrenSort{SortBySize, SortDescending}, []string{"z-dir", "b.txt", "a.txt", "c.txt"}},
		{ChildrenSort{SortByUpdatedAt, SortDescending}, []string{"z-dir", "a.txt", "b.txt", "c.txt"}},
		{ChildrenSort{SortByCreatedAt, SortAscending}, []string{"z-dir", "c.txt", "b.txt", "a.txt"}},
	}

	for _, testCase := range tests {
		page, err := repo.GetActiveChildrenByParentPath(parent, AllCategory, testCase.childrenSort, 1, 50)
		if err != nil {
			t.Fatalf("sort %+v: %v", testCase.childrenSort, err)
		}
		names := make([]string, 0, len(page.Items))
		for _, child := range page.Items {
			names = append(names, child.Name)
		}
		if !slices.Equal(names, testCase.expectedNames) {
			t.Fatalf("sort %+v: expected %v, got %v", testCase.childrenSort, testCase.expectedNames, names)
		}
	}
}

func TestPostgres_StarredAndRecentListAcrossFoldersAndSkipDeleted(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	truncateHomeFile(t, repo)

	mod := time.Date(2026, 6, 11, 10, 0, 0, 0, time.UTC)
	insertFileRow(t, repo, "raiz.txt", "/srv/raiz.txt", "/srv", 1, mod)
	insertFileRow(t, repo, "fundo.txt", "/srv/a/b/fundo.txt", "/srv/a/b", 1, mod)
	insertFileRow(t, repo, "apagado.txt", "/srv/a/apagado.txt", "/srv/a", 1, mod)
	insertFileRow(t, repo, "comum.txt", "/srv/comum.txt", "/srv", 1, mod)

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		statements := []string{
			"UPDATE home_file SET starred = TRUE WHERE name IN ('raiz.txt', 'fundo.txt', 'apagado.txt')",
			"UPDATE home_file SET deleted_at = now() WHERE name = 'apagado.txt'",
			"TRUNCATE recent_file",
			"INSERT INTO recent_file (ip_address, file_id, accessed_at) SELECT '10.0.0.1', id, now() - interval '3 hours' FROM home_file WHERE name = 'raiz.txt'",
			"INSERT INTO recent_file (ip_address, file_id, accessed_at) SELECT '10.0.0.2', id, now() - interval '1 hours' FROM home_file WHERE name = 'raiz.txt'",
			"INSERT INTO recent_file (ip_address, file_id, accessed_at) SELECT '10.0.0.1', id, now() - interval '2 hours' FROM home_file WHERE name = 'fundo.txt'",
			"INSERT INTO recent_file (ip_address, file_id, accessed_at) SELECT '10.0.0.1', id, now() FROM home_file WHERE name = 'apagado.txt'",
		}
		for _, statement := range statements {
			if _, execErr := tx.Exec(statement); execErr != nil {
				return execErr
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed starred and recent: %v", err)
	}

	starred, err := repo.GetStarredFiles(1, 50)
	if err != nil {
		t.Fatalf("GetStarredFiles: %v", err)
	}
	starredNames := []string{}
	for _, file := range starred.Items {
		starredNames = append(starredNames, file.Name)
	}
	if !slices.Equal(starredNames, []string{"fundo.txt", "raiz.txt"}) {
		t.Fatalf("starred must span folders and hide deleted rows, got %v", starredNames)
	}

	recent, err := repo.GetRecentlyAccessedFiles(1, 50)
	if err != nil {
		t.Fatalf("GetRecentlyAccessedFiles: %v", err)
	}
	recentNames := []string{}
	for _, file := range recent.Items {
		recentNames = append(recentNames, file.Name)
	}
	if !slices.Equal(recentNames, []string{"raiz.txt", "fundo.txt"}) {
		t.Fatalf("recent must be distinct, newest first and hide deleted rows, got %v", recentNames)
	}

	firstRecentPage, err := repo.GetRecentlyAccessedFiles(1, 1)
	if err != nil || len(firstRecentPage.Items) != 1 || !firstRecentPage.Pagination.HasNext {
		t.Fatalf("expected paginated recent with next page, got %+v err=%v", firstRecentPage, err)
	}
}
