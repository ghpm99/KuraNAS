package video

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/api/v1/video/playlist"
)

func TestApplyClassificationStampsClassificationAndCurrentVersion(t *testing.T) {
	metadata := VideoMetadataModel{Duration: "7200.000000", Height: 1080}
	metadata.ApplyClassification("film.mkv", "/x/film.mkv", "/x")

	if metadata.Classification != string(playlist.ClassMovie) {
		t.Fatalf("expected movie from duration and resolution, got %q", metadata.Classification)
	}
	if metadata.ClassificationVersion != CurrentVideoClassificationVersion {
		t.Fatalf("expected version %d, got %d", CurrentVideoClassificationVersion, metadata.ClassificationVersion)
	}
}

func TestCatalogAndRebuildAgreeOnPersistedClassification(t *testing.T) {
	persistedMovie := VideoFileModel{
		ID: 1, Name: "Show S01E01.mkv", ParentPath: "/films/feature", Path: "/films/feature/Show S01E01.mkv",
		PersistedClassification: string(playlist.ClassMovie),
	}
	secondPersistedMovie := VideoFileModel{
		ID: 3, Name: "Show S01E03.mkv", ParentPath: "/films/feature", Path: "/films/feature/Show S01E03.mkv",
		PersistedClassification: string(playlist.ClassMovie),
	}
	unclassifiedSeries := VideoFileModel{
		ID: 2, Name: "Show S01E02.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E02.mkv",
	}

	var rebuiltClassifications []string
	repo := &videoRepoMock{
		getCatalogVideosFn: func(limit int) ([]VideoFileModel, error) {
			return []VideoFileModel{persistedMovie, unclassifiedSeries}, nil
		},
		getRecentVideosFn: func(limit int) ([]VideoFileModel, error) { return nil, nil },
		getContinueWatchingFn: func(clientID string, limit int) ([]ContinueWatchingModel, error) {
			return nil, nil
		},
		getAllVideosWithMetadataFn: func() ([]VideoWithMetadataModel, error) {
			return []VideoWithMetadataModel{
				{
					VideoFileModel:     persistedMovie,
					MetaClassification: sql.NullString{String: persistedMovie.PersistedClassification, Valid: true},
				},
				{
					VideoFileModel:     secondPersistedMovie,
					MetaClassification: sql.NullString{String: secondPersistedMovie.PersistedClassification, Valid: true},
				},
				{VideoFileModel: unclassifiedSeries},
			}, nil
		},
		upsertAutoPlaylistFn: func(tx *sql.Tx, contextType, sourcePath, name, groupMode, classification string) (VideoPlaylistModel, error) {
			rebuiltClassifications = append(rebuiltClassifications, classification)
			return VideoPlaylistModel{ID: 100, Name: name}, nil
		},
		getPlaylistExclusionsFn:   func(playlistID int) (map[int]bool, error) { return map[int]bool{}, nil },
		deleteAutoPlaylistItemsFn: func(tx *sql.Tx, playlistID int) error { return nil },
		insertPlaylistItemsSrcFn:  func(tx *sql.Tx, playlistID int, videoIDs []int, sourceKind string) error { return nil },
	}
	svc := newVideoServiceForTest(t, repo)

	catalog, err := svc.GetHomeCatalog("client", 10)
	if err != nil {
		t.Fatalf("GetHomeCatalog failed: %v", err)
	}
	sectionVideoIDs := map[string][]int{}
	for _, section := range catalog.Sections {
		for _, item := range section.Items {
			sectionVideoIDs[section.Key] = append(sectionVideoIDs[section.Key], item.Video.ID)
		}
	}
	if len(sectionVideoIDs["movies"]) != 1 || sectionVideoIDs["movies"][0] != 1 {
		t.Fatalf("catalog must use the persisted classification, got %+v", sectionVideoIDs)
	}
	if len(sectionVideoIDs["series"]) != 1 || sectionVideoIDs["series"][0] != 2 {
		t.Fatalf("catalog must fall back to computing when not persisted, got %+v", sectionVideoIDs)
	}

	catalogEntry := videoModelToEntry(persistedMovie)
	if svc.PlaylistEngine.Classifier.Classify(catalogEntry).Classification != playlist.ClassMovie {
		t.Fatal("classifier must honor the persisted classification")
	}

	if err := svc.RebuildSmartPlaylists(); err != nil {
		t.Fatalf("RebuildSmartPlaylists failed: %v", err)
	}
	hasMovieGroup := false
	for _, classification := range rebuiltClassifications {
		if classification == string(playlist.ClassMovie) {
			hasMovieGroup = true
		}
	}
	if !hasMovieGroup {
		t.Fatalf("rebuild must group using the persisted classification, got %v", rebuiltClassifications)
	}
}
