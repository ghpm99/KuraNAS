package music

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func newTrackMetadataTestEnvironment(t *testing.T) *catalogTestEnvironment {
	t.Helper()
	environment := newCatalogTestEnvironment(t)
	environment.service.TrackMetadataStore = environment.audioRepository
	return environment
}

func (environment *catalogTestEnvironment) updateMetadata(t *testing.T, fileName string, body string) {
	t.Helper()
	if _, err := environment.service.UpdateTrackMetadata(environment.fileIDsByName[fileName], decodePatch(t, body)); err != nil {
		t.Fatalf("update metadata of %s with %s: %v", fileName, body, err)
	}
}

func (environment *catalogTestEnvironment) overrideRowCount(t *testing.T, fileName string) int {
	t.Helper()
	var rowCount int
	err := environment.dbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT COUNT(*) FROM audio_metadata_override WHERE file_id = $1`, environment.fileIDsByName[fileName]).Scan(&rowCount)
	})
	if err != nil {
		t.Fatalf("count overrides of %s: %v", fileName, err)
	}
	return rowCount
}

type effectiveNumbers struct {
	year        sql.NullString
	discNumber  sql.NullInt64
	trackNumber sql.NullInt64
}

func (environment *catalogTestEnvironment) readEffectiveNumbers(t *testing.T, fileName string) effectiveNumbers {
	t.Helper()
	var numbers effectiveNumbers
	err := environment.dbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT year, disc_number, track_no FROM audio_metadata_effective WHERE file_id = $1`,
			environment.fileIDsByName[fileName]).Scan(&numbers.year, &numbers.discNumber, &numbers.trackNumber)
	})
	if err != nil {
		t.Fatalf("read effective numbers of %s: %v", fileName, err)
	}
	return numbers
}

func TestUpdateTrackMetadataMovesUntaggedTrackIntoEveryGivenGenre_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, untaggedAndPopFixture(*environment))

	environment.updateMetadata(t, "nogenre.mp3", `{"genres": ["Rock", "Pop"]}`)

	if got := environment.genreKeys(t); !slices.Equal(got, []string{"pop", "rock"}) {
		t.Fatalf("genres = %v, want [pop rock]", got)
	}
	if got := environment.trackNamesOfGenre(t, "rock"); !slices.Equal(got, []string{"nogenre.mp3"}) {
		t.Fatalf("tracks of rock = %v", got)
	}
	popTracks := environment.trackNamesOfGenre(t, "pop")
	slices.Sort(popTracks)
	if !slices.Equal(popTracks, []string{"nogenre.mp3", "pop.mp3"}) {
		t.Fatalf("tracks of pop = %v", popTracks)
	}
}

func TestUpdateTrackMetadataKeepsCorrectedGenreWhenOnlyTitleChanges_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, untaggedAndPopFixture(*environment))
	environment.updateMetadata(t, "nogenre.mp3", `{"genres": ["Rock", "Pop"]}`)

	environment.updateMetadata(t, "nogenre.mp3", `{"title": "Renamed"}`)

	if got := environment.trackNamesOfGenre(t, "rock"); !slices.Equal(got, []string{"nogenre.mp3"}) {
		t.Fatalf("tracks of rock after title change = %v", got)
	}
	track, err := environment.service.GetLibraryTrackByID(environment.fileIDsByName["nogenre.mp3"])
	if err != nil {
		t.Fatalf("GetLibraryTrackByID: %v", err)
	}
	metadata := track.Metadata.(AudioMetadataModel)
	if metadata.Title != "Renamed" || metadata.Genre != "Rock;Pop" {
		t.Fatalf("effective title=%q genre=%q", metadata.Title, metadata.Genre)
	}
}

func TestUpdateTrackMetadataNullGenresRestoresTagGenreAndDropsEmptyOverride_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, untaggedAndPopFixture(*environment))
	environment.updateMetadata(t, "pop.mp3", `{"genres": ["Rock"]}`)
	if got := environment.genreKeys(t); !slices.Equal(got, []string{"rock"}) {
		t.Fatalf("genres with correction = %v, want [rock]", got)
	}

	environment.updateMetadata(t, "pop.mp3", `{"genres": null}`)

	if got := environment.genreKeys(t); !slices.Equal(got, []string{"pop"}) {
		t.Fatalf("genres after reset = %v, want [pop]", got)
	}
	if rowCount := environment.overrideRowCount(t, "pop.mp3"); rowCount != 0 {
		t.Fatalf("override row must disappear when every field is back to the tag, found %d", rowCount)
	}
}

func TestUpdateTrackMetadataKeepsOverrideRowWhileAnotherFieldIsCorrected_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, untaggedAndPopFixture(*environment))
	environment.updateMetadata(t, "pop.mp3", `{"genres": ["Rock"], "title": "Fixed"}`)

	environment.updateMetadata(t, "pop.mp3", `{"genres": []}`)

	if rowCount := environment.overrideRowCount(t, "pop.mp3"); rowCount != 1 {
		t.Fatalf("override row must stay while title is corrected, found %d", rowCount)
	}
	if got := environment.genreKeys(t); !slices.Equal(got, []string{"pop"}) {
		t.Fatalf("genres after reset = %v, want [pop]", got)
	}
}

func TestUpdateTrackMetadataEmptyTextRestoresTagValue_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, untaggedAndPopFixture(*environment))
	environment.updateMetadata(t, "pop.mp3", `{"title": "Fixed"}`)

	environment.updateMetadata(t, "pop.mp3", `{"title": ""}`)

	track, err := environment.service.GetLibraryTrackByID(environment.fileIDsByName["pop.mp3"])
	if err != nil {
		t.Fatalf("GetLibraryTrackByID: %v", err)
	}
	if title := track.Metadata.(AudioMetadataModel).Title; title != "Poppy" {
		t.Fatalf("title = %q, want the tag value Poppy", title)
	}
	if rowCount := environment.overrideRowCount(t, "pop.mp3"); rowCount != 0 {
		t.Fatalf("override row must disappear, found %d", rowCount)
	}
}

func TestUpdateTrackMetadataAlbumArtistMovesTrackToAnotherArtistAndAlbum_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "solo.mp3", parentPath: "/music/solo", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "S1", Artist: "Solo Guy", Album: "Debut"}},
	})
	before := environment.readKeys(t, "solo.mp3")

	environment.updateMetadata(t, "solo.mp3", `{"album_artist": "The Band", "album": "Debut Fixed"}`)

	after := environment.readKeys(t, "solo.mp3")
	if after == before {
		t.Fatalf("keys must change, still %+v", after)
	}
	if after.artistKey != "the band" || after.albumKey != "the band::debut fixed" || after.albumArtistLabel.Valid {
		t.Fatalf("keys after correction = %+v", after)
	}
	artists, err := environment.service.GetLibraryArtists(1, 50, DefaultCatalogSort())
	if err != nil {
		t.Fatalf("artists: %v", err)
	}
	if len(artists.Items) != 1 || artists.Items[0].Key != "the band" {
		t.Fatalf("artists = %+v, want only the band", artists.Items)
	}
}

func TestUpdateTrackMetadataPersistsNumbersAndNullRestoresTag_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "numbered.mp3", parentPath: "/m", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "N", Artist: "A", Album: "B", Year: "2001", TrackNumber: "9/12", DiscNumberText: "3"}},
	})
	tagNumbers := environment.readEffectiveNumbers(t, "numbered.mp3")

	environment.updateMetadata(t, "numbered.mp3", `{"year": 1999, "disc_number": 2, "track_number": 5}`)

	corrected := environment.readEffectiveNumbers(t, "numbered.mp3")
	if corrected.year.String != "1999" || corrected.discNumber.Int64 != 2 || corrected.trackNumber.Int64 != 5 {
		t.Fatalf("corrected numbers = %+v", corrected)
	}

	environment.updateMetadata(t, "numbered.mp3", `{"year": null, "disc_number": null, "track_number": null}`)

	if restored := environment.readEffectiveNumbers(t, "numbered.mp3"); restored != tagNumbers {
		t.Fatalf("restored numbers %+v, want tag numbers %+v", restored, tagNumbers)
	}
	if rowCount := environment.overrideRowCount(t, "numbered.mp3"); rowCount != 0 {
		t.Fatalf("override row must disappear, found %d", rowCount)
	}
}

func TestUpdateTrackMetadataReturnsNoRowsForDeletedOrUnindexedTrack_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "deleted.mp3", parentPath: "/m", updatedAt: environment.catalogBaseMoment, isDeleted: true,
			metadata: &AudioMetadataModel{Title: "Gone", Artist: "A", Album: "B"}},
		{name: "unindexed.mp3", parentPath: "/m", updatedAt: environment.catalogBaseMoment},
	})

	for _, fileName := range []string{"deleted.mp3", "unindexed.mp3"} {
		_, err := environment.service.UpdateTrackMetadata(environment.fileIDsByName[fileName], decodePatch(t, `{"title": "X"}`))
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("%s: expected sql.ErrNoRows, got %v", fileName, err)
		}
		if rowCount := environment.overrideRowCount(t, fileName); rowCount != 0 {
			t.Fatalf("%s: no override may be written, found %d", fileName, rowCount)
		}
	}
	if _, err := environment.service.UpdateTrackMetadata(999999, decodePatch(t, `{}`)); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown id: expected sql.ErrNoRows, got %v", err)
	}
}

func TestUpdateTrackMetadataNeverWritesTheAudioFile_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	audioDirectory := t.TempDir()
	audioPath := filepath.Join(audioDirectory, "song.mp3")
	if err := os.WriteFile(audioPath, []byte("fake audio payload"), 0o644); err != nil {
		t.Fatalf("write audio file: %v", err)
	}
	checksumBefore := sha256.Sum256([]byte("fake audio payload"))
	statBefore, _ := os.Stat(audioPath)
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "song.mp3", parentPath: audioDirectory, updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "Song", Artist: "A", Album: "B"}},
	})

	environment.updateMetadata(t, "song.mp3", `{"title": "Fixed", "genres": ["Rock"], "year": 2000}`)

	audioBytes, err := os.ReadFile(audioPath)
	if err != nil {
		t.Fatalf("read audio file: %v", err)
	}
	statAfter, _ := os.Stat(audioPath)
	if sha256.Sum256(audioBytes) != checksumBefore || !statAfter.ModTime().Equal(statBefore.ModTime()) {
		t.Fatal("audio file must not be touched by a metadata correction")
	}
	entries, _ := os.ReadDir(audioDirectory)
	if len(entries) != 1 {
		t.Fatalf("no sibling file may be created, found %d entries", len(entries))
	}
}

func TestUpdateTrackMetadataReturnsTrackWithCorrectedValues_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, untaggedAndPopFixture(*environment))

	track, err := environment.service.UpdateTrackMetadata(environment.fileIDsByName["nogenre.mp3"], decodePatch(t, `{"title": "Fixed", "artist": "New Artist", "genres": ["Rock"]}`))
	if err != nil {
		t.Fatalf("UpdateTrackMetadata: %v", err)
	}

	metadata, isAudioMetadata := track.Metadata.(AudioMetadataModel)
	if track.ID != environment.fileIDsByName["nogenre.mp3"] || !isAudioMetadata {
		t.Fatalf("unexpected track %+v", track)
	}
	if metadata.Title != "Fixed" || metadata.Artist != "New Artist" || metadata.Genre != "Rock" || metadata.Album != "Album Z" {
		t.Fatalf("returned metadata = %+v", metadata)
	}
}
