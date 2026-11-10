package trash_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/trash"
	"nas-go/api/internal/config"
	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type trashLifecycleFixture struct {
	database     *database.DbContext
	filesRepo    *files.Repository
	filesService *files.Service
	trashService *trash.Service
	folderPath   string
	songPath     string
	notePath     string
	folderID     int
	noteID       int
}

func newTrashLifecycleFixture(t *testing.T) *trashLifecycleFixture {
	t.Helper()
	dbContext := testutil.NewPostgresDB(t, "kuranas_trash_it")
	entryPoint := t.TempDir()
	previousEntryPoint := config.AppConfig.EntryPoint
	config.AppConfig.EntryPoint = entryPoint
	t.Cleanup(func() { config.AppConfig.EntryPoint = previousEntryPoint })

	filesRepo := files.NewRepository(dbContext)
	filesService := &files.Service{Repository: filesRepo, Tasks: make(chan utils.Task, 64)}
	trashService := trash.NewService(trash.NewRepository(dbContext), filesService)
	filesService.SetTrashBin(trashService)

	fixture := &trashLifecycleFixture{
		database:     dbContext,
		filesRepo:    filesRepo,
		filesService: filesService,
		trashService: trashService,
		folderPath:   filepath.Join(entryPoint, "album"),
		notePath:     filepath.Join(entryPoint, "note.txt"),
	}
	fixture.songPath = filepath.Join(fixture.folderPath, "song.txt")
	fixture.truncateTables(t)
	fixture.writeTree(t)
	fixture.folderID = fixture.insertRow(t, fixture.folderPath, files.Directory, 4096)
	fixture.insertRow(t, fixture.songPath, files.File, 7)
	fixture.noteID = fixture.insertRow(t, fixture.notePath, files.File, 5)
	return fixture
}

func (fixture *trashLifecycleFixture) truncateTables(t *testing.T) {
	t.Helper()
	err := fixture.database.ExecTx(func(tx *sql.Tx) error {
		_, truncateErr := tx.Exec("TRUNCATE home_file, trash_item RESTART IDENTITY CASCADE")
		return truncateErr
	})
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

func (fixture *trashLifecycleFixture) writeTree(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(fixture.folderPath, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	contentByPath := map[string]string{fixture.songPath: "musical", fixture.notePath: "notes"}
	for path, content := range contentByPath {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile %q: %v", path, err)
		}
	}
}

func (fixture *trashLifecycleFixture) insertRow(t *testing.T, path string, fileType files.FileType, size int64) int {
	t.Helper()
	var createdID int
	err := fixture.database.ExecTx(func(tx *sql.Tx) error {
		created, createErr := fixture.filesRepo.CreateFile(tx, files.FileModel{
			Name:       filepath.Base(path),
			Path:       path,
			ParentPath: filepath.Dir(path),
			Format:     filepath.Ext(path),
			Size:       size,
			UpdatedAt:  time.Now(),
			CreatedAt:  time.Now(),
			Type:       fileType,
		})
		createdID = created.ID
		return createErr
	})
	if err != nil {
		t.Fatalf("insert row %q: %v", path, err)
	}
	return createdID
}

func (fixture *trashLifecycleFixture) countRows(t *testing.T, condition string, arguments ...any) int {
	t.Helper()
	var rowCount int
	row := fixture.database.GetDatabase().QueryRow("SELECT COUNT(*) FROM home_file WHERE "+condition, arguments...)
	if err := row.Scan(&rowCount); err != nil {
		t.Fatalf("count rows (%s): %v", condition, err)
	}
	return rowCount
}

func (fixture *trashLifecycleFixture) onlyTrashItem(t *testing.T) trash.TrashItemDto {
	t.Helper()
	listing, err := fixture.trashService.GetItems(trash.DefaultListing, 1, 10)
	if err != nil || len(listing.Items) != 1 {
		t.Fatalf("expected exactly one trash item, got %+v err=%v", listing.Items, err)
	}
	return listing.Items[0]
}

func TestPostgres_DeletedFileIsListedAndComesBackWithTheSameRow(t *testing.T) {
	fixture := newTrashLifecycleFixture(t)

	if err := fixture.filesService.DeleteFileFromDisk(fixture.noteID, false); err != nil {
		t.Fatalf("DeleteFileFromDisk: %v", err)
	}
	item := fixture.onlyTrashItem(t)
	if item.Name != "note.txt" || item.IsDirectory || item.Size != 5 {
		t.Fatalf("trash item described wrongly: %+v", item)
	}
	if fixture.countRows(t, "id = $1 AND deleted_at IS NOT NULL", fixture.noteID) != 1 {
		t.Fatalf("deleted file must be soft-deleted in the index")
	}

	if _, err := fixture.trashService.RestoreItem(item.ID); err != nil {
		t.Fatalf("RestoreItem: %v", err)
	}
	if content, err := os.ReadFile(fixture.notePath); err != nil || string(content) != "notes" {
		t.Fatalf("restored bytes wrong: %q err=%v", content, err)
	}
	if fixture.countRows(t, "id = $1 AND path = $2 AND deleted_at IS NULL", fixture.noteID, fixture.notePath) != 1 {
		t.Fatalf("restored file must be active again with the same id")
	}
}

func TestPostgres_PurgedFolderLeavesNoBytesAndNoRows(t *testing.T) {
	fixture := newTrashLifecycleFixture(t)

	if err := fixture.filesService.DeleteFileFromDisk(fixture.folderID, false); err != nil {
		t.Fatalf("DeleteFileFromDisk: %v", err)
	}
	item := fixture.onlyTrashItem(t)
	if !item.IsDirectory || item.Size != 7 {
		t.Fatalf("folder item must be a directory sized by its content, got %+v", item)
	}

	if err := fixture.trashService.DeleteItemPermanently(item.ID); err != nil {
		t.Fatalf("DeleteItemPermanently: %v", err)
	}
	trashEntries, err := os.ReadDir(filepath.Join(filepath.Dir(fixture.folderPath), trash.DirName))
	if err != nil || len(trashEntries) != 0 {
		t.Fatalf("trash directory must be empty, got %v err=%v", trashEntries, err)
	}
	if fixture.countRows(t, "path = $1 OR path = $2", fixture.folderPath, fixture.songPath) != 0 {
		t.Fatalf("purged folder and its content must leave the index")
	}
	if fixture.countRows(t, "id = $1 AND deleted_at IS NULL", fixture.noteID) != 1 {
		t.Fatalf("files outside the purged folder must stay untouched")
	}
}

func TestPostgres_ExpiredItemIsPurgedByRetention(t *testing.T) {
	fixture := newTrashLifecycleFixture(t)
	if err := fixture.filesService.DeleteFileFromDisk(fixture.noteID, false); err != nil {
		t.Fatalf("DeleteFileFromDisk: %v", err)
	}
	agingErr := fixture.database.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec("UPDATE trash_item SET deleted_at = now() - interval '31 days'")
		return execErr
	})
	if agingErr != nil {
		t.Fatalf("age trash item: %v", agingErr)
	}

	outcome, err := fixture.trashService.PurgeExpired()

	if err != nil || outcome != (trash.PurgeOutcome{Purged: 1}) {
		t.Fatalf("expected the aged item purged, got %+v err=%v", outcome, err)
	}
	if fixture.countRows(t, "id = $1", fixture.noteID) != 0 {
		t.Fatalf("expired item must leave the index")
	}
}
