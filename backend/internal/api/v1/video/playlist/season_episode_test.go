package playlist

import (
	"fmt"
	"testing"
)

func TestParseSeasonEpisode(t *testing.T) {
	cases := []struct {
		fileName string
		season   int
		episode  int
		ok       bool
	}{
		{"Serie.S02E03.1080p.x264.mkv", 2, 3, true},
		{"show s2e3.mkv", 2, 3, true},
		{"Show S02 E03.mp4", 2, 3, true},
		{"Show 2x03.mkv", 2, 3, true},
		{"Show Season 3 Episode 7.mkv", 3, 7, true},
		{"Show Ep 3.mkv", 1, 3, true},
		{"Show Episode 3.mkv", 1, 3, true},
		{"Show Episódio 12.mkv", 1, 12, true},
		{"Show Cap 3.mkv", 1, 3, true},
		{"Show.Ep.3.1080p.mkv", 1, 3, true},
		{"Show - 03.mkv", 1, 3, true},
		{"Show - 03 [1080p].mkv", 1, 3, true},
		{"Show - 12 (BD).mkv", 1, 12, true},
		{"Show 1080p.mkv", 0, 0, false},
		{"Show x264.mkv", 0, 0, false},
		{"Show 720p h265.mkv", 0, 0, false},
		{"Show 2019.mkv", 0, 0, false},
		{"Movie - 2019.mkv", 0, 0, false},
		{"Movie (2019) 1920x1080.mkv", 0, 0, false},
		{"Movie 1080p x265.mkv", 0, 0, false},
		{"Show - 1080p.mkv", 0, 0, false},
		{"Holiday.mp4", 0, 0, false},
		{"", 0, 0, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.fileName, func(t *testing.T) {
			season, episode, ok := ParseSeasonEpisode(testCase.fileName)
			if ok != testCase.ok || season != testCase.season || episode != testCase.episode {
				t.Fatalf("ParseSeasonEpisode(%q) = %d, %d, %v; want %d, %d, %v",
					testCase.fileName, season, episode, ok, testCase.season, testCase.episode, testCase.ok)
			}
		})
	}
}

func TestParseSeasonEpisodeReadsEveryEpisodeKeywordNumber(t *testing.T) {
	for episodeNumber := 1; episodeNumber <= 24; episodeNumber++ {
		fileName := fmt.Sprintf("Ep %d.mkv", episodeNumber)
		season, episode, ok := ParseSeasonEpisode(fileName)
		if !ok || season != 1 || episode != episodeNumber {
			t.Fatalf("ParseSeasonEpisode(%q) = %d, %d, %v", fileName, season, episode, ok)
		}
	}
}

func TestAutomaticEpisodeNumberingPrefersCaptureProvenance(t *testing.T) {
	captureSeason := 4
	video := VideoEntry{Name: "Show S01E01.mkv", Series: &SeriesProvenance{Title: "Show", Season: &captureSeason, Episode: 9}}
	numbering := AutomaticEpisodeNumbering(video)
	if *numbering.Season != 4 || *numbering.Episode != 9 {
		t.Fatalf("numbering = %d/%d", *numbering.Season, *numbering.Episode)
	}

	withoutSeason := AutomaticEpisodeNumbering(VideoEntry{Series: &SeriesProvenance{Title: "Show", Episode: 2}})
	if *withoutSeason.Season != 1 || *withoutSeason.Episode != 2 {
		t.Fatalf("numbering = %d/%d", *withoutSeason.Season, *withoutSeason.Episode)
	}

	unknown := AutomaticEpisodeNumbering(VideoEntry{Name: "Holiday.mp4"})
	if unknown.Season != nil || unknown.Episode != nil {
		t.Fatalf("numbering = %+v", unknown)
	}
}

func TestSortEpisodesByNumbering(t *testing.T) {
	number := func(value int) *int { return &value }
	episodes := []NumberedEpisode{
		{VideoID: 1, Name: "z.mkv", Numbering: EpisodeNumbering{}},
		{VideoID: 2, Name: "b.mkv", Numbering: EpisodeNumbering{Season: number(2), Episode: number(1)}},
		{VideoID: 3, Name: "a.mkv", Numbering: EpisodeNumbering{Season: number(1), Episode: number(10)}},
		{VideoID: 4, Name: "c.mkv", Numbering: EpisodeNumbering{Season: number(1), Episode: number(2)}},
		{VideoID: 5, Name: "d.mkv", Numbering: EpisodeNumbering{Season: number(1)}},
	}
	SortEpisodesByNumbering(episodes)
	wantOrder := []int{4, 3, 5, 2, 1}
	for position, wantID := range wantOrder {
		if episodes[position].VideoID != wantID {
			t.Fatalf("position %d = %d, want %d", position, episodes[position].VideoID, wantID)
		}
	}
}

func TestEffectiveNumberingAppliesOverrideFieldByField(t *testing.T) {
	number := func(value int) *int { return &value }
	automatic := EpisodeNumbering{Season: number(1), Episode: number(3)}
	effective := EffectiveNumbering(automatic, EpisodeNumbering{Episode: number(9)})
	if *effective.Season != 1 || *effective.Episode != 9 {
		t.Fatalf("effective = %d/%d", *effective.Season, *effective.Episode)
	}
}

func TestIsSeriesLikePlaylist(t *testing.T) {
	cases := []struct {
		playlistType   string
		classification string
		want           bool
	}{
		{"series", "personal", true},
		{"folder", "series", true},
		{"folder", "anime", true},
		{"folder", "personal", false},
		{"movie", "movie", false},
		{"course", "course", false},
	}
	for _, testCase := range cases {
		if got := IsSeriesLikePlaylist(testCase.playlistType, testCase.classification); got != testCase.want {
			t.Fatalf("IsSeriesLikePlaylist(%q, %q) = %v", testCase.playlistType, testCase.classification, got)
		}
	}
}
