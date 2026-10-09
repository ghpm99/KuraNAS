package music

import (
	"database/sql"
	"errors"
	"testing"

	"nas-go/api/internal/api/v1/files"
)

func TestServiceGetLibraryTrackByIDReturnsTrackWithMetadata(t *testing.T) {
	var requestedIDs []int
	repository := &musicRepoMock{
		getLibraryFilesByIDsFn: func(fileIDs []int) ([]files.FileModel, error) {
			requestedIDs = fileIDs
			return []files.FileModel{{ID: 7, Name: "song.mp3", Metadata: AudioMetadataModel{Title: "Song", Artist: "Band"}}}, nil
		},
	}
	service := &Service{Repository: repository}

	track, err := service.GetLibraryTrackByID(7)
	if err != nil {
		t.Fatalf("GetLibraryTrackByID: %v", err)
	}
	if len(requestedIDs) != 1 || requestedIDs[0] != 7 {
		t.Fatalf("unexpected requested ids %v", requestedIDs)
	}
	if track.ID != 7 {
		t.Fatalf("unexpected track %+v", track)
	}
	metadata, isAudioMetadata := track.Metadata.(AudioMetadataModel)
	if !isAudioMetadata || metadata.Title != "Song" || metadata.Artist != "Band" {
		t.Fatalf("expected audio metadata on dto, got %+v", track.Metadata)
	}
}

func TestServiceGetLibraryTrackByIDReturnsNoRowsWhenNotAudioOrDeleted(t *testing.T) {
	repository := &musicRepoMock{
		getLibraryFilesByIDsFn: func(fileIDs []int) ([]files.FileModel, error) {
			return []files.FileModel{}, nil
		},
	}
	service := &Service{Repository: repository}

	if _, err := service.GetLibraryTrackByID(7); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestServiceGetLibraryTrackByIDPropagatesRepositoryError(t *testing.T) {
	repositoryError := errors.New("db down")
	repository := &musicRepoMock{
		getLibraryFilesByIDsFn: func(fileIDs []int) ([]files.FileModel, error) {
			return nil, repositoryError
		},
	}
	service := &Service{Repository: repository}

	if _, err := service.GetLibraryTrackByID(7); !errors.Is(err, repositoryError) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
