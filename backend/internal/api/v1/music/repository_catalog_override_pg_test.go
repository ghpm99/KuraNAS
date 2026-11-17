package music

import (
	"database/sql"
	"slices"
	"testing"
)

type catalogKeysRow struct {
	artistKey        string
	albumKey         string
	albumArtistLabel sql.NullString
}

func (environment *catalogTestEnvironment) setOverride(t *testing.T, fileName string, columnAssignments string, args ...any) {
	t.Helper()
	fileID := environment.fileIDsByName[fileName]
	err := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`INSERT INTO audio_metadata_override (file_id) VALUES ($1) ON CONFLICT (file_id) DO NOTHING`, fileID); err != nil {
			return err
		}
		_, err := tx.Exec(`UPDATE audio_metadata_override SET `+columnAssignments+` WHERE file_id = $1`, append([]any{fileID}, args...)...)
		return err
	})
	if err != nil {
		t.Fatalf("set override of %s: %v", fileName, err)
	}
}

func (environment *catalogTestEnvironment) clearOverride(t *testing.T, fileName string) {
	t.Helper()
	err := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`DELETE FROM audio_metadata_override WHERE file_id = $1`, environment.fileIDsByName[fileName])
		return err
	})
	if err != nil {
		t.Fatalf("clear override of %s: %v", fileName, err)
	}
}

func (environment *catalogTestEnvironment) recomputeKeys(t *testing.T, fileNames ...string) {
	t.Helper()
	fileIDs := make([]int, 0, len(fileNames))
	for _, fileName := range fileNames {
		fileIDs = append(fileIDs, environment.fileIDsByName[fileName])
	}
	err := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		return environment.audioRepository.RecomputeCatalogKeys(tx, fileIDs)
	})
	if err != nil {
		t.Fatalf("recompute keys: %v", err)
	}
}

func (environment *catalogTestEnvironment) reindex(t *testing.T, fileName string, tags AudioMetadataModel) {
	t.Helper()
	tags.FileId = environment.fileIDsByName[fileName]
	tags.Path = "/m/" + fileName
	err := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := environment.audioRepository.UpsertAudioMetadata(tx, tags)
		return err
	})
	if err != nil {
		t.Fatalf("reindex %s: %v", fileName, err)
	}
}

func (environment *catalogTestEnvironment) readKeys(t *testing.T, fileName string) catalogKeysRow {
	t.Helper()
	var keys catalogKeysRow
	err := environment.dbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT catalog_artist_key, catalog_album_key, catalog_album_artist_label FROM audio_metadata WHERE file_id = $1`,
			environment.fileIDsByName[fileName]).Scan(&keys.artistKey, &keys.albumKey, &keys.albumArtistLabel)
	})
	if err != nil {
		t.Fatalf("read keys of %s: %v", fileName, err)
	}
	return keys
}

func (environment *catalogTestEnvironment) genreKeys(t *testing.T) []string {
	t.Helper()
	genres, err := environment.service.GetLibraryGenres(1, 50, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("genres: %v", err)
	}
	keys := []string{}
	for _, genre := range genres.Items {
		keys = append(keys, genre.Key)
	}
	slices.Sort(keys)
	return keys
}

func (environment *catalogTestEnvironment) trackNamesOfGenre(t *testing.T, genreKey string) []string {
	t.Helper()
	tracks, err := environment.service.Repository.GetLibraryTrackIDsByGenre(genreKey, 1, 50)
	if err != nil {
		t.Fatalf("tracks of genre %s: %v", genreKey, err)
	}
	return environment.namesOf(tracks.Items)
}

func untaggedAndPopFixture(base catalogTestEnvironment) []catalogTrackFixture {
	return []catalogTrackFixture{
		{name: "nogenre.mp3", parentPath: "/m", updatedAt: base.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "Untagged", Artist: "Artist Z", Album: "Album Z"}},
		{name: "pop.mp3", parentPath: "/m", updatedAt: base.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "Poppy", Artist: "Artist P", Album: "Album P", Genre: "Pop"}},
	}
}

func TestGenreOverrideMovesUntaggedTrackIntoGenreGroups_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, untaggedAndPopFixture(*environment))

	if got := environment.genreKeys(t); !slices.Equal(got, []string{"pop"}) {
		t.Fatalf("genres before override = %v, want [pop]", got)
	}

	environment.setOverride(t, "nogenre.mp3", `genre = $2`, "Rock")
	environment.recomputeKeys(t, "nogenre.mp3")

	if got := environment.genreKeys(t); !slices.Equal(got, []string{"pop", "rock"}) {
		t.Fatalf("genres after override = %v, want [pop rock]", got)
	}
	if got := environment.trackNamesOfGenre(t, "rock"); !slices.Equal(got, []string{"nogenre.mp3"}) {
		t.Fatalf("tracks of rock = %v", got)
	}

	rockQueue, err := environment.service.GetLibraryQueueByGenre("rock")
	if err != nil || len(rockQueue.Items) != 1 {
		t.Fatalf("rock queue = %+v err=%v", rockQueue, err)
	}
}

func TestGenreOverrideSurvivesReindexingWithOriginalTags_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, untaggedAndPopFixture(*environment))
	environment.setOverride(t, "nogenre.mp3", `genre = $2`, "Rock")
	environment.recomputeKeys(t, "nogenre.mp3")

	environment.reindex(t, "nogenre.mp3", AudioMetadataModel{Title: "Untagged", Artist: "Artist Z", Album: "Album Z"})

	if got := environment.trackNamesOfGenre(t, "rock"); !slices.Equal(got, []string{"nogenre.mp3"}) {
		t.Fatalf("tracks of rock after reindex = %v", got)
	}
}

func TestClearingGenreOverrideRestoresTagGenre_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, untaggedAndPopFixture(*environment))
	environment.setOverride(t, "pop.mp3", `genre = $2`, "Rock")
	environment.recomputeKeys(t, "pop.mp3")

	if got := environment.genreKeys(t); !slices.Equal(got, []string{"rock"}) {
		t.Fatalf("genres with override = %v, want [rock]", got)
	}

	environment.clearOverride(t, "pop.mp3")
	environment.recomputeKeys(t, "pop.mp3")

	if got := environment.genreKeys(t); !slices.Equal(got, []string{"pop"}) {
		t.Fatalf("genres after clearing override = %v, want [pop]", got)
	}
}

func TestAlbumArtistOverrideMovesTrackAndReconcileKeepsCompilation_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "comp1.mp3", parentPath: "/music/comp", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "C1", Artist: "Artist A", Album: "Hits"}},
		{name: "comp2.mp3", parentPath: "/music/comp", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "C2", Artist: "Artist B", Album: "Hits"}},
		{name: "solo.mp3", parentPath: "/music/solo", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "S1", Artist: "Solo Guy", Album: "Debut"}},
	})
	if _, err := environment.audioRepository.ReconcileAlbumGroupings("Various Artists"); err != nil {
		t.Fatalf("reconcile before override: %v", err)
	}
	compilationKey := environment.readKeys(t, "comp1.mp3")
	if !compilationKey.albumArtistLabel.Valid || compilationKey.albumArtistLabel.String != "Various Artists" {
		t.Fatalf("compilation must be recognised before override: %+v", compilationKey)
	}

	environment.setOverride(t, "solo.mp3", `album_artist = $2, album = $3`, "The Band", "Debut Fixed")
	environment.recomputeKeys(t, "solo.mp3")
	if _, err := environment.audioRepository.ReconcileAlbumGroupings("Various Artists"); err != nil {
		t.Fatalf("reconcile after override: %v", err)
	}

	soloKeys := environment.readKeys(t, "solo.mp3")
	if soloKeys.artistKey != "the band" || soloKeys.albumKey != "the band::debut fixed" || soloKeys.albumArtistLabel.Valid {
		t.Fatalf("solo keys after override = %+v", soloKeys)
	}
	if after := environment.readKeys(t, "comp1.mp3"); after != compilationKey {
		t.Fatalf("compilation keys changed: before %+v after %+v", compilationKey, after)
	}

	artists, err := environment.service.GetLibraryArtists(1, 50, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("artists: %v", err)
	}
	artistKeys := []string{}
	for _, artist := range artists.Items {
		artistKeys = append(artistKeys, artist.Key)
	}
	slices.Sort(artistKeys)
	if !slices.Equal(artistKeys, []string{"artist a", "artist b", "the band"}) {
		t.Fatalf("artist keys = %v", artistKeys)
	}
}

func TestReindexKeepsAlbumArtistOverrideInCatalogKeys_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "solo.mp3", parentPath: "/m", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "S1", Artist: "Solo Guy", Album: "Debut"}},
	})
	environment.setOverride(t, "solo.mp3", `album_artist = $2`, "The Band")
	environment.recomputeKeys(t, "solo.mp3")

	environment.reindex(t, "solo.mp3", AudioMetadataModel{Title: "S1", Artist: "Solo Guy", Album: "Debut"})

	keys := environment.readKeys(t, "solo.mp3")
	if keys.artistKey != "the band" || keys.albumKey != "the band::debut" {
		t.Fatalf("keys after reindex = %+v", keys)
	}
}

func TestDisplayedValuesAndSearchUseOverriddenTrackMetadata_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "fixme.mp3", parentPath: "/m", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "Wrongtitle", Artist: "Wrongartist", Album: "Wrongalbum", TrackNumber: "9", Year: "1980"}},
	})
	environment.setOverride(t, "fixme.mp3", `title = $2, artist = $3, album = $4, year = $5, track_number = $6`,
		"Righttitle", "Rightartist", "Rightalbum", "2001", 4)
	environment.recomputeKeys(t, "fixme.mp3")
	fileID := environment.fileIDsByName["fixme.mp3"]

	hydrated, err := environment.service.Repository.GetLibraryFilesByIDs([]int{fileID})
	if err != nil || len(hydrated) != 1 {
		t.Fatalf("hydrate by id: %v len=%d", err, len(hydrated))
	}
	hydratedMetadata := hydrated[0].Metadata.(AudioMetadataModel)
	if hydratedMetadata.Title != "Righttitle" || hydratedMetadata.Artist != "Rightartist" || hydratedMetadata.Album != "Rightalbum" ||
		hydratedMetadata.Year != "2001" || hydratedMetadata.TrackNumber != "4" {
		t.Fatalf("hydrated metadata = %+v", hydratedMetadata)
	}

	queue, err := environment.service.Repository.GetLibraryQueueByFileIDs([]int{fileID})
	if err != nil || len(queue) != 1 || queue[0].Title != "Righttitle" || queue[0].Artist != "Rightartist" || queue[0].Album != "Rightalbum" {
		t.Fatalf("queue = %+v err=%v", queue, err)
	}

	summary, err := NewAudioSummaryRepository(environment.dbContext).GetAudioSummaryByFileID(fileID)
	if err != nil || summary.Title != "Righttitle" || summary.Artist != "Rightartist" {
		t.Fatalf("summary = %+v err=%v", summary, err)
	}

	foundByNewTitle, err := environment.service.Repository.SearchLibraryTracks("righttitle", 1, 10)
	if err != nil || len(foundByNewTitle.Items) != 1 {
		t.Fatalf("search by corrected title = %+v err=%v", foundByNewTitle.Items, err)
	}
	foundByOldTitle, err := environment.service.Repository.SearchLibraryTracks("wrongtitle", 1, 10)
	if err != nil || len(foundByOldTitle.Items) != 0 {
		t.Fatalf("search by replaced tag title must find nothing, got %+v err=%v", foundByOldTitle.Items, err)
	}
	foundByNewArtist, err := environment.service.Repository.SearchLibraryTracks("rightartist", 1, 10)
	if err != nil || len(foundByNewArtist.Items) != 1 {
		t.Fatalf("search by corrected artist = %+v err=%v", foundByNewArtist.Items, err)
	}
}

func TestCatalogKeysBackfillUsesEffectiveValues_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, untaggedAndPopFixture(*environment))
	environment.setOverride(t, "nogenre.mp3", `genre = $2`, "Rock")

	resetErr := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE audio_metadata SET catalog_artist_key = NULL, catalog_artist_label = NULL,
			catalog_album_key = NULL, catalog_album_label = NULL, catalog_genre_keys = NULL, catalog_genre_labels = NULL`)
		return err
	})
	if resetErr != nil {
		t.Fatalf("reset keys: %v", resetErr)
	}

	pendingRows, err := environment.audioRepository.ListAudioWithoutCatalogKeys(0, 10)
	if err != nil || len(pendingRows) != 2 {
		t.Fatalf("pending rows = %+v err=%v", pendingRows, err)
	}
	pendingFileIDs := []int{pendingRows[0].FileID, pendingRows[1].FileID}
	backfillErr := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		return environment.audioRepository.RecomputeCatalogKeys(tx, pendingFileIDs)
	})
	if backfillErr != nil {
		t.Fatalf("backfill: %v", backfillErr)
	}

	if got := environment.trackNamesOfGenre(t, "rock"); !slices.Equal(got, []string{"nogenre.mp3"}) {
		t.Fatalf("tracks of rock after backfill = %v", got)
	}
}
