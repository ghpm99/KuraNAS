package video

import (
	"database/sql"
	"slices"
	"testing"

	"nas-go/api/internal/testutil"
)

type seededVideoTechnicalInfo struct {
	height          int
	durationSeconds string
}

func seedFilterableLibrary(t *testing.T) *Repository {
	t.Helper()
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	ids := seedVideoFiles(t, repository, []seededVideoFile{
		{name: "sd-short.mp4", folder: "/data"},
		{name: "hd-medium-low.mkv", folder: "/data"},
		{name: "hd-top.mp4", folder: "/data"},
		{name: "fhd-medium-high.mp4", folder: "/data"},
		{name: "fhd-long.mkv", folder: "/data"},
		{name: "uhd-long.mkv", folder: "/data"},
		{name: "no-metadata.mkv", folder: "/data"},
	})

	technicalInfoByName := map[string]seededVideoTechnicalInfo{
		"sd-short.mp4":        {height: 719, durationSeconds: "1199.9"},
		"hd-medium-low.mkv":   {height: 720, durationSeconds: "1200"},
		"hd-top.mp4":          {height: 1079, durationSeconds: "3599"},
		"fhd-medium-high.mp4": {height: 1080, durationSeconds: "3599.5"},
		"fhd-long.mkv":        {height: 2159, durationSeconds: "3600"},
		"uhd-long.mkv":        {height: 2160, durationSeconds: "7200"},
	}
	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE video_metadata RESTART IDENTITY`); err != nil {
			return err
		}
		for name, info := range technicalInfoByName {
			if _, err := tx.Exec(`INSERT INTO video_metadata (file_id, path, duration, height) VALUES ($1, $2, $3, $4)`, ids[name], "/p/"+name, info.durationSeconds, info.height); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}
	return repository
}

func listFilteredNames(t *testing.T, repository *Repository, searchQuery string, rawFormats string, rawResolution string, rawDuration string) []string {
	t.Helper()
	filters, isValid := ParseLibraryVideoFilters(rawFormats, rawResolution, rawDuration)
	if !isValid {
		t.Fatalf("invalid filters %q %q %q", rawFormats, rawResolution, rawDuration)
	}
	page, err := repository.ListLibraryVideos(LibraryVideosRequest{
		Ordering:    LibraryVideoOrdering{Sort: LibraryVideoSortName, Direction: SortDirectionAscending},
		Page:        1,
		PageSize:    50,
		SearchQuery: searchQuery,
		Filters:     filters,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	names := namesOf(page.Items)
	slices.Sort(names)
	return names
}

func TestLibraryVideosFilters_Postgres(t *testing.T) {
	repository := seedFilterableLibrary(t)

	testCases := []struct {
		label         string
		searchQuery   string
		formats       string
		resolution    string
		duration      string
		expectedNames []string
	}{
		{"no filters includes video without metadata", "", "", "", "", []string{"fhd-long.mkv", "fhd-medium-high.mp4", "hd-medium-low.mkv", "hd-top.mp4", "no-metadata.mkv", "sd-short.mp4", "uhd-long.mkv"}},
		{"format only keeps video without metadata", "", "mkv", "", "", []string{"fhd-long.mkv", "hd-medium-low.mkv", "no-metadata.mkv", "uhd-long.mkv"}},
		{"format with several values", "", "mp4,mkv", "", "", []string{"fhd-long.mkv", "fhd-medium-high.mp4", "hd-medium-low.mkv", "hd-top.mp4", "no-metadata.mkv", "sd-short.mp4", "uhd-long.mkv"}},
		{"resolution sd", "", "", "sd", "", []string{"sd-short.mp4"}},
		{"resolution hd boundaries", "", "", "hd", "", []string{"hd-medium-low.mkv", "hd-top.mp4"}},
		{"resolution fhd boundaries", "", "", "fhd", "", []string{"fhd-long.mkv", "fhd-medium-high.mp4"}},
		{"resolution uhd", "", "", "uhd", "", []string{"uhd-long.mkv"}},
		{"duration short", "", "", "", "short", []string{"sd-short.mp4"}},
		{"duration medium boundaries", "", "", "", "medium", []string{"fhd-medium-high.mp4", "hd-medium-low.mkv", "hd-top.mp4"}},
		{"duration long boundaries", "", "", "", "long", []string{"fhd-long.mkv", "uhd-long.mkv"}},
		{"all filters combined", "", "mkv", "fhd", "long", []string{"fhd-long.mkv"}},
		{"resolution and duration combined", "", "", "hd", "medium", []string{"hd-medium-low.mkv", "hd-top.mp4"}},
		{"filter with search", "long", "", "", "long", []string{"fhd-long.mkv", "uhd-long.mkv"}},
		{"filter with search narrowing", "uhd", "", "", "long", []string{"uhd-long.mkv"}},
		{"filters with search yielding nothing", "sd", "", "uhd", "", []string{}},
	}

	for _, testCase := range testCases {
		names := listFilteredNames(t, repository, testCase.searchQuery, testCase.formats, testCase.resolution, testCase.duration)
		if !slices.Equal(names, testCase.expectedNames) {
			t.Fatalf("%s: want %v, got %v", testCase.label, testCase.expectedNames, names)
		}
	}
}
