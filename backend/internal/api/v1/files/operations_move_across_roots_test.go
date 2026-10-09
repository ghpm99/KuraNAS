package files

import (
	"database/sql"
	"nas-go/api/internal/roots"
	"nas-go/api/pkg/utils"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type crossRootFixture struct {
	primaryRoot    string
	secondRoot     string
	destinationDir string
	deletedPaths   []string
	service        *Service
	destinationID  int
}

func newCrossRootFixture(t *testing.T) *crossRootFixture {
	t.Helper()
	fixture := &crossRootFixture{
		primaryRoot: t.TempDir(),
		secondRoot:  t.TempDir(),
	}
	setEntryPointForTest(t, fixture.primaryRoot)
	t.Cleanup(roots.Reset)
	roots.Set([]roots.Root{
		{ID: 1, Path: fixture.primaryRoot, Label: "Principal", Enabled: true},
		{ID: 2, Path: fixture.secondRoot, Label: "Midia", Enabled: true},
	})

	fixture.destinationDir = filepath.Join(fixture.secondRoot, "filmes")
	if err := os.Mkdir(fixture.destinationDir, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	fixture.destinationID = 99
	return fixture
}

func (fixture *crossRootFixture) buildService(t *testing.T, sourceRecords []FileModel) {
	t.Helper()
	records := append(sourceRecords, FileModel{ID: fixture.destinationID, Name: "filmes", Path: fixture.destinationDir, Type: Directory})
	repo := &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			for _, record := range records {
				if record.ID == id {
					return record, true, nil
				}
			}
			return FileModel{}, false, nil
		},
		markDeletedSubtreeFn: func(transaction *sql.Tx, path string, deletedAt time.Time) (int64, error) {
			fixture.deletedPaths = append(fixture.deletedPaths, path)
			return 1, nil
		},
	}
	fixture.service = newFilesServiceForTest(t, repo)
	fixture.service.Tasks = make(chan utils.Task, 16)
	fixture.service.JobsRepository = newFilesJobsRepoMockForTest(t)
}

func TestMoveFileAcrossRootsMovesFileContent(t *testing.T) {
	fixture := newCrossRootFixture(t)
	sourcePath := filepath.Join(fixture.primaryRoot, "video.mp4")
	if err := os.WriteFile(sourcePath, []byte("bytes"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	fixture.buildService(t, []FileModel{{ID: 1, Name: "video.mp4", Path: sourcePath, Type: File, CheckSum: "stale-but-present"}})

	destinationFolderID := fixture.destinationID
	movedPath, err := fixture.service.MoveFile(1, &destinationFolderID, "")
	if err != nil {
		t.Fatalf("MoveFile returned error: %v", err)
	}

	expectedPath := filepath.Join(fixture.destinationDir, "video.mp4")
	if movedPath != expectedPath {
		t.Fatalf("expected %q, got %q", expectedPath, movedPath)
	}
	movedBytes, readErr := os.ReadFile(expectedPath)
	if readErr != nil || string(movedBytes) != "bytes" {
		t.Fatalf("destination content wrong: %q, %v", movedBytes, readErr)
	}
	if _, statErr := os.Stat(sourcePath); !os.IsNotExist(statErr) {
		t.Fatalf("expected source removed, stat err: %v", statErr)
	}
	if len(fixture.deletedPaths) != 1 || fixture.deletedPaths[0] != sourcePath {
		t.Fatalf("expected source rows soft-deleted, got %v", fixture.deletedPaths)
	}
}

func TestMoveFileAcrossRootsMovesFolderTree(t *testing.T) {
	fixture := newCrossRootFixture(t)
	sourceDir := filepath.Join(fixture.primaryRoot, "serie")
	if err := os.MkdirAll(filepath.Join(sourceDir, "temporada1", "vazia"), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "temporada1", "ep1.mkv"), []byte("episode"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	fixture.buildService(t, []FileModel{{ID: 1, Name: "serie", Path: sourceDir, Type: Directory}})

	destinationFolderID := fixture.destinationID
	if _, err := fixture.service.MoveFile(1, &destinationFolderID, ""); err != nil {
		t.Fatalf("MoveFile returned error: %v", err)
	}

	movedEpisode := filepath.Join(fixture.destinationDir, "serie", "temporada1", "ep1.mkv")
	if movedBytes, err := os.ReadFile(movedEpisode); err != nil || string(movedBytes) != "episode" {
		t.Fatalf("destination episode wrong: %q, %v", movedBytes, err)
	}
	if _, err := os.Stat(filepath.Join(fixture.destinationDir, "serie", "temporada1", "vazia")); err != nil {
		t.Fatalf("empty folder not copied: %v", err)
	}
	if _, err := os.Stat(sourceDir); !os.IsNotExist(err) {
		t.Fatalf("expected source folder removed, stat err: %v", err)
	}
}

func TestMoveFileAcrossRootsMovesColdDescendantsAndRemovesColdBytes(t *testing.T) {
	fixture := newCrossRootFixture(t)
	sourceDir := filepath.Join(fixture.primaryRoot, "arquivo")
	if err := os.Mkdir(sourceDir, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	coldPath := filepath.Join(t.TempDir(), "cold-copy.bin")
	if err := os.WriteFile(coldPath, []byte("cold bytes"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	coldRecord := FileModel{ID: 5, Name: "old.bin", Path: filepath.Join(sourceDir, "old.bin"), Type: File}
	coldRecord.PhysicalPath = sql.NullString{String: coldPath, Valid: true}

	fixture.buildService(t, []FileModel{{ID: 1, Name: "arquivo", Path: sourceDir, Type: Directory}})
	fixture.service.Repository.(*filesRepoMock).getFilesByPathPrefixFn = func(prefix string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
		return utils.PaginationResponse[FileModel]{Items: []FileModel{coldRecord}}, nil
	}

	destinationFolderID := fixture.destinationID
	if _, err := fixture.service.MoveFile(1, &destinationFolderID, ""); err != nil {
		t.Fatalf("MoveFile returned error: %v", err)
	}

	movedBytes, err := os.ReadFile(filepath.Join(fixture.destinationDir, "arquivo", "old.bin"))
	if err != nil || string(movedBytes) != "cold bytes" {
		t.Fatalf("cold descendant not materialized: %q, %v", movedBytes, err)
	}
	if _, statErr := os.Stat(coldPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected cold bytes removed, stat err: %v", statErr)
	}
}

func TestMoveFileAcrossRootsFailureLeavesSourceIntactAndRemovesPartialDestination(t *testing.T) {
	fixture := newCrossRootFixture(t)
	sourceDir := filepath.Join(fixture.primaryRoot, "quebrada")
	if err := os.Mkdir(sourceDir, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("first"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.Symlink(filepath.Join(sourceDir, "missing-target"), filepath.Join(sourceDir, "z-dangling")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	fixture.buildService(t, []FileModel{{ID: 1, Name: "quebrada", Path: sourceDir, Type: Directory}})

	destinationFolderID := fixture.destinationID
	_, err := fixture.service.MoveFile(1, &destinationFolderID, "")
	requireOperationError(t, err, http.StatusInternalServerError, "ERROR_COPY_FAILED")

	if _, statErr := os.Stat(filepath.Join(sourceDir, "a.txt")); statErr != nil {
		t.Fatalf("source file lost after failed move: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(fixture.destinationDir, "quebrada")); !os.IsNotExist(statErr) {
		t.Fatalf("partial destination not removed, stat err: %v", statErr)
	}
	for _, deletedPath := range fixture.deletedPaths {
		if deletedPath == sourceDir {
			t.Fatalf("source rows must not be soft-deleted on failure")
		}
	}
}

func TestMoveFileAcrossRootsRefusesStorageRoot(t *testing.T) {
	fixture := newCrossRootFixture(t)
	fixture.buildService(t, []FileModel{{ID: 1, Name: filepath.Base(fixture.secondRoot), Path: fixture.secondRoot, Type: Directory}})
	primaryFolderPath := filepath.Join(fixture.primaryRoot, "destino")
	if err := os.Mkdir(primaryFolderPath, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	fixture.service.Repository.(*filesRepoMock).getFileByIDFn = func(id int) (FileModel, bool, error) {
		if id == 1 {
			return FileModel{ID: 1, Name: "filmes-root", Path: fixture.secondRoot, Type: Directory}, true, nil
		}
		return FileModel{ID: id, Name: "destino", Path: primaryFolderPath, Type: Directory}, true, nil
	}

	destinationFolderID := 2
	_, err := fixture.service.MoveFile(1, &destinationFolderID, "")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_MOVE_ROOT_FORBIDDEN")
	if _, statErr := os.Stat(fixture.secondRoot); statErr != nil {
		t.Fatalf("storage root vanished: %v", statErr)
	}
}
