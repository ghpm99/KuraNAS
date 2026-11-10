package files

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/utils"
)

type coldPurgeFixture struct {
	repo        *Repository
	service     *Service
	entryPoint  string
	coldVolume  string
	folderPath  string
	episodePath string
}

func newColdPurgeFixture(t *testing.T) *coldPurgeFixture {
	t.Helper()
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)
	folderPath := filepath.Join(entryPoint, "shows")
	return &coldPurgeFixture{
		repo:        repo,
		service:     &Service{Repository: repo, Tasks: make(chan utils.Task, 8)},
		entryPoint:  entryPoint,
		coldVolume:  t.TempDir(),
		folderPath:  folderPath,
		episodePath: filepath.Join(folderPath, "episode.mkv"),
	}
}

func (fixture *coldPurgeFixture) writeColdCopy(t *testing.T, name string) string {
	t.Helper()
	coldPath := filepath.Join(fixture.coldVolume, name)
	if err := os.WriteFile(coldPath, []byte("cold"), 0644); err != nil {
		t.Fatalf("write cold copy %q: %v", coldPath, err)
	}
	return coldPath
}

func (fixture *coldPurgeFixture) insertColdFile(t *testing.T, path string, coldPath string) {
	t.Helper()
	insertFileRow(t, fixture.repo, filepath.Base(path), path, filepath.Dir(path), 4, tieredRestoreSeedMoment)
	err := fixture.repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec("UPDATE home_file SET physical_path = $2 WHERE path = $1", path, coldPath)
		return execErr
	})
	if err != nil {
		t.Fatalf("mark %q cold: %v", path, err)
	}
}

func TestPostgres_PurgeDeletedSubtreeRemovesColdCopiesOfDeletedDescendants(t *testing.T) {
	fixture := newColdPurgeFixture(t)
	episodeColdPath := fixture.writeColdCopy(t, "episode.mkv")
	insertSearchRow(t, fixture.repo, "shows", fixture.folderPath, fixture.entryPoint, Directory)
	fixture.insertColdFile(t, fixture.episodePath, episodeColdPath)
	softDeleteRow(t, fixture.repo, fixture.folderPath)
	softDeleteRow(t, fixture.repo, fixture.episodePath)

	if err := fixture.service.PurgeDeletedSubtree(fixture.folderPath); err != nil {
		t.Fatalf("PurgeDeletedSubtree: %v", err)
	}

	if _, err := os.Stat(episodeColdPath); !os.IsNotExist(err) {
		t.Fatalf("cold copy of the purged descendant must be gone, stat err=%v", err)
	}
	if paths := remainingPaths(t, fixture.repo); len(paths) != 0 {
		t.Fatalf("purged rows must be gone, got %v", paths)
	}
}

func TestPostgres_PurgeDeletedSubtreeKeepsColdCopiesOfActiveFiles(t *testing.T) {
	fixture := newColdPurgeFixture(t)
	recreatedColdPath := fixture.writeColdCopy(t, "recreated.mkv")
	recreatedPath := filepath.Join(fixture.folderPath, "recreated.mkv")
	insertSearchRow(t, fixture.repo, "shows", fixture.folderPath, fixture.entryPoint, Directory)
	fixture.insertColdFile(t, recreatedPath, recreatedColdPath)
	softDeleteRow(t, fixture.repo, fixture.folderPath)

	if err := fixture.service.PurgeDeletedSubtree(fixture.folderPath); err != nil {
		t.Fatalf("PurgeDeletedSubtree: %v", err)
	}

	if _, err := os.Stat(recreatedColdPath); err != nil {
		t.Fatalf("cold copy of an active file must survive: %v", err)
	}
	if paths := remainingPaths(t, fixture.repo); len(paths) != 1 || paths[0] != recreatedPath {
		t.Fatalf("active row must survive, got %v", paths)
	}
}
