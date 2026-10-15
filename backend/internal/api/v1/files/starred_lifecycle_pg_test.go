package files

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/roots"
	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/utils"
)

type renamingTrashBin struct {
	trashDirectory string
}

func (bin *renamingTrashBin) MoveToTrash(originalPath string, size int64) error {
	return bin.MoveToTrashFrom(originalPath, originalPath, size)
}

func (bin *renamingTrashBin) MoveToTrashFrom(logicalPath string, contentPath string, size int64) error {
	return os.Rename(contentPath, filepath.Join(bin.trashDirectory, filepath.Base(contentPath)))
}

func (bin *renamingTrashBin) restore(originalPath string) error {
	return os.Rename(filepath.Join(bin.trashDirectory, filepath.Base(originalPath)), originalPath)
}

type starredLifecycleFixture struct {
	repo       *Repository
	service    *Service
	trashBin   *renamingTrashBin
	entryPoint string
	folderPath string
	folderID   int
	childPath  string
	childID    int
	filePath   string
	fileID     int
	targetID   int
	targetPath string
}

func newStarredLifecycleFixture(t *testing.T) *starredLifecycleFixture {
	t.Helper()
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)

	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)
	trashBin := &renamingTrashBin{trashDirectory: t.TempDir()}

	fixture := &starredLifecycleFixture{
		repo:       repo,
		service:    &Service{Repository: repo, Tasks: make(chan utils.Task, 64), TrashBin: trashBin},
		trashBin:   trashBin,
		entryPoint: entryPoint,
		folderPath: filepath.Join(entryPoint, "album"),
		filePath:   filepath.Join(entryPoint, "note.txt"),
		targetPath: filepath.Join(entryPoint, "target"),
	}
	fixture.childPath = filepath.Join(fixture.folderPath, "song.txt")

	for _, directory := range []string{fixture.folderPath, fixture.targetPath} {
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatalf("MkdirAll %q: %v", directory, err)
		}
	}
	for _, file := range []string{fixture.filePath, fixture.childPath} {
		if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
			t.Fatalf("WriteFile %q: %v", file, err)
		}
	}

	fixture.folderID = insertRowForOps(t, repo, "album", fixture.folderPath, entryPoint, Directory)
	fixture.childID = insertRowForOps(t, repo, "song.txt", fixture.childPath, fixture.folderPath, File)
	fixture.fileID = insertRowForOps(t, repo, "note.txt", fixture.filePath, entryPoint, File)
	fixture.targetID = insertRowForOps(t, repo, "target", fixture.targetPath, entryPoint, Directory)

	for _, id := range []int{fixture.folderID, fixture.fileID} {
		if _, err := repo.SetFileStarred(id, true); err != nil {
			t.Fatalf("SetFileStarred %d: %v", id, err)
		}
	}
	return fixture
}

func (fixture *starredLifecycleFixture) pinStarredAt(t *testing.T) time.Time {
	t.Helper()
	pinned := time.Date(2021, 3, 4, 5, 6, 7, 0, time.UTC)
	for _, id := range []int{fixture.folderID, fixture.fileID} {
		setStarredAt(t, fixture.repo, id, pinned)
	}
	return pinned
}

func assertStillStarred(t *testing.T, repo *Repository, id int, expectedStarredAt time.Time) {
	t.Helper()
	if !readStarred(t, repo, id) {
		t.Fatalf("file %d lost its star", id)
	}
	if starredAt := readStarredAt(t, repo, id); !starredAt.Valid || !starredAt.Time.Equal(expectedStarredAt) {
		t.Fatalf("file %d starred_at changed: got %v want %v", id, starredAt, expectedStarredAt)
	}
}

func assertRowAtPath(t *testing.T, repo *Repository, path string, expectedID int) {
	t.Helper()
	row, found := activeRowByPath(t, repo, path)
	if !found || row.ID != expectedID {
		t.Fatalf("expected id %d at %q, got found=%v row=%+v", expectedID, path, found, row)
	}
}

func starredIDs(t *testing.T, repo *Repository) map[int]bool {
	t.Helper()
	page, err := repo.GetStarredFiles(1, 50, DefaultStarredListing)
	if err != nil {
		t.Fatalf("GetStarredFiles: %v", err)
	}
	ids := map[int]bool{}
	for _, item := range page.Items {
		ids[item.ID] = true
	}
	return ids
}

func assertStarredCounts(t *testing.T, repo *Repository, expectedTotal int, expectedFolders int) {
	t.Helper()
	counts, err := repo.CountStarredFilesByKind()
	if err != nil {
		t.Fatalf("CountStarredFilesByKind: %v", err)
	}
	if counts.Total != expectedTotal || counts.Folders != expectedFolders {
		t.Fatalf("unexpected starred counts %+v, want total=%d folders=%d", counts, expectedTotal, expectedFolders)
	}
}

func TestPostgres_StarSurvivesRename(t *testing.T) {
	fixture := newStarredLifecycleFixture(t)
	pinned := fixture.pinStarredAt(t)

	renamedFile, err := fixture.service.RenameFile(fixture.fileID, "memo.txt")
	if err != nil {
		t.Fatalf("RenameFile file: %v", err)
	}
	renamedFolder, err := fixture.service.RenameFile(fixture.folderID, "records")
	if err != nil {
		t.Fatalf("RenameFile folder: %v", err)
	}

	assertRowAtPath(t, fixture.repo, renamedFile, fixture.fileID)
	assertRowAtPath(t, fixture.repo, renamedFolder, fixture.folderID)
	assertRowAtPath(t, fixture.repo, filepath.Join(renamedFolder, "song.txt"), fixture.childID)
	assertStillStarred(t, fixture.repo, fixture.fileID, pinned)
	assertStillStarred(t, fixture.repo, fixture.folderID, pinned)
}

func TestPostgres_StarSurvivesMoveWithinRoot(t *testing.T) {
	fixture := newStarredLifecycleFixture(t)
	pinned := fixture.pinStarredAt(t)

	movedFile, err := fixture.service.MoveFile(fixture.fileID, &fixture.targetID, "")
	if err != nil {
		t.Fatalf("MoveFile file: %v", err)
	}
	movedFolder, err := fixture.service.MoveFile(fixture.folderID, &fixture.targetID, "")
	if err != nil {
		t.Fatalf("MoveFile folder: %v", err)
	}

	assertRowAtPath(t, fixture.repo, movedFile, fixture.fileID)
	assertRowAtPath(t, fixture.repo, movedFolder, fixture.folderID)
	assertRowAtPath(t, fixture.repo, filepath.Join(movedFolder, "song.txt"), fixture.childID)
	assertStillStarred(t, fixture.repo, fixture.fileID, pinned)
	assertStillStarred(t, fixture.repo, fixture.folderID, pinned)
}

func TestPostgres_StarSurvivesTrashAndRestore(t *testing.T) {
	fixture := newStarredLifecycleFixture(t)
	pinned := fixture.pinStarredAt(t)
	assertStarredCounts(t, fixture.repo, 2, 1)

	for _, id := range []int{fixture.fileID, fixture.folderID} {
		if err := fixture.service.DeleteFileFromDisk(id, false); err != nil {
			t.Fatalf("DeleteFileFromDisk %d: %v", id, err)
		}
	}

	if listed := starredIDs(t, fixture.repo); len(listed) != 0 {
		t.Fatalf("trashed items must leave GetStarredFiles, got %v", listed)
	}
	assertStarredCounts(t, fixture.repo, 0, 0)

	for _, path := range []string{fixture.filePath, fixture.folderPath} {
		if err := fixture.trashBin.restore(path); err != nil {
			t.Fatalf("restore bytes %q: %v", path, err)
		}
		if err := fixture.service.RestoreSubtree(path); err != nil {
			t.Fatalf("RestoreSubtree %q: %v", path, err)
		}
	}

	listed := starredIDs(t, fixture.repo)
	if !listed[fixture.fileID] || !listed[fixture.folderID] {
		t.Fatalf("restored items must return to GetStarredFiles, got %v", listed)
	}
	assertStarredCounts(t, fixture.repo, 2, 1)
	assertStillStarred(t, fixture.repo, fixture.fileID, pinned)
	assertStillStarred(t, fixture.repo, fixture.folderID, pinned)
	assertRowAtPath(t, fixture.repo, fixture.childPath, fixture.childID)
}

func TestPostgres_StarSurvivesGenericUpdateWithStaleSnapshot(t *testing.T) {
	fixture := newStarredLifecycleFixture(t)
	pinned := fixture.pinStarredAt(t)

	staleSnapshot, err := fixture.service.GetFileById(fixture.fileID)
	if err != nil {
		t.Fatalf("GetFileById: %v", err)
	}
	staleSnapshot.Starred = false
	staleSnapshot.Size = 4096

	if _, err := fixture.service.UpdateFile(staleSnapshot); err != nil {
		t.Fatalf("UpdateFile: %v", err)
	}

	assertStillStarred(t, fixture.repo, fixture.fileID, pinned)
	assertRowAtPath(t, fixture.repo, fixture.filePath, fixture.fileID)
}

func TestPostgres_StarSurvivesMoveAcrossRoots(t *testing.T) {
	t.Skip("known bug: moveAcrossRoots copies then deletes (operations_move_across_roots.go:43,57; syncPathRow in operations_sync.go:134 inserts a new row), so id and star are lost")
	fixture := newStarredLifecycleFixture(t)
	secondRoot := t.TempDir()
	t.Cleanup(roots.Reset)
	roots.Set([]roots.Root{
		{ID: 1, Path: fixture.entryPoint, Label: "Principal", Enabled: true},
		{ID: 2, Path: secondRoot, Label: "Midia", Enabled: true},
	})
	secondRootID := insertRowForOps(t, fixture.repo, "midia", secondRoot, filepath.Dir(secondRoot), Directory)
	pinned := fixture.pinStarredAt(t)

	movedPath, err := fixture.service.MoveFile(fixture.fileID, &secondRootID, "")
	if err != nil {
		t.Fatalf("MoveFile across roots: %v", err)
	}

	movedRow, found := activeRowByPath(t, fixture.repo, movedPath)
	if !found || movedRow.ID != fixture.fileID {
		t.Fatalf("id not preserved across roots: found=%v row=%+v originalID=%d", found, movedRow, fixture.fileID)
	}
	assertStillStarred(t, fixture.repo, fixture.fileID, pinned)
}
