package video

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
)

func TestCreateCustomPlaylistRejectsBlankAndOversizedNames(t *testing.T) {
	service := newVideoServiceForTest(t, &videoRepoMock{})

	if _, err := service.CreateCustomPlaylist(CreateCustomPlaylistRequest{Name: "   "}); !errors.Is(err, ErrPlaylistNameRequired) {
		t.Fatalf("blank name error = %v", err)
	}
	oversizedName := strings.Repeat("a", maxCustomPlaylistNameLength+1)
	if _, err := service.CreateCustomPlaylist(CreateCustomPlaylistRequest{Name: oversizedName}); !errors.Is(err, ErrPlaylistNameTooLong) {
		t.Fatalf("oversized name error = %v", err)
	}
}

func TestCreateCustomPlaylistTrimsNameAndGeneratesUniqueSourcePath(t *testing.T) {
	var receivedPaths []string
	var receivedName string
	repository := &videoRepoMock{createCustomPlaylistFn: func(tx *sql.Tx, sourcePath string, name string) (VideoPlaylistModel, error) {
		receivedPaths = append(receivedPaths, sourcePath)
		receivedName = name
		return VideoPlaylistModel{ID: 5, Type: "custom", SourcePath: sourcePath, Name: name}, nil
	}}
	service := newVideoServiceForTest(t, repository)

	for range 2 {
		if _, err := service.CreateCustomPlaylist(CreateCustomPlaylistRequest{Name: "  Favoritos  "}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if receivedName != "Favoritos" {
		t.Fatalf("name = %q", receivedName)
	}
	if receivedPaths[0] == receivedPaths[1] || !strings.HasPrefix(receivedPaths[0], "custom/") {
		t.Fatalf("source paths = %v", receivedPaths)
	}
}

func TestCreateCustomPlaylistWithVideoAddsItAndReportsItemCount(t *testing.T) {
	var addedPlaylistID, addedVideoID int
	repository := &videoRepoMock{
		getVideoFileByIDFn: func(id int) (VideoFileModel, error) { return VideoFileModel{ID: id}, nil },
		createCustomPlaylistFn: func(tx *sql.Tx, sourcePath string, name string) (VideoPlaylistModel, error) {
			return VideoPlaylistModel{ID: 5, Name: name}, nil
		},
		addPlaylistVideoManualFn: func(tx *sql.Tx, playlistID int, videoID int) error {
			addedPlaylistID, addedVideoID = playlistID, videoID
			return nil
		},
	}
	videoID := 42
	created, err := newVideoServiceForTest(t, repository).CreateCustomPlaylist(CreateCustomPlaylistRequest{Name: "x", VideoID: &videoID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addedPlaylistID != 5 || addedVideoID != 42 || created.ItemCount != 1 {
		t.Fatalf("added playlist=%d video=%d itemCount=%d", addedPlaylistID, addedVideoID, created.ItemCount)
	}
}

func TestCreateCustomPlaylistWithMissingVideoDoesNotCreate(t *testing.T) {
	wasCreated := false
	repository := &videoRepoMock{
		getVideoFileByIDFn: func(id int) (VideoFileModel, error) { return VideoFileModel{}, sql.ErrNoRows },
		createCustomPlaylistFn: func(tx *sql.Tx, sourcePath string, name string) (VideoPlaylistModel, error) {
			wasCreated = true
			return VideoPlaylistModel{}, nil
		},
	}
	videoID := 1
	_, err := newVideoServiceForTest(t, repository).CreateCustomPlaylist(CreateCustomPlaylistRequest{Name: "x", VideoID: &videoID})
	if !errors.Is(err, sql.ErrNoRows) || wasCreated {
		t.Fatalf("err=%v wasCreated=%v", err, wasCreated)
	}
}

func TestCreateCustomPlaylistPropagatesRepositoryFailure(t *testing.T) {
	repository := &videoRepoMock{createCustomPlaylistFn: func(*sql.Tx, string, string) (VideoPlaylistModel, error) {
		return VideoPlaylistModel{}, errors.New("insert failed")
	}}
	if _, err := newVideoServiceForTest(t, repository).CreateCustomPlaylist(CreateCustomPlaylistRequest{Name: "x"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteCustomPlaylistRefusesAutomaticPlaylist(t *testing.T) {
	wasDeleted := false
	repository := &videoRepoMock{
		getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) { return VideoPlaylistModel{ID: id, IsAuto: true}, nil },
		deleteCustomPlaylistFn: func(*sql.Tx, int) error { wasDeleted = true; return nil },
	}
	if err := newVideoServiceForTest(t, repository).DeleteCustomPlaylist(3); !errors.Is(err, ErrPlaylistAutomaticNotDeletable) || wasDeleted {
		t.Fatalf("err=%v wasDeleted=%v", err, wasDeleted)
	}
}

func TestDeleteCustomPlaylistReportsMissingPlaylist(t *testing.T) {
	repository := &videoRepoMock{getVideoPlaylistByIDFn: func(int) (VideoPlaylistModel, error) { return VideoPlaylistModel{}, sql.ErrNoRows }}
	if err := newVideoServiceForTest(t, repository).DeleteCustomPlaylist(3); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v", err)
	}
}

func TestDeleteCustomPlaylistDeletesCustomPlaylist(t *testing.T) {
	deletedID := 0
	repository := &videoRepoMock{
		getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) { return VideoPlaylistModel{ID: id}, nil },
		deleteCustomPlaylistFn: func(tx *sql.Tx, id int) error { deletedID = id; return nil },
	}
	if err := newVideoServiceForTest(t, repository).DeleteCustomPlaylist(3); err != nil || deletedID != 3 {
		t.Fatalf("err=%v deletedID=%d", err, deletedID)
	}
}
