package music

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"testing"
)

const (
	rockNRollTrackCount = 3
	rockTrackCount      = 5
)

func genreTags(trackLabel string, genre string) AudioMetadataModel {
	return AudioMetadataModel{Title: trackLabel, Artist: "Genre Band", Album: "Genre Album", Genre: genre}
}

func rockFamilyFixture(environment *catalogTestEnvironment) []catalogTrackFixture {
	fixtures := []catalogTrackFixture{}
	for position := 1; position <= rockNRollTrackCount; position++ {
		fileName := fmt.Sprintf("rocknroll%d.mp3", position)
		fixtures = append(fixtures, trackFixture(fileName, "/m", environment.catalogBaseMoment, genreTags(fileName, "Rock n Roll")))
	}
	for position := 1; position <= rockTrackCount; position++ {
		fileName := fmt.Sprintf("rock%d.mp3", position)
		fixtures = append(fixtures, trackFixture(fileName, "/m", environment.catalogBaseMoment, genreTags(fileName, "Rock")))
	}
	return fixtures
}

func (environment *catalogTestEnvironment) renameGenre(t *testing.T, genreKey string, name string) GenreRenameResultDto {
	t.Helper()
	renamedGenre, err := environment.service.RenameGenre(genreKey, name)
	if err != nil {
		t.Fatalf("rename genre %q to %q: %v", genreKey, name, err)
	}
	return renamedGenre
}

func (environment *catalogTestEnvironment) effectiveGenre(t *testing.T, fileName string) string {
	t.Helper()
	track, err := environment.service.GetLibraryTrackByID(environment.fileIDsByName[fileName])
	if err != nil {
		t.Fatalf("GetLibraryTrackByID %s: %v", fileName, err)
	}
	return track.Metadata.(AudioMetadataModel).Genre
}

func TestRenameGenreMergesIntoExistingGenreLeavingASingleRockGroup_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, rockFamilyFixture(environment))
	if got := environment.genreKeys(t); !slices.Equal(got, []string{"rock", "rock n roll"}) {
		t.Fatalf("genres before = %v", got)
	}

	renamedGenre := environment.renameGenre(t, "rock n roll", "Rock")

	want := GenreRenameResultDto{Key: "rock", Name: "Rock", TrackCount: rockNRollTrackCount + rockTrackCount, Merged: true}
	if renamedGenre != want {
		t.Fatalf("result = %+v, want %+v", renamedGenre, want)
	}
	if got := environment.genreKeys(t); !slices.Equal(got, []string{"rock"}) {
		t.Fatalf("genres after = %v, want only rock", got)
	}
	if got := environment.trackNamesOfGenre(t, "rock"); len(got) != rockNRollTrackCount+rockTrackCount {
		t.Fatalf("tracks of rock = %v", got)
	}
	if got := environment.trackNamesOfGenre(t, "rock n roll"); len(got) != 0 {
		t.Fatalf("tracks of rock n roll = %v, want none", got)
	}
}

func TestRenameGenreKeepsTheOtherGenresOfTheTrack_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		trackFixture("mixed.mp3", "/m", environment.catalogBaseMoment, genreTags("Mixed", "Rock n Roll;Blues")),
		trackFixture("both.mp3", "/m", environment.catalogBaseMoment, genreTags("Both", "Rock n Roll;Rock")),
		trackFixture("plain.mp3", "/m", environment.catalogBaseMoment, genreTags("Plain", "Rock")),
	})

	environment.renameGenre(t, "rock n roll", "Rock")

	if got := environment.effectiveGenre(t, "mixed.mp3"); got != "Rock;Blues" {
		t.Fatalf("mixed genre = %q, want Rock;Blues", got)
	}
	if got := environment.effectiveGenre(t, "both.mp3"); got != "Rock" {
		t.Fatalf("both genre = %q, want a single Rock", got)
	}
	if got := environment.genreKeys(t); !slices.Equal(got, []string{"blues", "rock"}) {
		t.Fatalf("genres = %v", got)
	}
	if got := environment.sortedNamesOfGenre(t, "rock"); !slices.Equal(got, []string{"both.mp3", "mixed.mp3", "plain.mp3"}) {
		t.Fatalf("tracks of rock = %v", got)
	}
}

func TestRenameGenrePreservesTheTitleCorrectionOfTheTrack_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, rockFamilyFixture(environment))
	environment.updateMetadata(t, "rocknroll1.mp3", `{"title": "Corrected Title"}`)

	environment.renameGenre(t, "rock n roll", "Rock")

	track, err := environment.service.GetLibraryTrackByID(environment.fileIDsByName["rocknroll1.mp3"])
	if err != nil {
		t.Fatalf("GetLibraryTrackByID: %v", err)
	}
	metadata := track.Metadata.(AudioMetadataModel)
	if metadata.Title != "Corrected Title" || metadata.Genre != "Rock" {
		t.Fatalf("effective title=%q genre=%q", metadata.Title, metadata.Genre)
	}
}

func TestRenameGenreSurvivesReindexingWithTheOriginalTags_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, rockFamilyFixture(environment))
	environment.renameGenre(t, "rock n roll", "Rock")

	for position := 1; position <= rockNRollTrackCount; position++ {
		fileName := fmt.Sprintf("rocknroll%d.mp3", position)
		environment.reindex(t, fileName, genreTags(fileName, "Rock n Roll"))
	}

	if got := environment.genreKeys(t); !slices.Equal(got, []string{"rock"}) {
		t.Fatalf("genres after reindex = %v, want only rock", got)
	}
	if got := environment.trackNamesOfGenre(t, "rock"); len(got) != rockNRollTrackCount+rockTrackCount {
		t.Fatalf("tracks of rock after reindex = %v", got)
	}
}

func TestRenameGenreToANewNameIsNotAMerge_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, rockFamilyFixture(environment))

	renamedGenre := environment.renameGenre(t, "rock n roll", "Classic Rock")

	want := GenreRenameResultDto{Key: "classic rock", Name: "Classic Rock", TrackCount: rockNRollTrackCount, Merged: false}
	if renamedGenre != want {
		t.Fatalf("result = %+v, want %+v", renamedGenre, want)
	}
	if got := environment.genreKeys(t); !slices.Equal(got, []string{"classic rock", "rock"}) {
		t.Fatalf("genres = %v", got)
	}
	if got := environment.trackNamesOfGenre(t, "rock"); len(got) != rockTrackCount {
		t.Fatalf("tracks of rock = %v, want the original %d", got, rockTrackCount)
	}
}

func TestRenameGenreChangingOnlyTheSpellingIsNotAMerge_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, rockFamilyFixture(environment))

	renamedGenre := environment.renameGenre(t, "rock n roll", "ROCK N ROLL")

	want := GenreRenameResultDto{Key: "rock n roll", Name: "Rock N Roll", TrackCount: rockNRollTrackCount, Merged: false}
	if renamedGenre != want {
		t.Fatalf("result = %+v, want %+v", renamedGenre, want)
	}
}

func TestRenameGenreIgnoresDeletedTracksAndAnswersNoRowsForUnknownGenre_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "gone.mp3", parentPath: "/m", updatedAt: environment.catalogBaseMoment, isDeleted: true,
			metadata: &AudioMetadataModel{Title: "Gone", Artist: "A", Album: "B", Genre: "Polka"}},
		trackFixture("alive.mp3", "/m", environment.catalogBaseMoment, genreTags("Alive", "Rock")),
	})

	_, err := environment.service.RenameGenre("polka", "Rock")

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
	if environment.overrideRowCount(t, "gone.mp3") != 0 {
		t.Fatalf("deleted track must not receive a correction")
	}
}
