package playlist

import (
	"fmt"
	"testing"
	"time"
)

func smallLibraryWithFolderSeries(episodeCount int) []VideoEntry {
	videos := make([]VideoEntry, 0, episodeCount+4)
	for episode := 1; episode <= episodeCount; episode++ {
		name := fmt.Sprintf("Dark Matter S01E%02d.mkv", episode)
		videos = append(videos, VideoEntry{ID: episode, Name: name, Path: "/tv/Dark Matter/" + name, ParentPath: "/tv/Dark Matter", CreatedAt: time.Now()})
	}
	looseVideos := []VideoEntry{
		{Name: "Inception.mkv", Path: "/movies/Inception.mkv", ParentPath: "/movies", Meta: &VideoMeta{Duration: 8880, Height: 1080}},
		{Name: "The Matrix.mkv", Path: "/movies/The Matrix.mkv", ParentPath: "/movies", Meta: &VideoMeta{Duration: 8160, Height: 1080}},
		{Name: "Family BBQ.mp4", Path: "/personal/Family BBQ.mp4", ParentPath: "/personal", Meta: &VideoMeta{Duration: 600, Height: 1080}},
		{Name: "tutorial.mp4", Path: "/clips/tutorial.mp4", ParentPath: "/clips", Meta: &VideoMeta{Duration: 30}},
	}
	for index, loose := range looseVideos {
		loose.ID = 1000 + index
		loose.CreatedAt = time.Now()
		videos = append(videos, loose)
	}
	return videos
}

func candidateContainingAllEpisodes(candidates []PlaylistCandidate, episodeCount int) *PlaylistCandidate {
	for index, candidate := range candidates {
		if candidate.PlaylistType != "series" {
			continue
		}
		episodesFound := 0
		for _, video := range candidate.Videos {
			if video.Video.Video.ID >= 1 && video.Video.Video.ID <= episodeCount {
				episodesFound++
			}
		}
		if episodesFound == episodeCount {
			return &candidates[index]
		}
	}
	return nil
}

func TestEngineKeepsSmallFolderSeriesAlongsideMixInTinyLibrary(t *testing.T) {
	for _, episodeCount := range []int{2, 3, 6, 10} {
		t.Run(fmt.Sprintf("%d_episodes", episodeCount), func(t *testing.T) {
			result := NewPlaylistEngine().Build(BuildInput{Videos: smallLibraryWithFolderSeries(episodeCount), ClientID: "client"})
			if candidateContainingAllEpisodes(result.Candidates, episodeCount) == nil {
				t.Fatalf("series of %d episodes was discarded; candidates: %v", episodeCount, candidateSourceKeys(result.Candidates))
			}
		})
	}
}

func candidateSourceKeys(candidates []PlaylistCandidate) []string {
	keys := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		keys = append(keys, candidate.SourceKey)
	}
	return keys
}

func seriesCandidate(sourceKey string, score float64, videoIDs ...int) PlaylistCandidate {
	return mixCandidateWith(sourceKey, "series", "sequential_series", score, videoIDs...)
}

func mixCandidateWith(sourceKey, playlistType, strategy string, score float64, videoIDs ...int) PlaylistCandidate {
	scored := make([]ScoredVideo, 0, len(videoIDs))
	for _, videoID := range videoIDs {
		scored = append(scored, ScoredVideo{Video: ClassifiedVideo{Video: VideoEntry{ID: videoID}}, Score: 1})
	}
	return PlaylistCandidate{SourceKey: sourceKey, PlaylistType: playlistType, Strategy: strategy, TotalScore: score, Videos: scored}
}

func TestDeduplicationDropsMixInsteadOfSeriesEvenWhenMixScoresHigher(t *testing.T) {
	engine := NewPlaylistEngine()
	candidates := []PlaylistCandidate{
		mixCandidateWith("random:smart_mix", "mixed", "random_smart", 500, 1, 2, 3, 4, 5, 6, 7),
		seriesCandidate("folder:/tv/show", 100, 1, 2, 3, 4, 5, 6),
	}

	result := engine.applyBusinessRules(candidates)

	if len(result) != 1 || result[0].SourceKey != "folder:/tv/show" {
		t.Fatalf("expected only the series to survive, got %v", candidateSourceKeys(result))
	}
}

func TestDeduplicationAmongSeriesCandidatesKeepsHigherScore(t *testing.T) {
	engine := NewPlaylistEngine()
	candidates := []PlaylistCandidate{
		seriesCandidate("series:a", 200, 1, 2, 3),
		seriesCandidate("folder:/b", 100, 1, 2, 3),
	}

	result := engine.applyBusinessRules(candidates)

	if len(result) != 1 || result[0].SourceKey != "series:a" {
		t.Fatalf("expected higher-scored series to survive, got %v", candidateSourceKeys(result))
	}
}

func TestDeduplicationKeepsMixThatDoesNotOverlapSeries(t *testing.T) {
	engine := NewPlaylistEngine()
	candidates := []PlaylistCandidate{
		mixCandidateWith("random:smart_mix", "mixed", "random_smart", 500, 10, 11, 12, 1),
		seriesCandidate("folder:/tv/show", 100, 1, 2, 3, 4, 5, 6),
	}

	result := engine.applyBusinessRules(candidates)

	if len(result) != 2 {
		t.Fatalf("expected both to survive, got %v", candidateSourceKeys(result))
	}
}
