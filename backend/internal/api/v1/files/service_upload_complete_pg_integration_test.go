package files

import (
	"net/http"
	"path/filepath"
	"testing"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/utils"
)

func TestPostgres_CompleteUploadAppearsInListingWithoutWorkers(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	truncateHomeFile(t, repo)
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)
	setCapturesPathForTest(t, t.TempDir())
	service := &Service{
		Repository:       repo,
		Tasks:            make(chan utils.Task, 32),
		JobsRepository:   newFilesJobsRepoMockForTest(t),
		AvailableBytesAt: func(string) (int64, error) { return 1 << 40, nil },
	}
	router := newCompleteRouter(service)
	uploadID := uploadAllChunks(t, router, initBody("song.mp3", "notes", `,"relative_path":"albums"`), "notes")

	recorder := postComplete(router, completeBody(uploadID))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	response := decodeCompleteResponse(t, recorder)
	albumsPath := filepath.Join(entryPoint, "albums")
	if response.JobID <= 0 || response.File.Path != filepath.Join(albumsPath, "song.mp3") {
		t.Fatalf("unexpected response: %+v", response)
	}
	if _, found := activeRowByPath(t, repo, albumsPath); !found {
		t.Fatalf("created folder missing from index")
	}
	fileRow, found := activeRowByPath(t, repo, response.File.Path)
	if !found || fileRow.ParentPath != albumsPath || fileRow.Size != 5 {
		t.Fatalf("published file missing from listing: found=%v row=%+v", found, fileRow)
	}
}
