package video

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

type seededTechnicalMetadata struct {
	path     string
	duration string
	width    int
	height   int
	codec    string
}

func seedTechnicalMetadataFixture(t *testing.T, repository *Repository) (map[string]int, int) {
	t.Helper()
	videoIDsByName := seedVideoFiles(t, repository, []seededVideoFile{
		{name: "full.mkv", folder: "/data/tech"},
		{name: "unparsable.mkv", folder: "/data/tech"},
		{name: "bare.mkv", folder: "/data/tech"},
	})
	playlistID := seedPlaylists(t, repository, []seededPlaylist{{sourcePath: "/tech", isAuto: false}})["/tech"]

	metadataByName := map[string][]seededTechnicalMetadata{
		"full.mkv": {
			{path: "/old/full.mkv", duration: "10", width: 640, height: 360, codec: "mpeg4"},
			{path: "/new/full.mkv", duration: "5400.25", width: 1920, height: 1080, codec: "h264"},
		},
		"unparsable.mkv": {
			{path: "/p/unparsable.mkv", duration: "N/A", width: 1280, height: 720, codec: "hevc"},
		},
	}
	orderedVideoIDs := []int{videoIDsByName["full.mkv"], videoIDsByName["unparsable.mkv"], videoIDsByName["bare.mkv"]}

	seedErr := repository.DbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE video_metadata RESTART IDENTITY`); err != nil {
			return err
		}
		for name, rows := range metadataByName {
			for _, row := range rows {
				if _, err := tx.Exec(
					`INSERT INTO video_metadata (file_id, path, duration, width, height, codec_name, classification) VALUES ($1, $2, $3, $4, $5, $6, 'movie')`,
					videoIDsByName[name], row.path, row.duration, row.width, row.height, row.codec,
				); err != nil {
					return err
				}
			}
		}
		if err := repository.InsertPlaylistItemsWithSource(tx, playlistID, orderedVideoIDs, "manual"); err != nil {
			return err
		}
		for _, videoID := range orderedVideoIDs {
			if _, err := repository.UpsertVideoWatchProgress(tx, VideoWatchProgressModel{
				ClientID: "tv", VideoID: videoID, PositionSeconds: 30, DurationSeconds: 100,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}
	return videoIDsByName, playlistID
}

func assertTechnicalMetadata(t *testing.T, source string, videosByName map[string]VideoFileModel) {
	t.Helper()
	full, ok := videosByName["full.mkv"]
	if !ok {
		t.Fatalf("%s: full.mkv missing", source)
	}
	if !full.DurationSeconds.Valid || full.DurationSeconds.Float64 != 5400.25 ||
		full.Width.Int64 != 1920 || full.Height.Int64 != 1080 || full.VideoCodec.String != "h264" {
		t.Fatalf("%s: full.mkv must carry the latest metadata row, got %+v", source, full)
	}
	if unparsable, ok := videosByName["unparsable.mkv"]; ok {
		if unparsable.DurationSeconds.Valid {
			t.Fatalf("%s: non numeric duration must stay null, got %+v", source, unparsable.DurationSeconds)
		}
		if unparsable.Width.Int64 != 1280 || unparsable.VideoCodec.String != "hevc" {
			t.Fatalf("%s: unparsable.mkv keeps resolution and codec, got %+v", source, unparsable)
		}
	}
	if bare, ok := videosByName["bare.mkv"]; ok {
		if bare.DurationSeconds.Valid || bare.Width.Valid || bare.Height.Valid || bare.VideoCodec.Valid {
			t.Fatalf("%s: file without metadata must have null technical fields, got %+v", source, bare)
		}
	}
}

func videosByNameOf(videos []VideoFileModel) map[string]VideoFileModel {
	videosByName := make(map[string]VideoFileModel, len(videos))
	for _, video := range videos {
		videosByName[video.Name] = video
	}
	return videosByName
}

func TestTechnicalMetadataIsScannedFromEveryVideoListing_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	_, playlistID := seedTechnicalMetadataFixture(t, repository)

	library, err := repository.ListLibraryVideos(LibraryVideosRequest{
		Ordering: LibraryVideoOrdering{Sort: LibraryVideoSortRecent, Direction: SortDirectionDescending},
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("library: %v", err)
	}
	if len(library.Items) != 3 {
		t.Fatalf("library: latest-row join must not duplicate files, got %d", len(library.Items))
	}
	assertTechnicalMetadata(t, "library", videosByNameOf(library.Items))

	folderVideos, err := repository.ListLibraryFolderVideos("/data/tech", 10, 0)
	if err != nil {
		t.Fatalf("folder: %v", err)
	}
	assertTechnicalMetadata(t, "folder", videosByNameOf(folderVideos))

	movies, err := repository.ListLibraryMovies(LibraryMovieSortName, 10, 0)
	if err != nil {
		t.Fatalf("movies: %v", err)
	}
	assertTechnicalMetadata(t, "movies", videosByNameOf(movies))

	playlistPage, err := repository.GetVideoPlaylistItemsPage(playlistID, 10, 0)
	if err != nil {
		t.Fatalf("playlist page: %v", err)
	}
	playlistVideos := make([]VideoFileModel, 0, len(playlistPage))
	for _, item := range playlistPage {
		playlistVideos = append(playlistVideos, item.Video)
	}
	assertTechnicalMetadata(t, "playlist page", videosByNameOf(playlistVideos))

	detailedItems, err := repository.GetVideoPlaylistItemsDetailed(playlistID)
	if err != nil {
		t.Fatalf("playlist detailed: %v", err)
	}
	detailedVideos := make([]VideoFileModel, 0, len(detailedItems))
	for _, item := range detailedItems {
		detailedVideos = append(detailedVideos, item.Video)
	}
	assertTechnicalMetadata(t, "playlist detailed", videosByNameOf(detailedVideos))

	continueWatching, err := repository.GetContinueWatchingVideos("tv", 10)
	if err != nil {
		t.Fatalf("continue watching: %v", err)
	}
	continueVideos := make([]VideoFileModel, 0, len(continueWatching))
	for _, entry := range continueWatching {
		continueVideos = append(continueVideos, entry.VideoFileModel)
	}
	assertTechnicalMetadata(t, "continue watching", videosByNameOf(continueVideos))
}
