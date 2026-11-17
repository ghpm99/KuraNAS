package music

import (
	"database/sql"
	"errors"
	"slices"
	"strings"
	"testing"
)

type genreSummaryRepositoryMock struct {
	*musicRepoMock
	summary MusicGroupSummaryDto
}

func (mock *genreSummaryRepositoryMock) GetLibraryGenreSummary(genreKey string) (MusicGroupSummaryDto, error) {
	return mock.summary, nil
}

func genreRenameService(store *trackMetadataStoreMock, summary MusicGroupSummaryDto) *Service {
	repository := &genreSummaryRepositoryMock{musicRepoMock: &musicRepoMock{}, summary: summary}
	return &Service{Repository: repository, TrackMetadataStore: store}
}

func genresOfGroup(group TrackMetadataPatchGroup) string {
	return strings.Join(group.Patch.Genres.Values, genreSeparator)
}

func TestServiceRenameGenreReplacesOnlyTheRenamedGenreKeepingTheOthers(t *testing.T) {
	store := &trackMetadataStoreMock{genreSources: []GenreTrackSource{
		{FileID: 1, Genre: "Rock n Roll"},
		{FileID: 2, Genre: "rock n roll;Blues"},
		{FileID: 3, Genre: "Blues, Rock N Roll , Jazz"},
	}}

	_, err := genreRenameService(store, MusicGroupSummaryDto{}).RenameGenre("rock n roll", "Rock")

	if err != nil {
		t.Fatalf("RenameGenre: %v", err)
	}
	if store.requestedGenreSourceOf != "rock n roll" || store.applyGroupsCallCount != 1 || len(store.appliedGroups) != 3 {
		t.Fatalf("source key=%q calls=%d groups=%+v", store.requestedGenreSourceOf, store.applyGroupsCallCount, store.appliedGroups)
	}
	wantGenresByFileID := map[int]string{1: "Rock", 2: "Rock;Blues", 3: "Blues;Rock;Jazz"}
	for _, group := range store.appliedGroups {
		if !group.Patch.Genres.IsSet || len(group.FileIDs) != 1 || genresOfGroup(group) != wantGenresByFileID[group.FileIDs[0]] {
			t.Fatalf("group %+v genres=%q", group, genresOfGroup(group))
		}
	}
}

func TestServiceRenameGenreKeepsOneCopyWhenTheTrackAlreadyHadTheTargetGenre(t *testing.T) {
	store := &trackMetadataStoreMock{genreSources: []GenreTrackSource{
		{FileID: 1, Genre: "Rock n Roll;Rock"},
		{FileID: 2, Genre: "Rock;Rock n Roll"},
	}}

	_, err := genreRenameService(store, MusicGroupSummaryDto{}).RenameGenre("rock n roll", "Rock")

	if err != nil {
		t.Fatalf("RenameGenre: %v", err)
	}
	wantGenresByFileID := map[int]string{1: "Rock", 2: "Rock"}
	if len(store.appliedGroups) != 1 || !slices.Equal(store.appliedGroups[0].FileIDs, []int{1, 2}) || genresOfGroup(store.appliedGroups[0]) != wantGenresByFileID[1] {
		t.Fatalf("groups = %+v, want one group of both tracks with genre Rock", store.appliedGroups)
	}
}

func TestServiceRenameGenreReportsMergeOnlyWhenTheNewKeyAlreadyHasTracks(t *testing.T) {
	scenarios := map[string]struct {
		existingTrackCounts map[string]int
		genreKey            string
		name                string
		wantMerged          bool
	}{
		"new key already used":       {map[string]int{"rock": 5}, "rock n roll", "Rock", true},
		"new key unused":             {map[string]int{}, "rock n roll", "Heavy Rock", false},
		"only spelling changes":      {map[string]int{"rock n roll": 3}, "rock n roll", "ROCK N ROLL", false},
		"same key different spacing": {map[string]int{"hip hop": 4}, "hip hop", "hip-hop", false},
		"accent-free spelling":       {map[string]int{"musica": 4}, "rock n roll", "Música", true},
	}
	for scenarioName, scenario := range scenarios {
		store := &trackMetadataStoreMock{
			genreSources:         []GenreTrackSource{{FileID: 1, Genre: scenario.genreKey}},
			trackCountByGenreKey: scenario.existingTrackCounts,
		}

		renamedGenre, err := genreRenameService(store, MusicGroupSummaryDto{}).RenameGenre(scenario.genreKey, scenario.name)

		if err != nil || renamedGenre.Merged != scenario.wantMerged {
			t.Fatalf("%s: merged=%v err=%v, want merged=%v", scenarioName, renamedGenre.Merged, err, scenario.wantMerged)
		}
	}
}

func TestServiceRenameGenreMergesAnUnaccentedGenreIntoItsAccentedSpelling(t *testing.T) {
	store := &trackMetadataStoreMock{genreSources: []GenreTrackSource{
		{FileID: 1, Genre: "Musica"},
		{FileID: 2, Genre: "Musica;Rock"},
	}}

	_, err := genreRenameService(store, MusicGroupSummaryDto{}).RenameGenre("musica", "Música")

	if err != nil {
		t.Fatalf("RenameGenre: %v", err)
	}
	wantGenresByFileID := map[int]string{1: "Música", 2: "Música;Rock"}
	for _, group := range store.appliedGroups {
		if genresOfGroup(group) != wantGenresByFileID[group.FileIDs[0]] {
			t.Fatalf("group %+v genres=%q", group, genresOfGroup(group))
		}
	}
}

func TestServiceRenameGenreReturnsTheGenreGroupAfterTheOperation(t *testing.T) {
	store := &trackMetadataStoreMock{genreSources: []GenreTrackSource{{FileID: 1, Genre: "Rock n Roll"}}}
	summary := MusicGroupSummaryDto{Key: "rock", Name: "Rock", TrackCount: 8}

	renamedGenre, err := genreRenameService(store, summary).RenameGenre("rock n roll", "rock")

	want := GenreRenameResultDto{Key: "rock", Name: "Rock", TrackCount: 8, Merged: false}
	if err != nil || renamedGenre != want {
		t.Fatalf("result = %+v err=%v, want %+v", renamedGenre, err, want)
	}
}

func TestServiceRenameGenreAnswersNoRowsWithoutWritingWhenNoTrackHasTheGenre(t *testing.T) {
	store := &trackMetadataStoreMock{}

	_, err := genreRenameService(store, MusicGroupSummaryDto{}).RenameGenre("polka", "Rock")

	if !errors.Is(err, sql.ErrNoRows) || store.applyGroupsCallCount != 0 {
		t.Fatalf("err=%v writes=%d", err, store.applyGroupsCallCount)
	}
}

func TestServiceRenameGenreRefusesTheNameFieldBeforeTouchingTheStore(t *testing.T) {
	longName := strings.Repeat("a", maxMetadataTextLength+1)
	refusals := map[string]struct {
		name       string
		messageKey string
	}{
		"empty":            {"", "ERROR_MUSIC_GENRE_NAME_EMPTY"},
		"blank":            {"   \t", "ERROR_MUSIC_GENRE_NAME_EMPTY"},
		"only punctuation": {" . - _ ", "ERROR_MUSIC_GENRE_NAME_EMPTY"},
		"too long":         {longName, "ERROR_MUSIC_METADATA_TEXT_TOO_LONG"},
		"semicolon":        {"Rock;Pop", "ERROR_MUSIC_GENRE_NAME_SEPARATOR"},
		"comma":            {"Rock, Pop", "ERROR_MUSIC_GENRE_NAME_SEPARATOR"},
		"pipe":             {"Rock|Pop", "ERROR_MUSIC_GENRE_NAME_SEPARATOR"},
		"slash":            {"Pop/Rock", "ERROR_MUSIC_GENRE_NAME_SEPARATOR"},
	}
	for refusalName, refusal := range refusals {
		store := &trackMetadataStoreMock{genreSources: []GenreTrackSource{{FileID: 1, Genre: "Rock n Roll"}}}

		_, err := genreRenameService(store, MusicGroupSummaryDto{}).RenameGenre("rock n roll", refusal.name)

		assertRefusedField(t, err, "name", refusal.messageKey)
		if store.requestedGenreSourceOf != "" || store.applyGroupsCallCount != 0 {
			t.Fatalf("%s: store must not be touched", refusalName)
		}
	}
}

func TestServiceRenameGenreAcceptsNameOfExactlyTheMaximumLength(t *testing.T) {
	store := &trackMetadataStoreMock{genreSources: []GenreTrackSource{{FileID: 1, Genre: "Rock n Roll"}}}

	_, err := genreRenameService(store, MusicGroupSummaryDto{}).RenameGenre("rock n roll", strings.Repeat("a", maxMetadataTextLength))

	if err != nil {
		t.Fatalf("name at the limit refused: %v", err)
	}
}

func TestServiceRenameGenreSurfacesAStoreWriteFailure(t *testing.T) {
	writeFailure := errors.New("write failed")
	store := &trackMetadataStoreMock{genreSources: []GenreTrackSource{{FileID: 1, Genre: "Rock n Roll"}}, applyError: writeFailure}

	_, err := genreRenameService(store, MusicGroupSummaryDto{}).RenameGenre("rock n roll", "Rock")

	if !errors.Is(err, writeFailure) {
		t.Fatalf("err = %v, want the write failure", err)
	}
}
