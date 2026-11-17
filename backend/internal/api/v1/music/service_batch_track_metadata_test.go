package music

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"
)

func decodeBatchRequest(t *testing.T, body string) BatchTracksMetadataRequest {
	t.Helper()
	var request BatchTracksMetadataRequest
	if err := json.Unmarshal([]byte(body), &request); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return request
}

func batchTestService(store *trackMetadataStoreMock) *Service {
	return &Service{Repository: &musicRepoMock{}, TrackMetadataStore: store}
}

func assertRefusedField(t *testing.T, err error, wantField string, wantMessageKey string) {
	t.Helper()
	var validationError *fieldValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("expected field validation error, got %v", err)
	}
	if validationError.Field != wantField || validationError.MessageKey != wantMessageKey {
		t.Fatalf("refused %q with %q, want %q with %q", validationError.Field, validationError.MessageKey, wantField, wantMessageKey)
	}
}

func TestServiceUpdateTracksMetadataRequiresExactlyOneSelector(t *testing.T) {
	requestBodies := map[string]string{
		"none":                 `{"fields": {"genres": ["Rock"]}}`,
		"file ids and album":   `{"fields": {"genres": ["Rock"]}, "file_ids": [1], "album_key": "a"}`,
		"album and folder":     `{"fields": {"genres": ["Rock"]}, "album_key": "a", "folder_key": "/m"}`,
		"all three":            `{"fields": {"genres": ["Rock"]}, "file_ids": [1], "album_key": "a", "folder_key": "/m"}`,
		"empty album key only": `{"fields": {"genres": ["Rock"]}, "album_key": ""}`,
	}
	for caseName, requestBody := range requestBodies {
		store := &trackMetadataStoreMock{}

		_, err := batchTestService(store).UpdateTracksMetadata(decodeBatchRequest(t, requestBody), false)

		assertRefusedField(t, err, "selector", "ERROR_MUSIC_METADATA_SELECTOR_INVALID")
		if store.applyCallCount != 0 {
			t.Fatalf("%s: nothing may be written", caseName)
		}
	}
}

func TestServiceUpdateTracksMetadataLimitsFileIDsToFiveHundred(t *testing.T) {
	tooManyFileIDs := make([]int, maxBatchTrackFileIDs+1)
	for position := range tooManyFileIDs {
		tooManyFileIDs[position] = position + 1
	}
	tooManyBody, _ := json.Marshal(map[string]any{"fields": map[string]any{"year": 1999}, "file_ids": tooManyFileIDs})

	for _, requestBody := range []string{`{"fields": {"year": 1999}, "file_ids": []}`, string(tooManyBody)} {
		_, err := batchTestService(&trackMetadataStoreMock{}).UpdateTracksMetadata(decodeBatchRequest(t, requestBody), false)

		assertRefusedField(t, err, "file_ids", "ERROR_MUSIC_METADATA_FILE_IDS_INVALID")
	}
}

func TestServiceUpdateTracksMetadataAcceptsExactlyFiveHundredFileIDs(t *testing.T) {
	fileIDs := make([]int, maxBatchTrackFileIDs)
	for position := range fileIDs {
		fileIDs[position] = position + 1
	}
	store := &trackMetadataStoreMock{indexedFileIDs: fileIDs}
	requestBody, _ := json.Marshal(map[string]any{"fields": map[string]any{"year": 1999}, "file_ids": fileIDs})

	result, err := batchTestService(store).UpdateTracksMetadata(decodeBatchRequest(t, string(requestBody)), true)
	if err != nil {
		t.Fatalf("UpdateTracksMetadata: %v", err)
	}
	if result.Matched != maxBatchTrackFileIDs {
		t.Fatalf("matched = %d", result.Matched)
	}
}

func TestServiceUpdateTracksMetadataRequiresAtLeastOneField(t *testing.T) {
	for _, requestBody := range []string{`{"file_ids": [1]}`, `{"fields": {}, "file_ids": [1]}`} {
		_, err := batchTestService(&trackMetadataStoreMock{}).UpdateTracksMetadata(decodeBatchRequest(t, requestBody), false)

		assertRefusedField(t, err, "fields", "ERROR_MUSIC_METADATA_FIELDS_EMPTY")
	}
}

func TestServiceUpdateTracksMetadataNamesTheInvalidFieldBeforeTouchingStorage(t *testing.T) {
	store := &trackMetadataStoreMock{indexedFileIDs: []int{1}}

	_, err := batchTestService(store).UpdateTracksMetadata(decodeBatchRequest(t, `{"fields": {"year": 12}, "file_ids": [1]}`), false)

	assertRefusedField(t, err, "year", "ERROR_MUSIC_METADATA_YEAR_INVALID")
	if store.applyCallCount != 0 || store.filteredFileIDs != nil {
		t.Fatal("storage must not be touched for an invalid request")
	}
}

func TestServiceUpdateTracksMetadataDryRunCountsWithoutWriting(t *testing.T) {
	store := &trackMetadataStoreMock{albumFileIDs: []int{4, 5, 6}}

	result, err := batchTestService(store).UpdateTracksMetadata(decodeBatchRequest(t, `{"fields": {"genres": ["Rock"]}, "album_key": "album-x"}`), true)
	if err != nil {
		t.Fatalf("UpdateTracksMetadata: %v", err)
	}

	if result.Matched != 3 || result.Updated != 0 || result.Failed == nil || len(result.Failed) != 0 {
		t.Fatalf("result = %+v", result)
	}
	if store.applyCallCount != 0 {
		t.Fatal("dry run must not write")
	}
	if store.requestedAlbumKey != "album-x" {
		t.Fatalf("album key = %q", store.requestedAlbumKey)
	}
}

func TestServiceUpdateTracksMetadataAppliesPatchOnceToEveryMatchedTrack(t *testing.T) {
	store := &trackMetadataStoreMock{folderFileIDs: []int{8, 9}}

	result, err := batchTestService(store).UpdateTracksMetadata(decodeBatchRequest(t, `{"fields": {"genres": ["Rock"]}, "folder_key": "/m/rock"}`), false)
	if err != nil {
		t.Fatalf("UpdateTracksMetadata: %v", err)
	}

	if result.Matched != 2 || result.Updated != 2 || result.Failed == nil || len(result.Failed) != 0 {
		t.Fatalf("result = %+v", result)
	}
	if store.applyCallCount != 1 || !slices.Equal(store.appliedFileIDs, []int{8, 9}) {
		t.Fatalf("apply calls=%d ids=%v", store.applyCallCount, store.appliedFileIDs)
	}
	if !store.appliedPatch.Genres.IsSet || !slices.Equal(store.appliedPatch.Genres.Values, []string{"Rock"}) {
		t.Fatalf("applied patch = %+v", store.appliedPatch)
	}
	if store.appliedLabel == "" || store.requestedFolderPath != "/m/rock" {
		t.Fatalf("label=%q folder=%q", store.appliedLabel, store.requestedFolderPath)
	}
}

func TestServiceUpdateTracksMetadataReportsUnknownFileIDsAsFailedAndWritesTheRest(t *testing.T) {
	store := &trackMetadataStoreMock{indexedFileIDs: []int{3, 7}}

	result, err := batchTestService(store).UpdateTracksMetadata(decodeBatchRequest(t, `{"fields": {"year": 2001}, "file_ids": [7, 999, 3, 555]}`), false)
	if err != nil {
		t.Fatalf("UpdateTracksMetadata: %v", err)
	}

	if result.Matched != 2 || result.Updated != 2 {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Failed) != 2 || result.Failed[0].FileID != 999 || result.Failed[1].FileID != 555 {
		t.Fatalf("failed = %+v", result.Failed)
	}
	for _, failure := range result.Failed {
		if failure.Error == "" {
			t.Fatalf("failure without message: %+v", failure)
		}
	}
	if !slices.Equal(store.appliedFileIDs, []int{3, 7}) {
		t.Fatalf("applied ids = %v", store.appliedFileIDs)
	}
}

func TestServiceUpdateTracksMetadataWritesNothingWhenNoTrackMatches(t *testing.T) {
	store := &trackMetadataStoreMock{}

	result, err := batchTestService(store).UpdateTracksMetadata(decodeBatchRequest(t, `{"fields": {"year": 2001}, "file_ids": [999]}`), false)
	if err != nil {
		t.Fatalf("UpdateTracksMetadata: %v", err)
	}

	if result.Matched != 0 || result.Updated != 0 || len(result.Failed) != 1 || store.applyCallCount != 0 {
		t.Fatalf("result = %+v apply=%d", result, store.applyCallCount)
	}
}

func TestServiceUpdateTracksMetadataReturnsTheTransactionErrorInsteadOfPerItemFailures(t *testing.T) {
	applyError := errors.New("transaction failed")
	store := &trackMetadataStoreMock{albumFileIDs: []int{1, 2}, applyError: applyError}

	_, err := batchTestService(store).UpdateTracksMetadata(decodeBatchRequest(t, `{"fields": {"year": 2001}, "album_key": "a"}`), false)

	if !errors.Is(err, applyError) {
		t.Fatalf("expected apply error, got %v", err)
	}
}

func TestServiceUpdateTracksMetadataPropagatesSelectionError(t *testing.T) {
	selectionError := errors.New("select failed")
	store := &trackMetadataStoreMock{selectionError: selectionError}

	_, err := batchTestService(store).UpdateTracksMetadata(decodeBatchRequest(t, `{"fields": {"year": 2001}, "folder_key": "/m"}`), false)

	if !errors.Is(err, selectionError) || store.applyCallCount != 0 {
		t.Fatalf("err=%v apply=%d", err, store.applyCallCount)
	}
}

func jsonIntList(numbers []int) string {
	encodedNumbers, _ := json.Marshal(numbers)
	return string(encodedNumbers)
}
