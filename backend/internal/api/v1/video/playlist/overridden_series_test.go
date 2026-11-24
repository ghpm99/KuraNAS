package playlist

import (
	"fmt"
	"testing"
	"time"
)

func moviesFolderVideos(folder string, count int) []VideoEntry {
	videos := make([]VideoEntry, 0, count)
	for index := 1; index <= count; index++ {
		name := fmt.Sprintf("Film %c.mkv", 'A'+index-1)
		videos = append(videos, VideoEntry{
			ID: index, Name: name, Path: folder + "/" + name, ParentPath: folder,
			CreatedAt: time.Now(), PersistedClassification: ClassMovie,
		})
	}
	return videos
}

func candidatesContainingVideo(candidates []PlaylistCandidate, videoID int) []PlaylistCandidate {
	found := []PlaylistCandidate{}
	for _, candidate := range candidates {
		for _, scored := range candidate.Videos {
			if scored.Video.Video.ID == videoID {
				found = append(found, candidate)
				break
			}
		}
	}
	return found
}

func TestEngineFormsSeriesFromSingleOverriddenVideoInMoviesFolder(t *testing.T) {
	videos := moviesFolderVideos("/data/Saga", 4)
	videos[0].PersistedClassification = ClassSeries
	videos[0].IsClassificationOverride = true

	result := NewPlaylistEngine().Build(BuildInput{Videos: videos, ClientID: "client"})

	var seriesCandidate *PlaylistCandidate
	for index, candidate := range candidatesContainingVideo(result.Candidates, 1) {
		if candidate.PlaylistType == "series" && candidate.Classification == ClassSeries && len(candidate.Videos) == 1 {
			seriesCandidate = &candidatesContainingVideo(result.Candidates, 1)[index]
		}
		if candidate.GroupMode == "folder" && candidate.Classification == ClassMovie {
			t.Fatalf("movie folder playlist %q still contains the overridden video", candidate.SourceKey)
		}
	}
	if seriesCandidate == nil {
		t.Fatalf("no one-item series candidate for the overridden video; candidates: %v", candidateSourceKeys(result.Candidates))
	}

	folderPlaylistFound := false
	for _, candidate := range result.Candidates {
		if candidate.SourceKey == "folder:/data/Saga" {
			folderPlaylistFound = true
			if candidate.Classification != ClassMovie || len(candidate.Videos) != 3 {
				t.Fatalf("folder playlist = %s with %d videos, want movie with 3", candidate.Classification, len(candidate.Videos))
			}
		}
	}
	if !folderPlaylistFound {
		t.Fatalf("movie folder playlist missing; candidates: %v", candidateSourceKeys(result.Candidates))
	}
}

func TestEngineSplitsOverriddenSeriesByTitlePrefixInGenericFolder(t *testing.T) {
	videos := []VideoEntry{
		{ID: 1, Name: "Alpha Show 01.mkv", ParentPath: "/data/Filmes", PersistedClassification: ClassSeries, IsClassificationOverride: true},
		{ID: 2, Name: "Alpha Show 02.mkv", ParentPath: "/data/Filmes", PersistedClassification: ClassSeries, IsClassificationOverride: true},
		{ID: 3, Name: "Beta Show 01.mkv", ParentPath: "/data/Filmes", PersistedClassification: ClassSeries, IsClassificationOverride: true},
		{ID: 4, Name: "Some Movie.mkv", ParentPath: "/data/Filmes", PersistedClassification: ClassMovie},
		{ID: 5, Name: "Other Movie.mkv", ParentPath: "/data/Filmes", PersistedClassification: ClassMovie},
	}

	classified := NewVideoClassifier().ClassifyAll(videos)
	context := &PlaylistContext{Videos: classified, VideosByFolder: map[string][]*ClassifiedVideo{}}
	for index := range classified {
		folder := classified[index].Video.ParentPath
		context.VideosByFolder[folder] = append(context.VideosByFolder[folder], &classified[index])
	}

	candidates := NewOverriddenSeriesStrategy().Build(context)

	if len(candidates) != 2 {
		t.Fatalf("candidates = %v, want 2 (one per title)", candidateSourceKeys(candidates))
	}
	sizesByName := map[string]int{}
	for _, candidate := range candidates {
		sizesByName[candidate.Name] = len(candidate.Videos)
	}
	if sizesByName["Alpha Show"] != 2 || sizesByName["Beta Show"] != 1 {
		t.Fatalf("sizes by name = %v", sizesByName)
	}
}

func TestEngineExcludesVideoOverriddenToMovieFromSeriesFolder(t *testing.T) {
	videos := make([]VideoEntry, 0, 4)
	for episode := 1; episode <= 4; episode++ {
		name := fmt.Sprintf("Dark Matter S01E%02d.mkv", episode)
		videos = append(videos, VideoEntry{ID: episode, Name: name, ParentPath: "/tv/Dark Matter", PersistedClassification: ClassSeries})
	}
	videos[1].PersistedClassification = ClassMovie
	videos[1].IsClassificationOverride = true

	result := NewPlaylistEngine().Build(BuildInput{Videos: videos, ClientID: "client"})

	for _, candidate := range candidatesContainingVideo(result.Candidates, 2) {
		if candidate.PlaylistType == "series" {
			t.Fatalf("video overridden to movie still in series candidate %q", candidate.SourceKey)
		}
	}
	if len(candidatesContainingVideo(result.Candidates, 1)) == 0 {
		t.Fatalf("remaining episodes lost their series playlist; candidates: %v", candidateSourceKeys(result.Candidates))
	}
}
