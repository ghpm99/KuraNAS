package playlist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func loadTranslationCatalog(t *testing.T, locale string) map[string]string {
	t.Helper()
	catalogPath := filepath.Join("..", "..", "..", "..", "..", "translations", locale+".json")
	content, err := os.ReadFile(catalogPath)
	if err != nil {
		t.Fatalf("read %s: %v", catalogPath, err)
	}
	catalog := map[string]string{}
	if err := json.Unmarshal(content, &catalog); err != nil {
		t.Fatalf("decode %s: %v", catalogPath, err)
	}
	return catalog
}

func allPlaylistNameKeys() []string {
	nameKeys := []string{NameKeyContinueWatching, NameKeyRecommended, NameKeyMixed, NameKeyUnknownSeries, NameKeyVideos}
	for _, nameKey := range classificationNameKeys {
		nameKeys = append(nameKeys, nameKey)
	}
	return nameKeys
}

func TestEveryPlaylistNameKeyIsTranslatedInBothLocales(t *testing.T) {
	for _, locale := range []string{"pt-BR", "en-US"} {
		catalog := loadTranslationCatalog(t, locale)
		for _, nameKey := range allPlaylistNameKeys() {
			if catalog[nameKey] == "" {
				t.Errorf("%s is missing key %s", locale, nameKey)
			}
		}
	}
}

func TestClassificationNameKeyCoversEveryClassificationAndFallsBack(t *testing.T) {
	classifications := []VideoClassification{ClassSeries, ClassMovie, ClassAnime, ClassCourse, ClassPersonal, ClassClip, ClassMusic, ClassProgram}
	seenKeys := map[string]bool{}
	for _, classification := range classifications {
		nameKey := classificationNameKey(classification)
		if nameKey == NameKeyVideos || seenKeys[nameKey] {
			t.Fatalf("classification %s mapped to unexpected key %s", classification, nameKey)
		}
		seenKeys[nameKey] = true
	}
	if classificationNameKey(VideoClassification("unknown")) != NameKeyVideos {
		t.Fatal("unknown classification must fall back to the generic videos key")
	}
}

func TestFolderAndSeriesCandidatesKeepDataNamesWithoutKey(t *testing.T) {
	folderCandidates := NewByFolderStrategy().Build(buildFolderContext(
		VideoEntry{ID: 1, Name: "Ep 1", ParentPath: "/shows/Naruto"},
		VideoEntry{ID: 2, Name: "Ep 2", ParentPath: "/shows/Naruto"},
	))
	if len(folderCandidates) != 1 || folderCandidates[0].Name != "Naruto" || folderCandidates[0].NameKey != "" {
		t.Fatalf("folder candidate must keep the folder name as data, got %+v", folderCandidates)
	}

	seriesCandidates := NewSequentialSeriesStrategy().Build(buildFolderContext(
		VideoEntry{ID: 1, Name: "Show S01E01.mkv", ParentPath: "/a"},
		VideoEntry{ID: 2, Name: "Show S01E02.mkv", ParentPath: "/a"},
	))
	if len(seriesCandidates) != 1 || seriesCandidates[0].Name != "Show" || seriesCandidates[0].NameKey != "" {
		t.Fatalf("series candidate must keep the detected prefix as data, got %+v", seriesCandidates)
	}
}

func TestCapturedSeriesCandidateKeepsCapturedTitleWithoutKey(t *testing.T) {
	season := 1
	entries := []VideoEntry{
		{ID: 1, Name: "a.mp4", ParentPath: "/c", Series: &SeriesProvenance{Title: "Frieren", Season: &season, Episode: 1}},
		{ID: 2, Name: "b.mp4", ParentPath: "/c", Series: &SeriesProvenance{Title: "Frieren", Season: &season, Episode: 2}},
	}
	candidates := NewCapturedSeriesStrategy().Build(buildFolderContext(entries...))
	if len(candidates) != 1 || candidates[0].Name != "Frieren" || candidates[0].NameKey != "" {
		t.Fatalf("captured series must keep its title as data, got %+v", candidates)
	}
}

func TestSeriesWithoutPrefixUsesUnknownSeriesKey(t *testing.T) {
	if seriesNameKey("") != NameKeyUnknownSeries {
		t.Fatal("empty prefix must use the unknown series key")
	}
	if seriesNameKey("show") != "" {
		t.Fatal("detected prefix is data and must not carry a key")
	}
}

func TestContinueWatchingCandidateUsesKeyInsteadOfLiteralName(t *testing.T) {
	videos := []ClassifiedVideo{
		{Video: VideoEntry{ID: 1, Name: "EP01.mkv", ParentPath: "/series"}, Classification: ClassSeries},
		{Video: VideoEntry{ID: 2, Name: "EP02.mkv", ParentPath: "/series"}, Classification: ClassSeries},
	}
	candidates := NewContinueWatchingStrategy().Build(&PlaylistContext{
		Videos:         videos,
		VideoByID:      indexByID(videos),
		VideosByFolder: indexByFolder(videos),
		ClientID:       "client1",
		PlaybackState:  &PlaybackSnapshot{VideoID: 1, CurrentTime: 120, Duration: 3600},
	})
	if len(candidates) != 1 || candidates[0].NameKey != NameKeyContinueWatching || candidates[0].Name != "" {
		t.Fatalf("continue candidate must carry only the key, got %+v", candidates)
	}
}

func TestRelatedContentCandidateUsesClassificationKey(t *testing.T) {
	videos := []ClassifiedVideo{
		{Video: VideoEntry{ID: 1, Name: "a.mkv", ParentPath: "/m"}, Classification: ClassMovie},
		{Video: VideoEntry{ID: 2, Name: "b.mkv", ParentPath: "/m"}, Classification: ClassMovie},
	}
	candidates := NewRelatedContentStrategy().Build(&PlaylistContext{Videos: videos})
	if len(candidates) == 0 {
		t.Fatal("expected related content candidate")
	}
	for _, candidate := range candidates {
		if candidate.NameKey != classificationNameKey(ClassMovie) || candidate.Name != "" {
			t.Fatalf("related candidate must carry the classification key, got %+v", candidate)
		}
	}
}

func TestToSmartGroupsCarriesNameKey(t *testing.T) {
	result := BuildResult{Candidates: []PlaylistCandidate{
		{SourceKey: "random:smart_mix", NameKey: NameKeyMixed},
		{SourceKey: "folder:/x", Name: "x"},
	}}
	groups := result.ToSmartGroups()
	if groups[0].NameKey != NameKeyMixed || groups[1].NameKey != "" || groups[1].Name != "x" {
		t.Fatalf("unexpected groups %+v", groups)
	}
}
