package video

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func posterNames(candidates []PosterCandidateModel) []string {
	names := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		names = append(names, candidate.Name)
	}
	return names
}

func seedPlaylistWithVideos(t *testing.T, repository *Repository, idsByName map[string]int, videoNames ...string) int {
	t.Helper()
	var playlistID int
	seedErr := repository.DbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE video_playlist RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		err := tx.QueryRow(
			`INSERT INTO video_playlist (type, source_path, name, is_auto, classification, is_hidden)
			 VALUES ('series', 'series:show', 'Show', TRUE, 'series', FALSE) RETURNING id`,
		).Scan(&playlistID)
		if err != nil {
			return err
		}
		for position, videoName := range videoNames {
			if _, err := tx.Exec(
				`INSERT INTO video_playlist_item (playlist_id, video_id, order_index, source_kind) VALUES ($1, $2, $3, 'auto')`,
				playlistID, idsByName[videoName], position,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed playlist: %v", seedErr)
	}
	return playlistID
}

func TestListPosterCandidatesInFolderIsCaseInsensitiveAndSkipsDeletedAndNonImages_Postgres(t *testing.T) {
	repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_video_it"))
	ids := seedVideoFiles(t, repository, []seededVideoFile{
		{name: "POSTER.JPG", folder: "/data/Show"},
		{name: "folder.png", folder: "/data/Show"},
		{name: "cover.webp", folder: "/data/Show", deleted: true},
		{name: "notes.txt", folder: "/data/Show"},
		{name: "episode.mkv", folder: "/data/Show"},
		{name: "poster.jpg", folder: "/data/Other"},
	})

	candidates, err := repository.ListPosterCandidatesInFolder("/data/Show")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(candidates) != 2 || candidates[0].ID != ids["POSTER.JPG"] || candidates[1].ID != ids["folder.png"] {
		t.Fatalf("candidates = %v, want POSTER.JPG and folder.png", posterNames(candidates))
	}

	poster, isFound := pickSeriesPoster(candidates)
	if !isFound || poster.ID != ids["POSTER.JPG"] {
		t.Fatalf("poster = %+v found=%v, want POSTER.JPG", poster, isFound)
	}
}

func TestCountVideosInFolderIgnoresDeletedAndImages_Postgres(t *testing.T) {
	repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_video_it"))
	seedVideoFiles(t, repository, []seededVideoFile{
		{name: "a.mkv", folder: "/data/Own"},
		{name: "gone.mkv", folder: "/data/Own", deleted: true},
		{name: "folder.jpg", folder: "/data/Own"},
		{name: "b.mkv", folder: "/data/Shared"},
		{name: "c.MP4", folder: "/data/Shared"},
	})

	ownCount, err := repository.CountVideosInFolder("/data/Own")
	if err != nil || ownCount != 1 {
		t.Fatalf("own count = %d err = %v, want 1", ownCount, err)
	}
	sharedCount, err := repository.CountVideosInFolder("/data/Shared")
	if err != nil || sharedCount != 2 {
		t.Fatalf("shared count = %d err = %v, want 2", sharedCount, err)
	}
}

func TestMoviePosterRuleOnSeededFolders_Postgres(t *testing.T) {
	repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_video_it"))
	ids := seedVideoFiles(t, repository, []seededVideoFile{
		{name: "Loose A.mkv", folder: "/data/Loose"},
		{name: "Loose B.mkv", folder: "/data/Loose"},
		{name: "Folder.JPG", folder: "/data/Loose"},
		{name: "Own.mkv", folder: "/data/Own"},
		{name: "folder.jpg", folder: "/data/Own"},
		{name: "Kodi.mkv", folder: "/data/Kodi"},
		{name: "Other.mkv", folder: "/data/Kodi"},
		{name: "kodi-poster.png", folder: "/data/Kodi"},
	})
	service := &Service{Repository: repository}

	if _, err := service.FindMoviePosterFileID(ids["Loose A.mkv"]); err != ErrPosterNotFound {
		t.Fatalf("loose movie: err = %v, want ErrPosterNotFound", err)
	}
	if _, err := service.FindMoviePosterFileID(ids["Other.mkv"]); err != ErrPosterNotFound {
		t.Fatalf("Other.mkv: err = %v, want ErrPosterNotFound", err)
	}
	ownPosterID, err := service.FindMoviePosterFileID(ids["Own.mkv"])
	if err != nil || ownPosterID != ids["folder.jpg"] {
		t.Fatalf("own folder movie: poster = %d err = %v", ownPosterID, err)
	}
	kodiPosterID, err := service.FindMoviePosterFileID(ids["Kodi.mkv"])
	if err != nil || kodiPosterID != ids["kodi-poster.png"] {
		t.Fatalf("kodi movie: poster = %d err = %v", kodiPosterID, err)
	}
}

func TestPlaylistDominantFolderPicksFolderWithMostItems_Postgres(t *testing.T) {
	repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_video_it"))
	ids := seedVideoFiles(t, repository, []seededVideoFile{
		{name: "e1.mkv", folder: "/data/Show/Season 1"},
		{name: "e2.mkv", folder: "/data/Show/Season 1"},
		{name: "e3.mkv", folder: "/data/Show/Season 2"},
		{name: "gone.mkv", folder: "/data/Show/Season 2", deleted: true},
		{name: "gone2.mkv", folder: "/data/Show/Season 2", deleted: true},
	})
	playlistID := seedPlaylistWithVideos(t, repository, ids, "e1.mkv", "e2.mkv", "e3.mkv", "gone.mkv", "gone2.mkv")

	folderPath, err := repository.GetPlaylistDominantFolder(playlistID)
	if err != nil || folderPath != "/data/Show/Season 1" {
		t.Fatalf("folder = %q err = %v, want Season 1", folderPath, err)
	}
}
