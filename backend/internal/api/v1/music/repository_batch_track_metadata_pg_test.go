package music

import (
	"fmt"
	"slices"
	"testing"
)

const untaggedAlbumTrackCount = 30

func untaggedAlbumTags(trackPosition int) AudioMetadataModel {
	return AudioMetadataModel{
		Title: fmt.Sprintf("Song %02d", trackPosition), Artist: "Batch Band", Album: "Batch Album",
		TrackNumber: fmt.Sprintf("%d", trackPosition),
	}
}

func untaggedAlbumFixture(environment *catalogTestEnvironment) []catalogTrackFixture {
	fixtures := []catalogTrackFixture{}
	for trackPosition := 1; trackPosition <= untaggedAlbumTrackCount; trackPosition++ {
		fixtures = append(fixtures, trackFixture(untaggedAlbumFileName(trackPosition), "/m", environment.catalogBaseMoment, untaggedAlbumTags(trackPosition)))
	}
	return append(fixtures,
		trackFixture("elsewhere.mp3", "/m/elsewhere", environment.catalogBaseMoment, AudioMetadataModel{Title: "Else", Artist: "Other Band", Album: "Other Album"}),
	)
}

func untaggedAlbumFileName(trackPosition int) string {
	return fmt.Sprintf("song%02d.mp3", trackPosition)
}

func untaggedAlbumFileNames() []string {
	fileNames := []string{}
	for trackPosition := 1; trackPosition <= untaggedAlbumTrackCount; trackPosition++ {
		fileNames = append(fileNames, untaggedAlbumFileName(trackPosition))
	}
	return fileNames
}

func (environment *catalogTestEnvironment) sortedNamesOfGenre(t *testing.T, genreKey string) []string {
	t.Helper()
	trackNames := environment.trackNamesOfGenre(t, genreKey)
	slices.Sort(trackNames)
	return trackNames
}

func (environment *catalogTestEnvironment) updateBatch(t *testing.T, requestBody string, isDryRun bool) BatchMetadataResultDto {
	t.Helper()
	result, err := environment.service.UpdateTracksMetadata(decodeBatchRequest(t, requestBody), isDryRun)
	if err != nil {
		t.Fatalf("batch update %s: %v", requestBody, err)
	}
	return result
}

func TestUpdateTracksMetadataByAlbumKeyMovesEveryTrackIntoTheGenre_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, untaggedAlbumFixture(environment))
	albumKey := environment.readKeys(t, "song01.mp3").albumKey

	result := environment.updateBatch(t, fmt.Sprintf(`{"fields": {"genres": ["Rock"]}, "album_key": %q}`, albumKey), false)

	if result.Matched != untaggedAlbumTrackCount || result.Updated != untaggedAlbumTrackCount || len(result.Failed) != 0 {
		t.Fatalf("result = %+v", result)
	}
	if got := environment.sortedNamesOfGenre(t, "rock"); !slices.Equal(got, untaggedAlbumFileNames()) {
		t.Fatalf("tracks of rock = %v", got)
	}
	if got := environment.genreKeys(t); !slices.Equal(got, []string{"rock"}) {
		t.Fatalf("genres = %v, want only rock", got)
	}
}

func TestUpdateTracksMetadataDryRunCountsTheAlbumAndWritesNothing_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, untaggedAlbumFixture(environment))
	albumKey := environment.readKeys(t, "song01.mp3").albumKey

	result := environment.updateBatch(t, fmt.Sprintf(`{"fields": {"genres": ["Rock"]}, "album_key": %q}`, albumKey), true)

	if result.Matched != untaggedAlbumTrackCount || result.Updated != 0 || result.Failed == nil || len(result.Failed) != 0 {
		t.Fatalf("result = %+v", result)
	}
	if got := environment.genreKeys(t); len(got) != 0 {
		t.Fatalf("dry run must not create genres, found %v", got)
	}
	if rowCount := environment.overrideRowCount(t, "song01.mp3"); rowCount != 0 {
		t.Fatalf("dry run must not write overrides, found %d", rowCount)
	}
}

func TestUpdateTracksMetadataByFolderKeyIncludesSubfoldersButNotSiblingPrefixes_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	base := environment.catalogBaseMoment
	tags := AudioMetadataModel{Title: "T", Artist: "A", Album: "B"}
	environment.seedTracks(t, []catalogTrackFixture{
		trackFixture("top.mp3", "/m/rock", base, tags),
		trackFixture("nested.mp3", "/m/rock/live/2020", base, tags),
		trackFixture("sibling.mp3", "/m/rockabilly", base, tags),
		trackFixture("other.mp3", "/m/pop", base, tags),
		{name: "gone.mp3", parentPath: "/m/rock", updatedAt: base, isDeleted: true, metadata: &tags},
	})

	dryRun := environment.updateBatch(t, `{"fields": {"genres": ["Rock"]}, "folder_key": "/m/rock"}`, true)
	if dryRun.Matched != 2 || dryRun.Updated != 0 {
		t.Fatalf("dry run = %+v", dryRun)
	}
	result := environment.updateBatch(t, `{"fields": {"genres": ["Rock"]}, "folder_key": "/m/rock"}`, false)

	if result.Matched != 2 || result.Updated != 2 {
		t.Fatalf("result = %+v", result)
	}
	if got := environment.sortedNamesOfGenre(t, "rock"); !slices.Equal(got, []string{"nested.mp3", "top.mp3"}) {
		t.Fatalf("tracks of rock = %v", got)
	}
}

func TestUpdateTracksMetadataByFileIDsReportsDeletedAndUnindexedAndWritesTheRest_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	base := environment.catalogBaseMoment
	tags := AudioMetadataModel{Title: "T", Artist: "A", Album: "B"}
	environment.seedTracks(t, []catalogTrackFixture{
		trackFixture("first.mp3", "/m", base, tags),
		trackFixture("second.mp3", "/m", base, tags),
		{name: "deleted.mp3", parentPath: "/m", updatedAt: base, isDeleted: true, metadata: &tags},
		{name: "unindexed.mp3", parentPath: "/m", updatedAt: base},
	})
	requestedNames := []string{"first.mp3", "deleted.mp3", "second.mp3", "unindexed.mp3"}
	requestedFileIDs := []int{}
	for _, requestedName := range requestedNames {
		requestedFileIDs = append(requestedFileIDs, environment.fileIDsByName[requestedName])
	}
	requestedFileIDs = append(requestedFileIDs, 987654)

	result := environment.updateBatch(t, fmt.Sprintf(`{"fields": {"genres": ["Rock"]}, "file_ids": %s}`, jsonIntList(requestedFileIDs)), false)

	if result.Matched != 2 || result.Updated != 2 {
		t.Fatalf("result = %+v", result)
	}
	failedFileIDs := []int{}
	for _, failure := range result.Failed {
		failedFileIDs = append(failedFileIDs, failure.FileID)
		if failure.Error == "" {
			t.Fatalf("failure without message: %+v", failure)
		}
	}
	wantFailedFileIDs := []int{environment.fileIDsByName["deleted.mp3"], environment.fileIDsByName["unindexed.mp3"], 987654}
	if !slices.Equal(failedFileIDs, wantFailedFileIDs) {
		t.Fatalf("failed = %v, want %v", failedFileIDs, wantFailedFileIDs)
	}
	if got := environment.sortedNamesOfGenre(t, "rock"); !slices.Equal(got, []string{"first.mp3", "second.mp3"}) {
		t.Fatalf("tracks of rock = %v", got)
	}
	for _, untouchedName := range []string{"deleted.mp3", "unindexed.mp3"} {
		if rowCount := environment.overrideRowCount(t, untouchedName); rowCount != 0 {
			t.Fatalf("%s must not get an override, found %d", untouchedName, rowCount)
		}
	}
}

func TestUpdateTracksMetadataSurvivesReindexingWithTheOriginalTags_Postgres(t *testing.T) {
	environment := newTrackMetadataTestEnvironment(t)
	environment.seedTracks(t, untaggedAlbumFixture(environment))
	albumKey := environment.readKeys(t, "song01.mp3").albumKey
	environment.updateBatch(t, fmt.Sprintf(`{"fields": {"genres": ["Rock"]}, "album_key": %q}`, albumKey), false)

	for trackPosition := 1; trackPosition <= untaggedAlbumTrackCount; trackPosition++ {
		environment.reindex(t, untaggedAlbumFileName(trackPosition), untaggedAlbumTags(trackPosition))
	}

	if got := environment.sortedNamesOfGenre(t, "rock"); !slices.Equal(got, untaggedAlbumFileNames()) {
		t.Fatalf("tracks of rock after reindex = %v", got)
	}
}
