package playlist

import "testing"

func buildFolderContext(entries ...VideoEntry) *PlaylistContext {
	videos := make([]ClassifiedVideo, 0, len(entries))
	for _, entry := range entries {
		videos = append(videos, ClassifiedVideo{Video: entry, Classification: ClassSeries})
	}
	videosByFolder := map[string][]*ClassifiedVideo{}
	for index := range videos {
		folder := videos[index].Video.ParentPath
		videosByFolder[folder] = append(videosByFolder[folder], &videos[index])
	}
	return &PlaylistContext{Videos: videos, VideosByFolder: videosByFolder}
}

func collectNames(candidate PlaylistCandidate) []string {
	names := make([]string, 0, len(candidate.Videos))
	for _, scoredVideo := range candidate.Videos {
		names = append(names, scoredVideo.Video.Video.Name)
	}
	return names
}

func TestByFolderStrategyOrdersEpisodesNaturally(t *testing.T) {
	context := buildFolderContext(
		VideoEntry{ID: 1, Name: "Ep 10", ParentPath: "/shows/Naruto"},
		VideoEntry{ID: 2, Name: "Ep 2", ParentPath: "/shows/Naruto"},
		VideoEntry{ID: 3, Name: "Ep 1", ParentPath: "/shows/Naruto"},
	)

	candidates := NewByFolderStrategy().Build(context)

	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	assertNameOrder(t, collectNames(candidates[0]), []string{"Ep 1", "Ep 2", "Ep 10"})
}

func TestByFolderStrategyBreaksNameTiesByID(t *testing.T) {
	context := buildFolderContext(
		VideoEntry{ID: 9, Name: "Ep 1", ParentPath: "/shows/Naruto"},
		VideoEntry{ID: 4, Name: "ep 1", ParentPath: "/shows/Naruto"},
	)

	candidates := NewByFolderStrategy().Build(context)

	if candidates[0].Videos[0].Video.Video.ID != 4 {
		t.Fatalf("expected id 4 first, got %d", candidates[0].Videos[0].Video.Video.ID)
	}
}

func TestSequentialSeriesStrategyOrdersSeasonEpisodeCodesNaturally(t *testing.T) {
	context := buildFolderContext(
		VideoEntry{ID: 1, Name: "Show S01E10.mkv", ParentPath: "/a"},
		VideoEntry{ID: 2, Name: "Show S01E02.mkv", ParentPath: "/a"},
	)

	candidates := NewSequentialSeriesStrategy().Build(context)

	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	assertNameOrder(t, collectNames(candidates[0]), []string{"Show S01E02.mkv", "Show S01E10.mkv"})
}

func assertNameOrder(t *testing.T, actual, expected []string) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("got %v, want %v", actual, expected)
	}
	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("got %v, want %v", actual, expected)
		}
	}
}
