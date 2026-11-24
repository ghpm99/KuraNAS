package video

import (
	"fmt"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestRebuildMakesSmallFolderSeriesAppearInSeriesSection_Postgres(t *testing.T) {
	for _, episodeCount := range []int{2, 6, 10} {
		t.Run(fmt.Sprintf("%d_episodes", episodeCount), func(t *testing.T) {
			repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_video_it"))
			service := NewService(repository).(*Service)
			seeds := []seededVideoFile{}
			for episode := 1; episode <= episodeCount; episode++ {
				seeds = append(seeds, seededVideoFile{name: fmt.Sprintf("Dark Matter S01E%02d.mkv", episode), folder: "/data/Dark Matter"})
			}
			seeds = append(seeds,
				seededVideoFile{name: "Home Video 1.mp4", folder: "/data/Other0"},
				seededVideoFile{name: "Home Video 2.mp4", folder: "/data/Other1"},
				seededVideoFile{name: "Home Video 3.mp4", folder: "/data/Other2"},
			)
			seedVideoFiles(t, repository, seeds)

			if err := service.RebuildSmartPlaylists(); err != nil {
				t.Fatalf("rebuild: %v", err)
			}

			seriesPage, err := service.GetPlaylistsBySection(PlaylistSectionRequest{Section: PlaylistSectionSeries, Page: 1, PageSize: 20})
			if err != nil {
				t.Fatalf("series section: %v", err)
			}
			for _, playlist := range seriesPage.Items {
				if playlist.SourcePath == "/data/Dark Matter" || playlist.Name == "Dark Matter" {
					return
				}
			}
			t.Fatalf("series of %d episodes missing from series section: %+v", episodeCount, seriesPage.Items)
		})
	}
}
