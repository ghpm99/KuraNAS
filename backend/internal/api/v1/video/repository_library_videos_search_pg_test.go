package video

import (
	"database/sql"
	"slices"
	"testing"

	"nas-go/api/internal/testutil"
)

func searchLibraryVideoNames(t *testing.T, repository *Repository, searchQuery string) []string {
	t.Helper()
	page, err := repository.ListLibraryVideos(LibraryVideosRequest{
		Ordering:    LibraryVideoOrdering{Sort: LibraryVideoSortName, Direction: SortDirectionAscending},
		Page:        1,
		PageSize:    20,
		SearchQuery: searchQuery,
	})
	if err != nil {
		t.Fatalf("search %q: %v", searchQuery, err)
	}
	return namesOf(page.Items)
}

func TestLibraryVideosSearchFoldsAccentsAndCase_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	seedVideoFiles(t, repository, []seededVideoFile{
		{name: "Vídeo de Férias.mp4", folder: "/data/misc"},
		{name: "clip.mkv", folder: "/data/Férias 2024"},
		{name: "100%.mp4", folder: "/data/misc"},
		{name: "1000.mp4", folder: "/data/misc"},
		{name: "a_b.mp4", folder: "/data/misc"},
		{name: "axb.mp4", folder: "/data/misc"},
		{name: "Férias removido.mp4", folder: "/data/misc", deleted: true},
	})

	testCases := []struct {
		description   string
		searchQuery   string
		expectedNames []string
	}{
		{"unaccented term finds accented name", "video", []string{"Vídeo de Férias.mp4"}},
		{"unaccented term finds accented name and folder", "ferias", []string{"clip.mkv", "Vídeo de Férias.mp4"}},
		{"accented uppercase term finds accented name", "VÍDEO", []string{"Vídeo de Férias.mp4"}},
		{"uppercase term", "FERIAS", []string{"clip.mkv", "Vídeo de Férias.mp4"}},
		{"parent folder only", "2024", []string{"clip.mkv"}},
		{"percent is literal", "%", []string{"100%.mp4"}},
		{"underscore is literal", "a_b", []string{"a_b.mp4"}},
		{"matches name containing the extension text", "mkv", []string{"clip.mkv"}},
	}

	for _, testCase := range testCases {
		names := searchLibraryVideoNames(t, repository, testCase.searchQuery)
		if !slices.Equal(names, testCase.expectedNames) && !(len(names) == 0 && len(testCase.expectedNames) == 0) {
			t.Fatalf("%s: search %q want %v, got %v", testCase.description, testCase.searchQuery, testCase.expectedNames, names)
		}
	}
}

func TestParentPathFoldTrigramIndexExists_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)

	var hasExtension, hasIndex bool
	err := repository.DbContext.QueryTx(func(tx *sql.Tx) error {
		if scanErr := tx.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm')").Scan(&hasExtension); scanErr != nil {
			return scanErr
		}
		return tx.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'home_file_active_parent_path_fold_trigram')").Scan(&hasIndex)
	})
	if err != nil {
		t.Fatalf("inspect index: %v", err)
	}
	if hasExtension && !hasIndex {
		t.Fatalf("pg_trgm is installed but the parent_path trigram index is missing")
	}
}
