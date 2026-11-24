package video

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"nas-go/api/internal/testutil"
)

var sagaFileNames = buildSagaFileNames()

const (
	sagaFolder                        = "/data/Saga"
	unrelatedClassificationVideoCount = 60
)

type classificationFixture struct {
	repository         *Repository
	metadataRepository *VideoMetadataRepository
	service            *Service
	videoIDs           map[string]int
}

func buildSagaFileNames() []string {
	titleWords := []string{
		"Alpha", "Bravo", "Charlie", "Delta", "Echo", "Foxtrot", "Golf", "Hotel", "India", "Juliet", "Kilo", "Lima",
		"Mike", "November", "Oscar", "Papa", "Quebec", "Romeo", "Sierra", "Tango", "Uniform", "Victor", "Whiskey", "Yankee",
	}
	fileNames := make([]string, 0, len(titleWords))
	for _, titleWord := range titleWords {
		fileNames = append(fileNames, titleWord+" Film.mkv")
	}
	return fileNames
}

func newClassificationFixture(t *testing.T) *classificationFixture {
	t.Helper()
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository).(*Service)

	seeds := []seededVideoFile{}
	for _, name := range sagaFileNames {
		seeds = append(seeds, seededVideoFile{name: name, folder: sagaFolder})
	}
	for unrelatedNumber := 1; unrelatedNumber <= unrelatedClassificationVideoCount; unrelatedNumber++ {
		seeds = append(seeds, seededVideoFile{name: fmt.Sprintf("Home Video %d.mp4", unrelatedNumber), folder: fmt.Sprintf("/data/Other%d", unrelatedNumber%3)})
	}
	videoIDs := seedVideoFiles(t, repository, seeds)

	fixture := &classificationFixture{
		repository:         repository,
		metadataRepository: NewVideoMetadataRepository(dbContext),
		service:            service,
		videoIDs:           videoIDs,
	}
	for _, name := range sagaFileNames {
		fixture.upsertMetadata(t, name, "movie", CurrentVideoClassificationVersion)
	}
	if err := service.RebuildSmartPlaylists(); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	return fixture
}

func (fixture *classificationFixture) upsertMetadata(t *testing.T, fileName string, classification string, classificationVersion int) VideoMetadataModel {
	t.Helper()
	var stored VideoMetadataModel
	err := fixture.repository.DbContext.ExecTx(func(tx *sql.Tx) error {
		var upsertErr error
		stored, upsertErr = fixture.metadataRepository.UpsertVideoMetadata(tx, VideoMetadataModel{
			FileId:                fixture.videoIDs[fileName],
			Path:                  sagaFolder + "/" + fileName,
			Duration:              "7200.0",
			Classification:        classification,
			ClassificationVersion: classificationVersion,
		})
		return upsertErr
	})
	if err != nil {
		t.Fatalf("upsert metadata %s: %v", fileName, err)
	}
	return stored
}

func (fixture *classificationFixture) movieNames(t *testing.T) []string {
	t.Helper()
	page, err := fixture.service.ListLibraryMovies(LibraryMoviesRequest{Sort: LibraryMovieSortName, Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("list movies: %v", err)
	}
	names := []string{}
	for _, item := range page.Items {
		names = append(names, item.Name)
	}
	return names
}

func (fixture *classificationFixture) folderPlaylistClassification(t *testing.T) string {
	t.Helper()
	var classification string
	err := fixture.repository.DbContext.GetDatabase().QueryRow(
		`SELECT vp.classification FROM video_playlist vp
		 JOIN video_playlist_item vpi ON vpi.playlist_id = vp.id
		 WHERE vp.is_auto AND vp.group_mode = 'folder' AND vpi.video_id = $1
		 ORDER BY vp.id LIMIT 1`, fixture.videoIDs[sagaFileNames[0]]).Scan(&classification)
	if err != nil {
		t.Fatalf("folder playlist: %v", err)
	}
	return classification
}

func (fixture *classificationFixture) setClassification(t *testing.T, fileName string, classification *string) VideoClassificationDto {
	t.Helper()
	response, err := fixture.service.SetVideoClassification(fixture.videoIDs[fileName], classification)
	if err != nil {
		t.Fatalf("set classification: %v", err)
	}
	fixture.service.pendingRebuilds.Wait()
	return response
}

func (fixture *classificationFixture) forceOverridesExceptLast(t *testing.T, classification string) {
	t.Helper()
	database := fixture.repository.DbContext.GetDatabase()
	for _, fileName := range sagaFileNames[:len(sagaFileNames)-1] {
		var err error
		if classification == "" {
			_, err = database.Exec(`DELETE FROM video_manual_override WHERE file_id = $1`, fixture.videoIDs[fileName])
		} else {
			_, err = database.Exec(`INSERT INTO video_manual_override (file_id, classification) VALUES ($1, $2)
				ON CONFLICT (file_id) DO UPDATE SET classification = EXCLUDED.classification`, fixture.videoIDs[fileName], classification)
		}
		if err != nil {
			t.Fatalf("force override %s: %v", fileName, err)
		}
	}
}

func stringPointer(text string) *string { return &text }

func TestClassificationOverrideMovesVideoBetweenMoviesAndSeries_Postgres(t *testing.T) {
	fixture := newClassificationFixture(t)
	if got := fixture.movieNames(t); !reflect.DeepEqual(got, sagaFileNames) {
		t.Fatalf("movies before override = %v", got)
	}
	if got := fixture.folderPlaylistClassification(t); got != "movie" {
		t.Fatalf("folder playlist classification = %q, want movie", got)
	}

	response := fixture.setClassification(t, sagaFileNames[0], stringPointer("series"))
	if response.Classification != "series" || !response.IsOverride || response.FileID != fixture.videoIDs[sagaFileNames[0]] {
		t.Fatalf("response = %+v", response)
	}
	if got := fixture.movieNames(t); !reflect.DeepEqual(got, sagaFileNames[1:]) {
		t.Fatalf("movies right after override = %v, want %v", got, sagaFileNames[1:])
	}

	fixture.forceOverridesExceptLast(t, "series")
	fixture.setClassification(t, sagaFileNames[len(sagaFileNames)-1], stringPointer("series"))
	if got := fixture.movieNames(t); len(got) != 0 {
		t.Fatalf("movies after overriding all = %v, want none", got)
	}
	if got := fixture.folderPlaylistClassification(t); got != "series" {
		t.Fatalf("folder playlist classification after rebuild = %q, want series", got)
	}

	fixture.forceOverridesExceptLast(t, "")
	response = fixture.setClassification(t, sagaFileNames[len(sagaFileNames)-1], nil)
	if response.Classification != "movie" || response.IsOverride {
		t.Fatalf("response after clearing = %+v", response)
	}
	if got := fixture.movieNames(t); !reflect.DeepEqual(got, sagaFileNames) {
		t.Fatalf("movies after clearing = %v", got)
	}
	if got := fixture.folderPlaylistClassification(t); got != "movie" {
		t.Fatalf("folder playlist classification after clearing = %q, want movie", got)
	}
	assertOverrideRowCount(t, fixture.repository, 0)
}

func TestClassificationOtherIsStoredAsProgramForSections_Postgres(t *testing.T) {
	fixture := newClassificationFixture(t)

	response := fixture.setClassification(t, sagaFileNames[0], stringPointer("other"))
	if response.Classification != "other" || !response.IsOverride {
		t.Fatalf("response = %+v", response)
	}
	if got := fixture.movieNames(t); !reflect.DeepEqual(got, sagaFileNames[1:]) {
		t.Fatalf("movies = %v, want %v", got, sagaFileNames[1:])
	}
	var rebuildInputClassification string
	err := fixture.repository.DbContext.GetDatabase().QueryRow(
		`SELECT COALESCE(CASE vmo.classification WHEN 'other' THEN 'program' ELSE vmo.classification END, '')
		 FROM video_manual_override vmo WHERE vmo.file_id = $1`, fixture.videoIDs[sagaFileNames[0]]).Scan(&rebuildInputClassification)
	if err != nil || rebuildInputClassification != "program" {
		t.Fatalf("mapped classification = %q, err = %v", rebuildInputClassification, err)
	}
	videos, err := fixture.repository.GetAllVideosWithMetadata()
	if err != nil {
		t.Fatalf("rebuild input: %v", err)
	}
	for _, video := range videos {
		if video.ID == fixture.videoIDs[sagaFileNames[0]] && video.MetaClassification.String != "program" {
			t.Fatalf("rebuild input classification = %q, want program", video.MetaClassification.String)
		}
	}
}

func TestClassificationOverrideSurvivesIndexingWrites_Postgres(t *testing.T) {
	fixture := newClassificationFixture(t)
	fixture.setClassification(t, sagaFileNames[0], stringPointer("series"))

	stored := fixture.upsertMetadata(t, sagaFileNames[0], "movie", CurrentVideoClassificationVersion)
	if err := fixture.metadataRepository.UpdateVideoClassification(stored.ID, "movie"); err != nil {
		t.Fatalf("backfill classification: %v", err)
	}

	if got := fixture.movieNames(t); !reflect.DeepEqual(got, sagaFileNames[1:]) {
		t.Fatalf("movies after indexing = %v, want %v", got, sagaFileNames[1:])
	}
	if err := fixture.service.RebuildSmartPlaylists(); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	assertOverrideRowCount(t, fixture.repository, 1)
}

func TestClassificationRejectsUnknownValueAndUnknownFile_Postgres(t *testing.T) {
	fixture := newClassificationFixture(t)

	_, err := fixture.service.SetVideoClassification(fixture.videoIDs[sagaFileNames[0]], stringPointer("anime"))
	var validationError *VideoFieldValidationError
	if !asValidationError(err, &validationError) || validationError.Field != "classification" {
		t.Fatalf("err = %v, want classification validation error", err)
	}
	if _, err := fixture.service.SetVideoClassification(999999, stringPointer("movie")); err == nil || !isNoRows(err) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
	assertOverrideRowCount(t, fixture.repository, 0)
}

func TestEpisodeOverrideReordersImmediatelyAndSurvivesRebuild_Postgres(t *testing.T) {
	fixture := newSeriesFixture(t, "/data/Show", shuffledEpisodeNames("Ep ", 16))
	playlistID := fixture.seriesPlaylistID(t)

	response, err := fixture.service.SetVideoEpisode(fixture.videoIDs["Ep 5.mkv"], intPointer(2), intPointer(1))
	if err != nil {
		t.Fatalf("set episode: %v", err)
	}
	if response.Season == nil || *response.Season != 2 || response.Episode == nil || *response.Episode != 1 {
		t.Fatalf("response = %+v", response)
	}
	orderedNames := fixture.orderedNames(t, playlistID)
	if orderedNames[len(orderedNames)-1] != "Ep 5.mkv" {
		t.Fatalf("Ep 5 must move to the end right away, got %v", orderedNames)
	}

	if err := fixture.service.RebuildSmartPlaylists(); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if got := fixture.orderedNames(t, fixture.seriesPlaylistID(t)); !reflect.DeepEqual(got, orderedNames) {
		t.Fatalf("order after rebuild = %v, want %v", got, orderedNames)
	}

	cleared, err := fixture.service.SetVideoEpisode(fixture.videoIDs["Ep 5.mkv"], nil, nil)
	if err != nil {
		t.Fatalf("clear episode: %v", err)
	}
	if cleared.Season == nil || *cleared.Season != 1 || cleared.Episode == nil || *cleared.Episode != 5 {
		t.Fatalf("effective after clearing = %+v", cleared)
	}
	if got := fixture.orderedNames(t, playlistID); got[4] != "Ep 5.mkv" {
		t.Fatalf("Ep 5 must return to position 5, got %v", got)
	}
	assertOverrideRowCount(t, fixture.repository, 0)
}

func TestEpisodeOverrideRejectsOutOfRangeNumbers_Postgres(t *testing.T) {
	fixture := newSeriesFixture(t, "/data/Show", shuffledEpisodeNames("Ep ", 16))
	fileID := fixture.videoIDs["Ep 5.mkv"]

	for _, testCase := range []struct {
		season, episode *int
		wantField       string
	}{
		{intPointer(-1), intPointer(1), "season"},
		{intPointer(1), intPointer(0), "episode"},
	} {
		_, err := fixture.service.SetVideoEpisode(fileID, testCase.season, testCase.episode)
		var validationError *VideoFieldValidationError
		if !asValidationError(err, &validationError) || validationError.Field != testCase.wantField {
			t.Fatalf("err = %v, want %s validation error", err, testCase.wantField)
		}
	}
	if _, err := fixture.service.SetVideoEpisode(999999, intPointer(1), intPointer(1)); !isNoRows(err) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
	assertOverrideRowCount(t, fixture.repository, 0)
}

func assertOverrideRowCount(t *testing.T, repository *Repository, wantCount int) {
	t.Helper()
	var rowCount int
	if err := repository.DbContext.GetDatabase().QueryRow(`SELECT COUNT(*) FROM video_manual_override`).Scan(&rowCount); err != nil {
		t.Fatalf("count overrides: %v", err)
	}
	if rowCount != wantCount {
		t.Fatalf("override rows = %d, want %d", rowCount, wantCount)
	}
}

func asValidationError(err error, target **VideoFieldValidationError) bool {
	return errors.As(err, target)
}

func isNoRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }
