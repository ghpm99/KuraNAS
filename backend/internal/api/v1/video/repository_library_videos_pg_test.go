package video

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
)

func TestLibraryVideosOrderBySortAndDirection_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	ids := seedVideoFiles(t, repository, []seededVideoFile{
		{name: "b-clip.mkv", folder: "/data"},
		{name: "A-clip.mkv", folder: "/data"},
		{name: "c-clip.mp4", folder: "/data"},
		{name: "d-clip.mp4", folder: "/data"},
		{name: "gone.mkv", folder: "/data", deleted: true},
	})

	baseTime := time.Now()
	sizes := map[string]int{"b-clip.mkv": 200, "A-clip.mkv": 300, "c-clip.mp4": 100, "d-clip.mp4": 400}
	ageHours := map[string]time.Duration{"b-clip.mkv": 1, "A-clip.mkv": 3, "c-clip.mp4": 2, "d-clip.mp4": 4}
	durations := map[string]string{"b-clip.mkv": "60.5", "A-clip.mkv": "7200", "c-clip.mp4": "N/A"}

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE video_metadata RESTART IDENTITY`); err != nil {
			return err
		}
		for name, size := range sizes {
			if _, err := tx.Exec(`UPDATE home_file SET size = $2, updated_at = $3 WHERE id = $1`, ids[name], size, baseTime.Add(-ageHours[name]*time.Hour)); err != nil {
				return err
			}
		}
		for name, duration := range durations {
			if _, err := tx.Exec(`INSERT INTO video_metadata (file_id, path, duration) VALUES ($1, $2, $3)`, ids[name], "/p/"+name, duration); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	testCases := []struct {
		sort          LibraryVideoSort
		direction     SortDirection
		expectedNames []string
	}{
		{LibraryVideoSortRecent, SortDirectionDescending, []string{"b-clip.mkv", "c-clip.mp4", "A-clip.mkv", "d-clip.mp4"}},
		{LibraryVideoSortRecent, SortDirectionAscending, []string{"d-clip.mp4", "A-clip.mkv", "c-clip.mp4", "b-clip.mkv"}},
		{LibraryVideoSortName, SortDirectionAscending, []string{"A-clip.mkv", "b-clip.mkv", "c-clip.mp4", "d-clip.mp4"}},
		{LibraryVideoSortName, SortDirectionDescending, []string{"d-clip.mp4", "c-clip.mp4", "b-clip.mkv", "A-clip.mkv"}},
		{LibraryVideoSortSize, SortDirectionDescending, []string{"d-clip.mp4", "A-clip.mkv", "b-clip.mkv", "c-clip.mp4"}},
		{LibraryVideoSortSize, SortDirectionAscending, []string{"c-clip.mp4", "b-clip.mkv", "A-clip.mkv", "d-clip.mp4"}},
		{LibraryVideoSortDuration, SortDirectionDescending, []string{"A-clip.mkv", "b-clip.mkv", "d-clip.mp4", "c-clip.mp4"}},
		{LibraryVideoSortDuration, SortDirectionAscending, []string{"b-clip.mkv", "A-clip.mkv", "c-clip.mp4", "d-clip.mp4"}},
	}

	for _, testCase := range testCases {
		page, err := repository.ListLibraryVideos(LibraryVideosRequest{
			Ordering: LibraryVideoOrdering{Sort: testCase.sort, Direction: testCase.direction},
			Page:     1,
			PageSize: 10,
		})
		if err != nil {
			t.Fatalf("%s %s: %v", testCase.sort, testCase.direction, err)
		}
		names := namesOf(page.Items)
		if len(names) != len(testCase.expectedNames) {
			t.Fatalf("%s %s: want %v, got %v", testCase.sort, testCase.direction, testCase.expectedNames, names)
		}
		for position := range names {
			if names[position] != testCase.expectedNames[position] {
				t.Fatalf("%s %s: want %v, got %v", testCase.sort, testCase.direction, testCase.expectedNames, names)
			}
		}
	}
}
