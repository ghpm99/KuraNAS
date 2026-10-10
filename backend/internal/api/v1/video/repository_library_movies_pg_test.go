package video

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
)

func namesOf(videos []VideoFileModel) []string {
	names := make([]string, 0, len(videos))
	for _, video := range videos {
		names = append(names, video.Name)
	}
	return names
}

func TestLibraryMoviesListOnlyActiveMovieClassifiedVideosInEachSortOrder_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	ids := seedVideoFiles(t, repository, []seededVideoFile{
		{name: "b-film.mkv", folder: "/data/Random"},
		{name: "a-film.mkv", folder: "/data/Other"},
		{name: "c-film.mp4", folder: "/data/Filmes"},
		{name: "episode.mkv", folder: "/data/Random"},
		{name: "unclassified.mkv", folder: "/data/Random"},
		{name: "gone-film.mkv", folder: "/data/Random", deleted: true},
		{name: "notes.txt", folder: "/data/Random"},
	})

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE video_metadata RESTART IDENTITY`); err != nil {
			return err
		}
		classifications := map[string]string{
			"b-film.mkv": "movie", "a-film.mkv": "movie", "c-film.mp4": "movie",
			"episode.mkv": "series", "gone-film.mkv": "movie", "notes.txt": "movie",
		}
		for name, classification := range classifications {
			if _, err := tx.Exec(`INSERT INTO video_metadata (file_id, path, classification) VALUES ($1, $2, $3)`, ids[name], "/p/"+name, classification); err != nil {
				return err
			}
		}
		base := time.Now()
		recency := map[string]time.Duration{"a-film.mkv": 3 * time.Hour, "c-film.mp4": 2 * time.Hour, "b-film.mkv": time.Hour}
		for name, age := range recency {
			if _, err := tx.Exec(`UPDATE home_file SET updated_at = $2 WHERE id = $1`, ids[name], base.Add(-age)); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	assertNames := func(sort LibraryMovieSort, limit int, offset int, expected []string) {
		t.Helper()
		movies, err := repository.ListLibraryMovies(sort, limit, offset)
		if err != nil {
			t.Fatalf("list %s: %v", sort, err)
		}
		names := namesOf(movies)
		if len(names) != len(expected) {
			t.Fatalf("sort %s: want %v, got %v", sort, expected, names)
		}
		for position := range expected {
			if names[position] != expected[position] {
				t.Fatalf("sort %s: want %v, got %v", sort, expected, names)
			}
		}
	}

	assertNames(LibraryMovieSortName, 10, 0, []string{"a-film.mkv", "b-film.mkv", "c-film.mp4"})
	assertNames(LibraryMovieSortRecent, 10, 0, []string{"b-film.mkv", "c-film.mp4", "a-film.mkv"})
	assertNames(LibraryMovieSortName, 2, 2, []string{"c-film.mp4"})
}
