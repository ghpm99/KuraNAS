package captures

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"
)

func newPostgresPromoteFixture(t *testing.T) (*Service, *Repository, string, string) {
	t.Helper()
	dbContext := testutil.NewPostgresDB(t, "kuranas_captures_it")
	repository := NewRepository(dbContext)

	rootDir := t.TempDir()
	videosDir := filepath.Join(rootDir, "videos")
	stagingDir := filepath.Join(rootDir, "capturas", "show")
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		t.Fatal(err)
	}
	receivedPath := filepath.Join(stagingDir, "recording.mp4")
	if err := os.WriteFile(receivedPath, []byte("received-bytes"), 0644); err != nil {
		t.Fatal(err)
	}

	librariesProvider := &librariesProviderMock{getByCategoryFn: func(libraries.LibraryCategory) (libraries.LibraryDto, error) {
		return libraries.LibraryDto{Path: videosDir}, nil
	}}
	service := &Service{
		Repository:        repository,
		LibrariesProvider: librariesProvider,
		FilesProvider:     newHomeFileRecordingProvider(dbContext),
	}
	return service, repository, receivedPath, videosDir
}

func newHomeFileRecordingProvider(dbContext *database.DbContext) *filesProviderMock {
	return &filesProviderMock{
		createFileFn: func(fileDto files.FileDto) (files.FileDto, error) {
			err := dbContext.ExecTx(func(tx *sql.Tx) error {
				return tx.QueryRow(
					`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at)
					 VALUES ($1, $2, $3, $4, $5, now(), now()) RETURNING id`,
					fileDto.Name, fileDto.Path, fileDto.ParentPath, fileDto.Format, fileDto.Size,
				).Scan(&fileDto.ID)
			})
			return fileDto, err
		},
		deleteFn: func(id int) error {
			return dbContext.ExecTx(func(tx *sql.Tx) error {
				_, err := tx.Exec(`DELETE FROM home_file WHERE id = $1`, id)
				return err
			})
		},
	}
}

func insertReceivedCapture(t *testing.T, repository *Repository, receivedPath string) CaptureModel {
	t.Helper()
	var created CaptureModel
	err := repository.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		var createErr error
		created, createErr = repository.CreateCapture(tx, CaptureModel{
			Name:        "show",
			FileName:    "recording.mp4",
			FilePath:    receivedPath,
			MediaType:   "recording",
			MimeType:    "video/mp4",
			Size:        14,
			CreatedAt:   time.Now(),
			RawMetadata: json.RawMessage(`{"title":"Show","season":1,"episode":1,"episode_title":"Pilot"}`),
		})
		return createErr
	})
	if err != nil {
		t.Fatalf("insert capture: %v", err)
	}
	t.Cleanup(func() {
		_ = repository.GetDbContext().ExecTx(func(tx *sql.Tx) error {
			_ = repository.DeleteCapture(tx, created.ID)
			return nil
		})
	})
	return created
}

func countFilesUnder(t *testing.T, dir string) int {
	t.Helper()
	count := 0
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}

func TestPromoteCaptureRemuxFailureEndsFailedAndKeepsReceivedFile_Postgres(t *testing.T) {
	service, repository, receivedPath, videosDir := newPostgresPromoteFixture(t)
	created := insertReceivedCapture(t, repository, receivedPath)

	remuxFails := true
	swapRemux(t, func(src, dest string) error {
		if remuxFails {
			_ = os.WriteFile(dest, []byte("partial"), 0644)
			return os.ErrInvalid
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})

	if err := service.PromoteCapture(created.ID); err == nil {
		t.Fatal("expected promotion error when remux fails")
	}

	failed, err := repository.GetCaptureByID(created.ID)
	if err != nil {
		t.Fatalf("reload capture: %v", err)
	}
	if failed.Status != CaptureStatusFailed {
		t.Fatalf("expected failed status, got %s", failed.Status)
	}
	if failed.FailureReason != i18n.GetMessage("CAPTURE_PROMOTION_FAILED_REMUX") || failed.FailureReason == "" {
		t.Fatalf("unexpected failure reason %q", failed.FailureReason)
	}
	if failed.FileID != nil {
		t.Fatalf("expected file_id detached, got %v", *failed.FileID)
	}
	if failed.FilePath != receivedPath || failed.FileName != "recording.mp4" {
		t.Fatalf("expected capture to point at received file, got %q / %q", failed.FilePath, failed.FileName)
	}
	if _, statErr := os.Stat(receivedPath); statErr != nil {
		t.Fatalf("expected received file kept: %v", statErr)
	}
	if count := countFilesUnder(t, videosDir); count != 0 {
		t.Fatalf("expected nothing in the videos library, found %d files", count)
	}

	remuxFails = false
	if err := service.PromoteCapture(created.ID); err != nil {
		t.Fatalf("retry after fixing the cause failed: %v", err)
	}
	retried, err := repository.GetCaptureByID(created.ID)
	if err != nil {
		t.Fatalf("reload capture: %v", err)
	}
	if retried.Status != CaptureStatusPromoted || retried.FailureReason != "" {
		t.Fatalf("expected promoted without reason, got %s %q", retried.Status, retried.FailureReason)
	}
	if count := countFilesUnder(t, videosDir); count != 1 {
		t.Fatalf("expected one file in the videos library, found %d", count)
	}
}

func TestDeleteFailedCaptureRemovesReceivedFile_Postgres(t *testing.T) {
	service, repository, receivedPath, videosDir := newPostgresPromoteFixture(t)
	created := insertReceivedCapture(t, repository, receivedPath)

	swapRemux(t, func(src, dest string) error { return os.ErrInvalid })
	if err := service.PromoteCapture(created.ID); err == nil {
		t.Fatal("expected promotion error when remux fails")
	}

	libraryFileKept, err := service.DeleteCapture(created.ID)
	if err != nil {
		t.Fatalf("DeleteCapture: %v", err)
	}
	if libraryFileKept {
		t.Fatal("expected no library file kept for a failed capture")
	}
	if _, statErr := os.Stat(receivedPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected received file removed, stat err = %v", statErr)
	}
	if count := countFilesUnder(t, videosDir); count != 0 {
		t.Fatalf("expected nothing in the videos library, found %d files", count)
	}
	if _, err := repository.GetCaptureByID(created.ID); err == nil {
		t.Fatal("expected capture row gone")
	}
}
