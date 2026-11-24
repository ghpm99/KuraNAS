package video

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"nas-go/api/internal/testutil"
)

type seriesFixture struct {
	repository *Repository
	service    *Service
	videoIDs   map[string]int

	firstFileName string
}

const unrelatedVideoCount = 12

func newSeriesFixture(t *testing.T, folder string, fileNames []string) *seriesFixture {
	t.Helper()
	repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_video_it"))
	service := NewService(repository).(*Service)

	seeds := make([]seededVideoFile, 0, len(fileNames)+unrelatedVideoCount)
	for _, name := range fileNames {
		seeds = append(seeds, seededVideoFile{name: name, folder: folder})
	}
	for unrelatedNumber := 1; unrelatedNumber <= unrelatedVideoCount; unrelatedNumber++ {
		seeds = append(seeds, seededVideoFile{name: fmt.Sprintf("Home Video %d.mp4", unrelatedNumber), folder: fmt.Sprintf("/data/Other%d", unrelatedNumber%3)})
	}
	videoIDs := seedVideoFiles(t, repository, seeds)
	if err := service.RebuildSmartPlaylists(); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	return &seriesFixture{repository: repository, service: service, videoIDs: videoIDs, firstFileName: fileNames[0]}
}

func (fixture *seriesFixture) seriesPlaylistID(t *testing.T) int {
	t.Helper()
	var playlistID int
	err := fixture.repository.DbContext.GetDatabase().QueryRow(
		`SELECT p.id FROM video_playlist p
		 JOIN video_playlist_item i ON i.playlist_id = p.id
		 WHERE p.is_auto AND (p.type = 'series' OR p.classification IN ('series', 'anime')) AND i.video_id = $1
		 ORDER BY p.id LIMIT 1`, fixture.videoIDs[fixture.firstFileName]).Scan(&playlistID)
	if err != nil {
		t.Fatalf("series playlist not found: %v", err)
	}
	return playlistID
}

func (fixture *seriesFixture) orderedNames(t *testing.T, playlistID int) []string {
	t.Helper()
	page, err := fixture.service.GetPlaylistItemsPage("tv", playlistID, nil, 1, 500)
	if err != nil {
		t.Fatalf("items page: %v", err)
	}
	names := make([]string, 0, len(page.Items))
	for _, item := range page.Items {
		names = append(names, item.Video.Name)
	}
	return names
}

func (fixture *seriesFixture) overrideEpisode(t *testing.T, fileName string, season any, episode any) {
	t.Helper()
	_, err := fixture.repository.DbContext.GetDatabase().Exec(
		`INSERT INTO video_manual_override (file_id, season, episode) VALUES ($1, $2, $3)
		 ON CONFLICT (file_id) DO UPDATE SET season = EXCLUDED.season, episode = EXCLUDED.episode`,
		fixture.videoIDs[fileName], season, episode)
	if err != nil {
		t.Fatalf("override %s: %v", fileName, err)
	}
}

func shuffledEpisodeNames(prefix string, count int) []string {
	names := make([]string, count)
	for position := range names {
		names[position] = fmt.Sprintf("%s%d.mkv", prefix, position+1)
	}
	rand.New(rand.NewSource(7)).Shuffle(len(names), func(left, right int) {
		names[left], names[right] = names[right], names[left]
	})
	return names
}

func TestRebuildOrdersTwentyFourEpisodesNaturally_Postgres(t *testing.T) {
	fixture := newSeriesFixture(t, "/data/Show", shuffledEpisodeNames("Ep ", 24))
	playlistID := fixture.seriesPlaylistID(t)

	page, err := fixture.service.GetPlaylistItemsPage("tv", playlistID, nil, 1, 50)
	if err != nil {
		t.Fatalf("items page: %v", err)
	}
	if len(page.Items) != 24 {
		t.Fatalf("items = %d, want 24", len(page.Items))
	}
	for position, item := range page.Items {
		wantName := fmt.Sprintf("Ep %d.mkv", position+1)
		if item.Video.Name != wantName || item.OrderIndex != position {
			t.Fatalf("position %d = %s/%d, want %s", position, item.Video.Name, item.OrderIndex, wantName)
		}
		if item.Season == nil || *item.Season != 1 || item.Episode == nil || *item.Episode != position+1 {
			t.Fatalf("position %d numbering = %v/%v", position, item.Season, item.Episode)
		}
	}
}

func TestRebuildPlacesEpisodeTenAfterEpisodeTwo_Postgres(t *testing.T) {
	fixture := newSeriesFixture(t, "/data/Show", shuffledEpisodeNames("Ep ", 16))
	playlistID := fixture.seriesPlaylistID(t)

	want := []string{"Ep 1.mkv", "Ep 2.mkv", "Ep 3.mkv", "Ep 4.mkv", "Ep 5.mkv", "Ep 6.mkv", "Ep 7.mkv", "Ep 8.mkv", "Ep 9.mkv", "Ep 10.mkv", "Ep 11.mkv", "Ep 12.mkv", "Ep 13.mkv", "Ep 14.mkv", "Ep 15.mkv", "Ep 16.mkv"}
	if got := fixture.orderedNames(t, playlistID); !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestRebuildOrdersBySeasonBeforeEpisode_Postgres(t *testing.T) {
	shuffledNames := shuffledEpisodeNames("", 15)
	fileNames := make([]string, 0, len(shuffledNames))
	want := []string{}
	for _, seasonNumber := range []int{1, 2} {
		seasonEpisodeCount := 10
		if seasonNumber == 2 {
			seasonEpisodeCount = 5
		}
		for episodeNumber := 1; episodeNumber <= seasonEpisodeCount; episodeNumber++ {
			want = append(want, fmt.Sprintf("Show S%02dE%02d.mkv", seasonNumber, episodeNumber))
		}
	}
	for _, shuffledName := range shuffledNames {
		var position int
		fmt.Sscanf(shuffledName, "%d.mkv", &position)
		fileNames = append(fileNames, want[position-1])
	}
	fixture := newSeriesFixture(t, "/data/Show", fileNames)
	playlistID := fixture.seriesPlaylistID(t)

	if got := fixture.orderedNames(t, playlistID); !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestOverrideChangesOrderAfterRenumberAndSurvivesRebuild_Postgres(t *testing.T) {
	fixture := newSeriesFixture(t, "/data/Show", shuffledEpisodeNames("Ep ", 16))
	playlistID := fixture.seriesPlaylistID(t)

	fixture.overrideEpisode(t, "Ep 2.mkv", nil, 30)
	fixture.overrideEpisode(t, "Ep 5.mkv", 2, 1)
	if err := fixture.service.RenumberSeriesPlaylist(playlistID); err != nil {
		t.Fatalf("renumber: %v", err)
	}

	want := []string{"Ep 1.mkv", "Ep 3.mkv", "Ep 4.mkv", "Ep 6.mkv", "Ep 7.mkv", "Ep 8.mkv", "Ep 9.mkv", "Ep 10.mkv", "Ep 11.mkv", "Ep 12.mkv", "Ep 13.mkv", "Ep 14.mkv", "Ep 15.mkv", "Ep 16.mkv", "Ep 2.mkv", "Ep 5.mkv"}
	if got := fixture.orderedNames(t, playlistID); !reflect.DeepEqual(got, want) {
		t.Fatalf("order after renumber = %v, want %v", got, want)
	}

	page, err := fixture.service.GetPlaylistItemsPage("tv", playlistID, nil, 1, 50)
	if err != nil {
		t.Fatalf("items page: %v", err)
	}
	effective := map[string][2]int{}
	for _, item := range page.Items {
		effective[item.Video.Name] = [2]int{*item.Season, *item.Episode}
	}
	if effective["Ep 2.mkv"] != [2]int{1, 30} || effective["Ep 5.mkv"] != [2]int{2, 1} || effective["Ep 3.mkv"] != [2]int{1, 3} {
		t.Fatalf("effective numbering = %v", effective)
	}

	if err := fixture.service.RebuildSmartPlaylists(); err != nil {
		t.Fatalf("second rebuild: %v", err)
	}
	if got := fixture.orderedNames(t, fixture.seriesPlaylistID(t)); !reflect.DeepEqual(got, want) {
		t.Fatalf("order after rebuild = %v, want %v", got, want)
	}
	assertOrderIndexesContiguous(t, fixture.repository, fixture.seriesPlaylistID(t), len(want))
}

func TestRenumberSeriesPlaylistRejectsPlaylistsThatAreNotAutomaticSeries_Postgres(t *testing.T) {
	repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_video_it"))
	service := NewService(repository).(*Service)
	seedVideoFiles(t, repository, []seededVideoFile{{name: "a.mkv", folder: "/data/A"}})
	ownerPlaylistID := seedPlaylists(t, repository, []seededPlaylist{{sourcePath: "/mine", isAuto: false}})["/mine"]

	if err := service.RenumberSeriesPlaylist(ownerPlaylistID); err != ErrPlaylistNotAutomaticSeries {
		t.Fatalf("err = %v, want ErrPlaylistNotAutomaticSeries", err)
	}
}

func TestPlaylistItemsFilterBySeasonAcrossThreePages_Postgres(t *testing.T) {
	fileNames := []string{}
	for season := 1; season <= 3; season++ {
		for episode := 1; episode <= 5; episode++ {
			fileNames = append(fileNames, fmt.Sprintf("Show S%02dE%02d.mkv", season, episode))
		}
	}
	fixture := newSeriesFixture(t, "/data/Show", fileNames)
	playlistID := fixture.seriesPlaylistID(t)
	secondSeason := 2

	var collected []string
	for pageNumber := 1; pageNumber <= 3; pageNumber++ {
		page, err := fixture.service.GetPlaylistItemsPage("tv", playlistID, &secondSeason, pageNumber, 2)
		if err != nil {
			t.Fatalf("page %d: %v", pageNumber, err)
		}
		wantHasNext := pageNumber < 3
		if page.Pagination.HasNext != wantHasNext {
			t.Fatalf("page %d has_next = %v, want %v", pageNumber, page.Pagination.HasNext, wantHasNext)
		}
		for _, item := range page.Items {
			if *item.Season != 2 {
				t.Fatalf("page %d returned season %d", pageNumber, *item.Season)
			}
			collected = append(collected, item.Video.Name)
		}
	}
	want := []string{"Show S02E01.mkv", "Show S02E02.mkv", "Show S02E03.mkv", "Show S02E04.mkv", "Show S02E05.mkv"}
	if !reflect.DeepEqual(collected, want) {
		t.Fatalf("collected = %v, want %v", collected, want)
	}

	emptySeason := 9
	page, err := fixture.service.GetPlaylistItemsPage("tv", playlistID, &emptySeason, 1, 2)
	if err != nil || len(page.Items) != 0 || page.Pagination.HasNext {
		t.Fatalf("unknown season page = %+v err=%v", page, err)
	}
}

func TestPlaylistItemsSeasonFilterUsesEffectiveSeason_Postgres(t *testing.T) {
	fileNames := []string{}
	for episodeNumber := 1; episodeNumber <= 16; episodeNumber++ {
		fileNames = append(fileNames, fmt.Sprintf("Show S01E%02d.mkv", episodeNumber))
	}
	fixture := newSeriesFixture(t, "/data/Show", fileNames)
	playlistID := fixture.seriesPlaylistID(t)
	fixture.overrideEpisode(t, "Show S01E02.mkv", 2, nil)

	secondSeason := 2
	page, err := fixture.service.GetPlaylistItemsPage("tv", playlistID, &secondSeason, 1, 10)
	if err != nil || len(page.Items) != 1 || page.Items[0].Video.Name != "Show S01E02.mkv" {
		t.Fatalf("season 2 page = %+v err=%v", page, err)
	}
}

func assertOrderIndexesContiguous(t *testing.T, repository *Repository, playlistID int, wantCount int) {
	t.Helper()
	if got := len(playlistVideoIDsInOrder(t, repository, playlistID)); got != wantCount {
		t.Fatalf("items = %d, want %d", got, wantCount)
	}
}
