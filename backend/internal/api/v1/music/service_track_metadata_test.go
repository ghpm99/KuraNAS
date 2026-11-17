package music

import (
	"database/sql"
	"errors"
	"testing"

	"nas-go/api/internal/api/v1/files"
)

type trackMetadataStoreMock struct {
	applyError     error
	appliedFileIDs []int
	appliedPatch   TrackMetadataPatch
	appliedLabel   string
	applyCallCount int

	indexedFileIDs      []int
	albumFileIDs        []int
	folderFileIDs       []int
	selectionError      error
	filteredFileIDs     []int
	requestedAlbumKey   string
	requestedFolderPath string

	genreSources           []GenreTrackSource
	trackCountByGenreKey   map[string]int
	appliedGroups          []TrackMetadataPatchGroup
	applyGroupsCallCount   int
	requestedGenreSourceOf string
}

func (mock *trackMetadataStoreMock) ApplyTrackMetadataPatchGroups(tx *sql.Tx, groups []TrackMetadataPatchGroup, compilationArtistLabel string) error {
	mock.applyGroupsCallCount++
	mock.appliedGroups = groups
	return mock.applyError
}

func (mock *trackMetadataStoreMock) ListGenreTrackSources(genreKey string) ([]GenreTrackSource, error) {
	mock.requestedGenreSourceOf = genreKey
	return mock.genreSources, mock.selectionError
}

func (mock *trackMetadataStoreMock) CountActiveTracksByGenreKey(genreKey string) (int, error) {
	return mock.trackCountByGenreKey[genreKey], mock.selectionError
}

func (mock *trackMetadataStoreMock) FilterIndexedTrackFileIDs(fileIDs []int) ([]int, error) {
	mock.filteredFileIDs = fileIDs
	return mock.indexedFileIDs, mock.selectionError
}

func (mock *trackMetadataStoreMock) ListTrackFileIDsByAlbumKey(albumKey string) ([]int, error) {
	mock.requestedAlbumKey = albumKey
	return mock.albumFileIDs, mock.selectionError
}

func (mock *trackMetadataStoreMock) ListTrackFileIDsByFolder(folderPath string) ([]int, error) {
	mock.requestedFolderPath = folderPath
	return mock.folderFileIDs, mock.selectionError
}

func (mock *trackMetadataStoreMock) ApplyTrackMetadataPatch(tx *sql.Tx, fileIDs []int, patch TrackMetadataPatch, compilationArtistLabel string) error {
	mock.applyCallCount++
	mock.appliedFileIDs = fileIDs
	mock.appliedPatch = patch
	mock.appliedLabel = compilationArtistLabel
	return mock.applyError
}

func serviceWithIndexedTrack(store *trackMetadataStoreMock, trackModels []files.FileModel) *Service {
	repository := &musicRepoMock{
		getLibraryFilesByIDsFn: func(fileIDs []int) ([]files.FileModel, error) { return trackModels, nil },
	}
	return &Service{Repository: repository, TrackMetadataStore: store}
}

func indexedTrackModels() []files.FileModel {
	return []files.FileModel{{ID: 7, Name: "song.mp3", Metadata: AudioMetadataModel{ID: 3, Title: "Song"}}}
}

func TestServiceUpdateTrackMetadataAppliesPatchAndReturnsTrack(t *testing.T) {
	store := &trackMetadataStoreMock{}
	service := serviceWithIndexedTrack(store, indexedTrackModels())

	track, err := service.UpdateTrackMetadata(7, decodePatch(t, `{"title": "Fixed"}`))
	if err != nil {
		t.Fatalf("UpdateTrackMetadata: %v", err)
	}

	if store.applyCallCount != 1 || len(store.appliedFileIDs) != 1 || store.appliedFileIDs[0] != 7 {
		t.Fatalf("apply calls=%d ids=%v", store.applyCallCount, store.appliedFileIDs)
	}
	if !store.appliedPatch.Title.IsSet || *store.appliedPatch.Title.Value != "Fixed" {
		t.Fatalf("applied patch = %+v", store.appliedPatch)
	}
	if store.appliedLabel == "" {
		t.Fatal("compilation artist label must be passed")
	}
	if track.ID != 7 {
		t.Fatalf("unexpected track %+v", track)
	}
}

func TestServiceUpdateTrackMetadataRefusesInvalidPatchBeforeTouchingStorage(t *testing.T) {
	store := &trackMetadataStoreMock{}
	service := serviceWithIndexedTrack(store, indexedTrackModels())

	_, err := service.UpdateTrackMetadata(7, decodePatch(t, `{"year": 12}`))

	var validationError *fieldValidationError
	if !errors.As(err, &validationError) || validationError.Field != "year" {
		t.Fatalf("expected year validation error, got %v", err)
	}
	if store.applyCallCount != 0 {
		t.Fatal("storage must not be called for an invalid patch")
	}
}

func TestServiceUpdateTrackMetadataReturnsNoRowsWhenTrackIsMissing(t *testing.T) {
	store := &trackMetadataStoreMock{}
	service := serviceWithIndexedTrack(store, []files.FileModel{})

	if _, err := service.UpdateTrackMetadata(7, decodePatch(t, `{}`)); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if store.applyCallCount != 0 {
		t.Fatal("storage must not be called for a missing track")
	}
}

func TestServiceUpdateTrackMetadataReturnsNoRowsWhenTrackHasNoAudioMetadata(t *testing.T) {
	store := &trackMetadataStoreMock{}
	service := serviceWithIndexedTrack(store, []files.FileModel{{ID: 7, Name: "song.mp3", Metadata: AudioMetadataModel{}}})

	if _, err := service.UpdateTrackMetadata(7, decodePatch(t, `{}`)); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if store.applyCallCount != 0 {
		t.Fatal("storage must not be called for a track without audio metadata")
	}
}

func TestServiceUpdateTrackMetadataPropagatesStorageError(t *testing.T) {
	storageError := errors.New("db down")
	service := serviceWithIndexedTrack(&trackMetadataStoreMock{applyError: storageError}, indexedTrackModels())

	if _, err := service.UpdateTrackMetadata(7, decodePatch(t, `{}`)); !errors.Is(err, storageError) {
		t.Fatalf("expected storage error, got %v", err)
	}
}
