package playlist

import (
	"testing"
	"time"
)

func TestClassifier_SeriesDetection(t *testing.T) {
	c := NewVideoClassifier()

	tests := []struct {
		name     string
		video    VideoEntry
		wantClass VideoClassification
	}{
		{
			name:      "episode pattern S01E02",
			video:     VideoEntry{Name: "Breaking Bad S01E02.mkv", Path: "/tv/Breaking Bad S01E02.mkv", ParentPath: "/tv"},
			wantClass: ClassSeries,
		},
		{
			name:      "episode pattern EP05 in anime folder",
			video:     VideoEntry{Name: "Naruto EP05.mp4", Path: "/anime/Naruto EP05.mp4", ParentPath: "/anime"},
			wantClass: ClassSeries, // episode_pattern (priority 1) beats anime_path (priority 2)
		},
		{
			name:      "movie by path",
			video:     VideoEntry{Name: "Inception.mkv", Path: "/movies/Inception.mkv", ParentPath: "/movies"},
			wantClass: ClassMovie,
		},
		{
			name:      "course by keyword",
			video:     VideoEntry{Name: "Aula 01 - Introducao.mp4", Path: "/cursos/go/Aula 01 - Introducao.mp4", ParentPath: "/cursos/go"},
			wantClass: ClassCourse,
		},
		{
			name:  "clip by short duration",
			video: VideoEntry{Name: "random.mp4", Path: "/personal/random.mp4", ParentPath: "/personal", Meta: &VideoMeta{Duration: 30}},
			wantClass: ClassClip,
		},
		{
			name:  "movie by long duration and HD",
			video: VideoEntry{Name: "something.mkv", Path: "/stuff/something.mkv", ParentPath: "/stuff", Meta: &VideoMeta{Duration: 7200, Height: 1080}},
			wantClass: ClassMovie,
		},
		{
			name:      "personal fallback",
			video:     VideoEntry{Name: "family.mp4", Path: "/personal/family.mp4", ParentPath: "/personal"},
			wantClass: ClassPersonal,
		},
		{
			name:      "program by steam keyword",
			video:     VideoEntry{Name: "gameplay.mp4", Path: "/steam/recordings/gameplay.mp4", ParentPath: "/steam/recordings"},
			wantClass: ClassProgram,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Classify(tt.video)
			if result.Classification != tt.wantClass {
				t.Errorf("got %s, want %s (rules: %v, confidence: %.2f)",
					result.Classification, tt.wantClass, result.MatchedRules, result.Confidence)
			}
		})
	}
}

func TestInferTitlePrefix(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Breaking Bad S01E02.mkv", "breaking bad"},
		{"Naruto EP05.mp4", "naruto"},
		{"My.Show.S02E10.720p.mkv", "my show 720p"},
		{"[SubGroup] Anime - 03.mkv", "anime"},
		{"Movie.mkv", "movie"},
		{"Aula 01.mp4", "aula"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferTitlePrefix(tt.name)
			if got != tt.want {
				t.Errorf("InferTitlePrefix(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsGenericFolderName(t *testing.T) {
	generic := []string{"videos", "Downloads", "MEDIA", "desktop", "", "misc"}
	for _, name := range generic {
		if !IsGenericFolderName(name) {
			t.Errorf("expected %q to be generic", name)
		}
	}

	specific := []string{"Breaking Bad", "Season 1", "My Course", "Vacation 2024"}
	for _, name := range specific {
		if IsGenericFolderName(name) {
			t.Errorf("expected %q to NOT be generic", name)
		}
	}
}

func TestByFolderStrategy(t *testing.T) {
	strategy := NewByFolderStrategy()
	ctx := buildTestContext(t)

	candidates := strategy.Build(ctx)

	// Deve agrupar /series/breaking_bad (2 videos) mas nao /downloads (generico)
	found := false
	for _, c := range candidates {
		if c.SourceKey == "folder:/series/breaking_bad" {
			found = true
			if len(c.Videos) != 2 {
				t.Errorf("expected 2 videos in folder group, got %d", len(c.Videos))
			}
			if c.Classification != ClassSeries {
				t.Errorf("expected series classification, got %s", c.Classification)
			}
		}
		if c.SourceKey == "folder:/downloads" {
			t.Error("generic folder /downloads should not create a group")
		}
	}
	if !found {
		t.Error("expected folder group for /series/breaking_bad")
	}
}

func TestSequentialSeriesStrategy_CrossFolder(t *testing.T) {
	strategy := NewSequentialSeriesStrategy()

	// Videos da mesma serie em pastas DIFERENTES
	videos := []ClassifiedVideo{
		{Video: VideoEntry{ID: 1, Name: "Show S01E01.mkv", ParentPath: "/folder_a"}, Classification: ClassSeries},
		{Video: VideoEntry{ID: 2, Name: "Show S01E02.mkv", ParentPath: "/folder_b"}, Classification: ClassSeries},
		{Video: VideoEntry{ID: 3, Name: "Other Movie.mkv", ParentPath: "/folder_c"}, Classification: ClassMovie},
	}

	ctx := &PlaylistContext{
		Videos:         videos,
		VideoByID:      indexByID(videos),
		VideosByFolder: indexByFolder(videos),
	}

	candidates := strategy.Build(ctx)

	// Deve agrupar "show" cross-folder
	found := false
	for _, c := range candidates {
		if c.Name == "Show" {
			found = true
			if len(c.Videos) != 2 {
				t.Errorf("expected 2 videos in series group, got %d", len(c.Videos))
			}
		}
	}
	if !found {
		t.Error("expected cross-folder series grouping for 'show'")
	}
}

func TestContinueWatchingStrategy(t *testing.T) {
	strategy := NewContinueWatchingStrategy()

	videos := []ClassifiedVideo{
		{Video: VideoEntry{ID: 1, Name: "EP01.mkv", ParentPath: "/series"}, Classification: ClassSeries},
		{Video: VideoEntry{ID: 2, Name: "EP02.mkv", ParentPath: "/series"}, Classification: ClassSeries},
	}

	ctx := &PlaylistContext{
		Videos:         videos,
		VideoByID:      indexByID(videos),
		VideosByFolder: indexByFolder(videos),
		ClientID:       "client1",
		PlaybackState:  &PlaybackSnapshot{VideoID: 1, CurrentTime: 120, Duration: 3600},
	}

	candidates := strategy.Build(ctx)
	if len(candidates) == 0 {
		t.Fatal("expected continue watching candidate")
	}
	if candidates[0].Videos[0].Video.Video.ID != 1 {
		t.Error("first video should be the in-progress one")
	}
	if candidates[0].Videos[0].Score != 100.0 {
		t.Errorf("in-progress video should have max score, got %.1f", candidates[0].Videos[0].Score)
	}
}

func TestContinueWatchingStrategy_NoState(t *testing.T) {
	strategy := NewContinueWatchingStrategy()
	ctx := &PlaylistContext{}
	candidates := strategy.Build(ctx)
	if len(candidates) != 0 {
		t.Error("expected no candidates without playback state")
	}
}

func TestScoringEngine(t *testing.T) {
	engine := NewScoringEngine()

	candidate := PlaylistCandidate{
		Videos: []ScoredVideo{
			{
				Video: ClassifiedVideo{
					Video:          VideoEntry{ID: 1, Name: "Show S01E01.mkv", Meta: &VideoMeta{Duration: 2400, Height: 1080}},
					Classification: ClassSeries,
					Confidence:     0.9,
				},
				Score: 1.0,
			},
			{
				Video: ClassifiedVideo{
					Video:          VideoEntry{ID: 2, Name: "random_clip.mp4", Meta: &VideoMeta{Duration: 15, Height: 360}},
					Classification: ClassClip,
					Confidence:     0.5,
				},
				Score: 1.0,
			},
		},
	}

	ctx := &ScoringContext{Candidate: &candidate}
	engine.ScoreCandidate(&candidate, ctx)

	// O episodio da serie deve ter score maior que o clip
	if candidate.Videos[0].Score <= candidate.Videos[1].Score {
		t.Errorf("series episode should score higher than clip: %.1f vs %.1f",
			candidate.Videos[0].Score, candidate.Videos[1].Score)
	}
}

func TestStateMachine_ValidTransitions(t *testing.T) {
	sm := NewPlaybackStateMachine("client1")

	steps := []struct {
		event    PlaybackEvent
		expected PlaybackState
	}{
		{EventStart, StateLoading},
		{EventLoaded, StatePlaying},
		{EventPause, StatePaused},
		{EventResume, StatePlaying},
		{EventComplete, StateCompleted},
		{EventStart, StateLoading}, // reset apos terminal
	}

	for _, step := range steps {
		if err := sm.HandleEvent(step.event, 1, 1, 0, 3600); err != nil {
			t.Fatalf("unexpected error for event %s: %v", step.event, err)
		}
		if sm.State != step.expected {
			t.Fatalf("after event %s: expected state %s, got %s", step.event, step.expected, sm.State)
		}
	}
}

func TestStateMachine_InvalidTransition(t *testing.T) {
	sm := NewPlaybackStateMachine("client1")

	// Tentar pausar sem estar playing
	err := sm.HandleEvent(EventPause, 1, 1, 0, 0)
	if err == nil {
		t.Fatal("expected error for invalid transition idle->pause")
	}
}

func TestStateMachine_Skip(t *testing.T) {
	sm := NewPlaybackStateMachine("client1")
	sm.HandleEvent(EventStart, 1, 1, 0, 3600)
	sm.HandleEvent(EventLoaded, 1, 1, 0, 3600)
	sm.HandleEvent(EventSkip, 1, 1, 30, 3600)

	if sm.State != StateSkipped {
		t.Fatalf("expected skipped state, got %s", sm.State)
	}
	if !sm.ShouldAutoAdvance() {
		t.Error("skipped should trigger auto advance")
	}
}

func TestStateMachine_WatchedPercent(t *testing.T) {
	sm := NewPlaybackStateMachine("client1")
	sm.Position = 300
	sm.Duration = 600

	if pct := sm.WatchedPercent(); pct != 50.0 {
		t.Errorf("expected 50%%, got %.1f%%", pct)
	}
}

func TestBehaviorAnalyzer(t *testing.T) {
	analyzer := NewBehaviorAnalyzer()

	videos := map[int]*ClassifiedVideo{
		1: {Video: VideoEntry{ID: 1}, Classification: ClassSeries},
		2: {Video: VideoEntry{ID: 2}, Classification: ClassSeries},
		3: {Video: VideoEntry{ID: 3}, Classification: ClassMovie},
	}

	events := []BehaviorEvent{
		{ClientID: "c1", VideoID: 1, EventType: EventStarted, Duration: 2400},
		{ClientID: "c1", VideoID: 1, EventType: EventCompleted, Duration: 2400, Position: 2400},
		{ClientID: "c1", VideoID: 2, EventType: EventStarted, Duration: 2400},
		{ClientID: "c1", VideoID: 2, EventType: EventCompleted, Duration: 2400, Position: 2400},
		{ClientID: "c1", VideoID: 3, EventType: EventStarted, Duration: 7200},
		{ClientID: "c1", VideoID: 3, EventType: EventSkipped, Duration: 7200, Position: 300},
	}

	profile := analyzer.BuildProfile(events, videos)
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}

	// 2 de 3 completados
	if profile.CompletionRate < 0.6 || profile.CompletionRate > 0.7 {
		t.Errorf("expected ~66%% completion rate, got %.2f", profile.CompletionRate)
	}

	// Series deve ter afinidade alta
	if profile.PreferredTypes[ClassSeries] <= 0 {
		t.Error("expected positive affinity for series")
	}

	// Video 3 deve estar nos skipped
	foundSkipped := false
	for _, id := range profile.RecentlySkipped {
		if id == 3 {
			foundSkipped = true
		}
	}
	if !foundSkipped {
		t.Error("expected video 3 in recently skipped")
	}
}

func TestFullEngine_Integration(t *testing.T) {
	engine := NewPlaylistEngine()

	input := BuildInput{
		Videos: []VideoEntry{
			{ID: 1, Name: "Breaking Bad S01E01.mkv", Path: "/series/bb/Breaking Bad S01E01.mkv", ParentPath: "/series/bb", CreatedAt: time.Now()},
			{ID: 2, Name: "Breaking Bad S01E02.mkv", Path: "/series/bb/Breaking Bad S01E02.mkv", ParentPath: "/series/bb", CreatedAt: time.Now()},
			{ID: 3, Name: "Breaking Bad S01E03.mkv", Path: "/series/bb/Breaking Bad S01E03.mkv", ParentPath: "/series/bb", CreatedAt: time.Now()},
			{ID: 4, Name: "Inception.mkv", Path: "/movies/Inception.mkv", ParentPath: "/movies", Meta: &VideoMeta{Duration: 8880, Height: 1080}, CreatedAt: time.Now()},
			{ID: 5, Name: "The Matrix.mkv", Path: "/movies/The Matrix.mkv", ParentPath: "/movies", Meta: &VideoMeta{Duration: 8160, Height: 1080}, CreatedAt: time.Now()},
			{ID: 6, Name: "tutorial.mp4", Path: "/steam/tutorial.mp4", ParentPath: "/steam", Meta: &VideoMeta{Duration: 30}, CreatedAt: time.Now()},
			{ID: 7, Name: "Family BBQ 2024.mp4", Path: "/personal/Family BBQ 2024.mp4", ParentPath: "/personal", Meta: &VideoMeta{Duration: 600, Height: 1080}, CreatedAt: time.Now()},
		},
		ClientID: "test_client",
	}

	result := engine.Build(input)

	if len(result.Candidates) == 0 {
		t.Fatal("expected at least one playlist candidate")
	}
	if len(result.ClassifiedVideos) != 7 {
		t.Errorf("expected 7 classified videos, got %d", len(result.ClassifiedVideos))
	}
	if len(result.StrategyReasons) == 0 {
		t.Error("expected strategy reasons")
	}

	// Verificar que Breaking Bad foi agrupado (por folder ou por prefix)
	foundBB := false
	for _, c := range result.Candidates {
		for _, sv := range c.Videos {
			if sv.Video.Video.ID == 1 || sv.Video.Video.ID == 2 || sv.Video.Video.ID == 3 {
				// Verificar que os 3 episodios estao juntos em algum candidate
				bbCount := 0
				for _, v := range c.Videos {
					if v.Video.Video.ID >= 1 && v.Video.Video.ID <= 3 {
						bbCount++
					}
				}
				if bbCount == 3 {
					foundBB = true
				}
				break
			}
		}
		if foundBB {
			break
		}
	}
	if !foundBB {
		t.Error("expected Breaking Bad episodes grouped together in some candidate")
		for _, c := range result.Candidates {
			t.Logf("  candidate: %s (%s) with %d videos", c.Name, c.SourceKey, len(c.Videos))
		}
	}

	// Verificar que tutorial/steam nao cria playlist de programa
	for _, c := range result.Candidates {
		for _, sv := range c.Videos {
			if sv.Video.Video.ID == 6 && c.PlaylistType == "series" {
				t.Error("tutorial should not be in a series playlist")
			}
		}
	}

	// Testar compatibilidade com formato antigo
	groups := result.ToSmartGroups()
	if len(groups) == 0 {
		t.Error("expected smart groups from conversion")
	}
	for _, g := range groups {
		if g.SourceKey == "" || g.Name == "" {
			t.Errorf("smart group missing key/name: %+v", g)
		}
	}
}

func TestEngine_WithBehavior(t *testing.T) {
	engine := NewPlaylistEngine()

	now := time.Now()
	input := BuildInput{
		Videos: []VideoEntry{
			{ID: 1, Name: "Show EP01.mkv", ParentPath: "/shows", Path: "/shows/Show EP01.mkv", CreatedAt: now},
			{ID: 2, Name: "Show EP02.mkv", ParentPath: "/shows", Path: "/shows/Show EP02.mkv", CreatedAt: now},
			{ID: 3, Name: "Movie.mkv", ParentPath: "/movies", Path: "/movies/Movie.mkv", CreatedAt: now, Meta: &VideoMeta{Duration: 7200, Height: 1080}},
			{ID: 4, Name: "Clip.mp4", ParentPath: "/clips", Path: "/clips/Clip.mp4", CreatedAt: now, Meta: &VideoMeta{Duration: 20}},
			{ID: 5, Name: "Extra.mkv", ParentPath: "/shows", Path: "/shows/Extra.mkv", CreatedAt: now},
		},
		ClientID:      "c1",
		PlaybackState: &PlaybackSnapshot{VideoID: 1, PlaylistID: 1, CurrentTime: 300, Duration: 2400},
		BehaviorEvents: []BehaviorEvent{
			{ClientID: "c1", VideoID: 1, EventType: EventStarted, Duration: 2400},
			{ClientID: "c1", VideoID: 2, EventType: EventStarted, Duration: 2400},
			{ClientID: "c1", VideoID: 2, EventType: EventCompleted, Duration: 2400, Position: 2400},
			{ClientID: "c1", VideoID: 4, EventType: EventStarted, Duration: 20},
			{ClientID: "c1", VideoID: 4, EventType: EventSkipped, Duration: 20, Position: 5},
		},
	}

	result := engine.Build(input)

	// Deve ter "continue watching" como um dos candidates
	hasContinue := false
	for _, c := range result.Candidates {
		if c.PlaylistType == "continue" {
			hasContinue = true
			break
		}
	}
	if !hasContinue {
		t.Error("expected continue_watching candidate when playback in progress")
	}
}

func TestEngine_EmptyInput(t *testing.T) {
	engine := NewPlaylistEngine()
	result := engine.Build(BuildInput{})

	if len(result.Candidates) != 0 {
		t.Errorf("expected 0 candidates for empty input, got %d", len(result.Candidates))
	}
}

func TestDeduplication(t *testing.T) {
	engine := NewPlaylistEngine()
	engine.MinPlaylistSize = 1

	// Dois candidates com mesmos videos (>80% overlap)
	candidates := []PlaylistCandidate{
		{
			SourceKey:  "a",
			TotalScore: 100,
			Videos: []ScoredVideo{
				{Video: ClassifiedVideo{Video: VideoEntry{ID: 1}}, Score: 50},
				{Video: ClassifiedVideo{Video: VideoEntry{ID: 2}}, Score: 50},
			},
		},
		{
			SourceKey:  "b",
			TotalScore: 50,
			Videos: []ScoredVideo{
				{Video: ClassifiedVideo{Video: VideoEntry{ID: 1}}, Score: 25},
				{Video: ClassifiedVideo{Video: VideoEntry{ID: 2}}, Score: 25},
			},
		},
	}

	result := engine.applyBusinessRules(candidates)
	if len(result) != 1 {
		t.Errorf("expected 1 candidate after dedup, got %d", len(result))
	}
	if result[0].SourceKey != "a" {
		t.Error("expected higher-scored candidate to survive")
	}
}

func TestChain_FallbackToFolder(t *testing.T) {
	chain := NewStrategyChain() // vazio, sem handlers
	ctx := &PlaylistContext{}
	decision := chain.Resolve(ctx)

	if len(decision.Strategies) == 0 {
		t.Fatal("expected fallback strategy")
	}
	if decision.Strategies[0].Name() != "by_folder" {
		t.Errorf("expected by_folder fallback, got %s", decision.Strategies[0].Name())
	}
}

func TestChain_DefaultFullPipeline(t *testing.T) {
	chain := DefaultStrategyChain()

	videos := []ClassifiedVideo{
		{Video: VideoEntry{ID: 1, ParentPath: "/series"}, Classification: ClassSeries},
		{Video: VideoEntry{ID: 2, ParentPath: "/series"}, Classification: ClassSeries},
		{Video: VideoEntry{ID: 3, ParentPath: "/movies"}, Classification: ClassMovie},
		{Video: VideoEntry{ID: 4, ParentPath: "/movies"}, Classification: ClassMovie},
		{Video: VideoEntry{ID: 5, ParentPath: "/personal"}, Classification: ClassPersonal},
	}

	ctx := &PlaylistContext{
		Videos:         videos,
		VideoByID:      indexByID(videos),
		VideosByFolder: indexByFolder(videos),
	}

	decision := chain.Resolve(ctx)
	if len(decision.Strategies) < 2 {
		t.Errorf("expected multiple strategies for diverse content, got %d", len(decision.Strategies))
	}

	// Deve incluir series detection e folder grouping
	strategyNames := map[string]bool{}
	for _, s := range decision.Strategies {
		strategyNames[s.Name()] = true
	}
	if !strategyNames["sequential_series"] {
		t.Error("expected sequential_series strategy")
	}
	if !strategyNames["by_folder"] {
		t.Error("expected by_folder strategy")
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func buildTestContext(t *testing.T) *PlaylistContext {
	t.Helper()

	classifier := NewVideoClassifier()
	videos := []VideoEntry{
		{ID: 1, Name: "Breaking Bad S01E01.mkv", Path: "/series/breaking_bad/Breaking Bad S01E01.mkv", ParentPath: "/series/breaking_bad", CreatedAt: time.Now()},
		{ID: 2, Name: "Breaking Bad S01E02.mkv", Path: "/series/breaking_bad/Breaking Bad S01E02.mkv", ParentPath: "/series/breaking_bad", CreatedAt: time.Now()},
		{ID: 3, Name: "Random.mp4", Path: "/downloads/Random.mp4", ParentPath: "/downloads", CreatedAt: time.Now()},
		{ID: 4, Name: "Movie.mkv", Path: "/movies/Movie.mkv", ParentPath: "/movies", CreatedAt: time.Now()},
	}

	classified := classifier.ClassifyAll(videos)

	return &PlaylistContext{
		Videos:         classified,
		VideoByID:      indexByID(classified),
		VideosByFolder: indexByFolder(classified),
	}
}

func indexByID(videos []ClassifiedVideo) map[int]*ClassifiedVideo {
	m := make(map[int]*ClassifiedVideo, len(videos))
	for i := range videos {
		m[videos[i].Video.ID] = &videos[i]
	}
	return m
}

func indexByFolder(videos []ClassifiedVideo) map[string][]*ClassifiedVideo {
	m := make(map[string][]*ClassifiedVideo)
	for i := range videos {
		folder := videos[i].Video.ParentPath
		m[folder] = append(m[folder], &videos[i])
	}
	return m
}
