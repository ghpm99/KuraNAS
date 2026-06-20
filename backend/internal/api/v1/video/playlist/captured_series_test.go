package playlist

import "testing"

func season(n int) *int { return &n }

func capturedVideo(id int, parent, name, title string, s *int, ep int) ClassifiedVideo {
	return ClassifiedVideo{
		Video: VideoEntry{
			ID:         id,
			Name:       name,
			ParentPath: parent,
			Series:     &SeriesProvenance{Title: title, Season: s, Episode: ep},
		},
		Classification: ClassSeries,
	}
}

func contextOf(videos []ClassifiedVideo) *PlaylistContext {
	return &PlaylistContext{
		Videos:         videos,
		VideoByID:      indexByID(videos),
		VideosByFolder: indexByFolder(videos),
	}
}

func TestCapturedSeriesStrategy_SplitsDistinctSeries(t *testing.T) {
	// Duas series diferentes capturadas pelo plugin, cada uma em sua pasta.
	videos := []ClassifiedVideo{
		capturedVideo(1, "/Filmes/Frieren/Temporada 1", "E1.mp4", "Frieren e a Jornada para o Alem", season(1), 1),
		capturedVideo(2, "/Filmes/Frieren/Temporada 1", "E2.mp4", "Frieren e a Jornada para o Alem", season(1), 2),
		capturedVideo(3, "/Filmes/Dandadan/Temporada 1", "E1.mp4", "Dandadan", season(1), 1),
	}

	candidates := NewCapturedSeriesStrategy().Build(contextOf(videos))

	if len(candidates) != 2 {
		t.Fatalf("expected 2 distinct series playlists, got %d", len(candidates))
	}

	byKey := map[string]PlaylistCandidate{}
	for _, c := range candidates {
		byKey[c.SourceKey] = c
		if c.PlaylistType != "series" || c.GroupMode != "captured_series" {
			t.Errorf("unexpected type/mode for %s: %s/%s", c.SourceKey, c.PlaylistType, c.GroupMode)
		}
	}

	frieren, ok := byKey["series:capture:frieren e a jornada para o alem"]
	if !ok {
		t.Fatalf("missing Frieren playlist, got keys %v", keysOf(byKey))
	}
	if frieren.Name != "Frieren e a Jornada para o Alem" {
		t.Errorf("expected title-cased name, got %q", frieren.Name)
	}
	if len(frieren.Videos) != 2 {
		t.Errorf("expected 2 episodes in Frieren, got %d", len(frieren.Videos))
	}

	if dandadan, ok := byKey["series:capture:dandadan"]; !ok {
		t.Error("missing Dandadan playlist")
	} else if len(dandadan.Videos) != 1 {
		t.Errorf("expected 1 episode in Dandadan, got %d", len(dandadan.Videos))
	}
}

func TestCapturedSeriesStrategy_OrdersBySeasonEpisode(t *testing.T) {
	// Entrada fora de ordem (e cruzando temporadas) deve sair por (season, episode).
	videos := []ClassifiedVideo{
		capturedVideo(10, "/x", "b.mp4", "Show", season(2), 1),
		capturedVideo(11, "/x", "c.mp4", "Show", season(1), 3),
		capturedVideo(12, "/x", "a.mp4", "Show", season(1), 1),
	}

	candidates := NewCapturedSeriesStrategy().Build(contextOf(videos))
	if len(candidates) != 1 {
		t.Fatalf("expected 1 playlist, got %d", len(candidates))
	}

	gotOrder := make([]int, 0, 3)
	for _, sv := range candidates[0].Videos {
		gotOrder = append(gotOrder, sv.Video.Video.ID)
	}
	want := []int{12, 11, 10} // S1E1, S1E3, S2E1
	for i := range want {
		if gotOrder[i] != want[i] {
			t.Fatalf("episode order = %v, want %v", gotOrder, want)
		}
	}
}

func TestRelatedContentStrategy_ExcludesCapturedSeries(t *testing.T) {
	// Uma serie capturada + dois videos series avulsos (sem proveniencia).
	// O balde generico "Series" deve conter apenas os avulsos.
	videos := []ClassifiedVideo{
		capturedVideo(1, "/Filmes/Frieren/Temporada 1", "E1.mp4", "Frieren", season(1), 1),
		capturedVideo(2, "/Filmes/Frieren/Temporada 1", "E2.mp4", "Frieren", season(1), 2),
		{Video: VideoEntry{ID: 3, Name: "avulso a.mkv", ParentPath: "/tv"}, Classification: ClassSeries},
		{Video: VideoEntry{ID: 4, Name: "avulso b.mkv", ParentPath: "/tv"}, Classification: ClassSeries},
	}

	candidates := NewRelatedContentStrategy().Build(contextOf(videos))

	for _, c := range candidates {
		if c.SourceKey != "related:series" {
			continue
		}
		if len(c.Videos) != 2 {
			t.Fatalf("expected only the 2 non-captured videos in related:series, got %d", len(c.Videos))
		}
		for _, sv := range c.Videos {
			if sv.Video.Video.Series != nil {
				t.Errorf("captured video %d leaked into the generic bucket", sv.Video.Video.ID)
			}
		}
	}
}

func TestFullEngine_CapturedSeriesPreservesOrder(t *testing.T) {
	engine := NewPlaylistEngine()

	input := BuildInput{
		Videos: []VideoEntry{
			{ID: 1, Name: "E2.mp4", ParentPath: "/Filmes/Frieren/Temporada 1", Series: &SeriesProvenance{Title: "Frieren", Season: season(1), Episode: 2}},
			{ID: 2, Name: "E1.mp4", ParentPath: "/Filmes/Frieren/Temporada 1", Series: &SeriesProvenance{Title: "Frieren", Season: season(1), Episode: 1}},
			{ID: 3, Name: "E3.mp4", ParentPath: "/Filmes/Frieren/Temporada 1", Series: &SeriesProvenance{Title: "Frieren", Season: season(1), Episode: 3}},
		},
	}

	groups := engine.Build(input).ToSmartGroups()

	var frieren *SmartGroup
	for i := range groups {
		if groups[i].SourceKey == "series:capture:frieren" {
			frieren = &groups[i]
			break
		}
	}
	if frieren == nil {
		t.Fatalf("expected captured series group, got %d groups", len(groups))
	}

	// ToSmartGroups deve preservar a ordem de episodio (E1, E2, E3 => IDs 2,1,3),
	// nao reordenar por ID de arquivo.
	want := []int{2, 1, 3}
	if len(frieren.VideoIDs) != len(want) {
		t.Fatalf("VideoIDs = %v, want %v", frieren.VideoIDs, want)
	}
	for i := range want {
		if frieren.VideoIDs[i] != want[i] {
			t.Fatalf("persisted order = %v, want %v (episode order, not file-id order)", frieren.VideoIDs, want)
		}
	}
}

func keysOf(m map[string]PlaylistCandidate) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
