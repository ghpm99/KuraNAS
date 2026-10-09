package search

import (
	"context"
	"errors"
	"nas-go/api/pkg/ai"
	"testing"
	"time"
)

type searchRepositoryMock struct {
	searchFilesFn          func(query string, limit int) ([]FileResultModel, error)
	searchFoldersFn        func(query string, limit int) ([]FolderResultModel, error)
	searchArtistsFn        func(query string, limit int) ([]ArtistResultModel, error)
	searchAlbumsFn         func(query string, limit int) ([]AlbumResultModel, error)
	searchMusicPlaylistsFn func(query string, limit int) ([]MusicPlaylistResultModel, error)
	searchVideoPlaylistsFn func(query string, limit int) ([]VideoPlaylistResultModel, error)
	searchVideosFn         func(query string, limit int) ([]VideoResultModel, error)
	searchImagesFn         func(query string, limit int) ([]ImageResultModel, error)
}

func (m *searchRepositoryMock) SearchFiles(query string, limit int) ([]FileResultModel, error) {
	return m.searchFilesFn(query, limit)
}
func (m *searchRepositoryMock) SearchFolders(query string, limit int) ([]FolderResultModel, error) {
	return m.searchFoldersFn(query, limit)
}
func (m *searchRepositoryMock) SearchArtists(query string, limit int) ([]ArtistResultModel, error) {
	return m.searchArtistsFn(query, limit)
}
func (m *searchRepositoryMock) SearchAlbums(query string, limit int) ([]AlbumResultModel, error) {
	return m.searchAlbumsFn(query, limit)
}
func (m *searchRepositoryMock) SearchMusicPlaylists(query string, limit int) ([]MusicPlaylistResultModel, error) {
	return m.searchMusicPlaylistsFn(query, limit)
}
func (m *searchRepositoryMock) SearchVideoPlaylists(query string, limit int) ([]VideoPlaylistResultModel, error) {
	return m.searchVideoPlaylistsFn(query, limit)
}
func (m *searchRepositoryMock) SearchVideos(query string, limit int) ([]VideoResultModel, error) {
	return m.searchVideosFn(query, limit)
}
func (m *searchRepositoryMock) SearchImages(query string, limit int) ([]ImageResultModel, error) {
	return m.searchImagesFn(query, limit)
}

type searchAIMock struct {
	executeFn func(ctx context.Context, req ai.Request) (ai.Response, error)
}

func (m *searchAIMock) Execute(ctx context.Context, req ai.Request) (ai.Response, error) {
	return m.executeFn(ctx, req)
}

func emptyRepo() *searchRepositoryMock {
	return &searchRepositoryMock{
		searchFilesFn:          func(string, int) ([]FileResultModel, error) { return nil, nil },
		searchFoldersFn:        func(string, int) ([]FolderResultModel, error) { return nil, nil },
		searchArtistsFn:        func(string, int) ([]ArtistResultModel, error) { return nil, nil },
		searchAlbumsFn:         func(string, int) ([]AlbumResultModel, error) { return nil, nil },
		searchMusicPlaylistsFn: func(string, int) ([]MusicPlaylistResultModel, error) { return nil, nil },
		searchVideoPlaylistsFn: func(string, int) ([]VideoPlaylistResultModel, error) { return nil, nil },
		searchVideosFn:         func(string, int) ([]VideoResultModel, error) { return nil, nil },
		searchImagesFn:         func(string, int) ([]ImageResultModel, error) { return nil, nil },
	}
}

func TestSearchServiceReturnsEmptyPayloadForBlankQuery(t *testing.T) {
	service := NewService(&searchRepositoryMock{
		searchFilesFn:   func(string, int) ([]FileResultModel, error) { t.Fatal("unexpected files call"); return nil, nil },
		searchFoldersFn: func(string, int) ([]FolderResultModel, error) { t.Fatal("unexpected folders call"); return nil, nil },
		searchArtistsFn: func(string, int) ([]ArtistResultModel, error) { t.Fatal("unexpected artists call"); return nil, nil },
		searchAlbumsFn:  func(string, int) ([]AlbumResultModel, error) { t.Fatal("unexpected albums call"); return nil, nil },
		searchMusicPlaylistsFn: func(string, int) ([]MusicPlaylistResultModel, error) {
			t.Fatal("unexpected playlists call")
			return nil, nil
		},
		searchVideoPlaylistsFn: func(string, int) ([]VideoPlaylistResultModel, error) {
			t.Fatal("unexpected video playlists call")
			return nil, nil
		},
		searchVideosFn: func(string, int) ([]VideoResultModel, error) { t.Fatal("unexpected videos call"); return nil, nil },
		searchImagesFn: func(string, int) ([]ImageResultModel, error) { t.Fatal("unexpected images call"); return nil, nil },
	}, nil)

	response, err := service.SearchGlobal("   ", 0)
	if err != nil {
		t.Fatalf("SearchGlobal returned error: %v", err)
	}

	if response.Query != "" {
		t.Fatalf("expected empty normalized query, got %q", response.Query)
	}
	if len(response.Files)+len(response.Folders)+len(response.Artists)+len(response.Albums)+len(response.Playlists)+len(response.Videos)+len(response.Images) != 0 {
		t.Fatalf("expected all result buckets to be empty: %+v", response)
	}
}

func TestSearchServiceMapsSearchBucketsAndClampsLimit(t *testing.T) {
	recordedLimits := []int{}
	recordLimit := func(limit int) {
		recordedLimits = append(recordedLimits, limit)
	}

	service := NewService(&searchRepositoryMock{
		searchFilesFn: func(query string, limit int) ([]FileResultModel, error) {
			recordLimit(limit)
			if query != "mix" {
				return nil, nil
			}
			return []FileResultModel{{ID: 1, Name: "song.mp3", Path: "/media/song.mp3", ParentPath: "/media", Format: ".mp3", Starred: true}}, nil
		},
		searchFoldersFn: func(string, int) ([]FolderResultModel, error) {
			recordLimit(maxSearchLimit)
			return []FolderResultModel{{ID: 2, Name: "Photos", Path: "/photos", ParentPath: "/", Starred: false}}, nil
		},
		searchArtistsFn: func(string, int) ([]ArtistResultModel, error) {
			recordLimit(maxSearchLimit)
			return []ArtistResultModel{{Artist: "AC/DC", TrackCount: 4, AlbumCount: 2}}, nil
		},
		searchAlbumsFn: func(string, int) ([]AlbumResultModel, error) {
			recordLimit(maxSearchLimit)
			return []AlbumResultModel{{Artist: "Miles Davis", Album: "Kind of Blue", Year: "1959", TrackCount: 5}}, nil
		},
		searchMusicPlaylistsFn: func(string, int) ([]MusicPlaylistResultModel, error) {
			recordLimit(maxSearchLimit)
			return []MusicPlaylistResultModel{{ID: 3, Name: "Morning", Description: "Focus", IsSystem: true, TrackCount: 8}}, nil
		},
		searchVideoPlaylistsFn: func(string, int) ([]VideoPlaylistResultModel, error) {
			recordLimit(maxSearchLimit)
			return []VideoPlaylistResultModel{{ID: 4, Name: "Severance", Type: "series", Classification: "series", SourcePath: "/videos/severance", IsAuto: true, ItemCount: 9}}, nil
		},
		searchVideosFn: func(string, int) ([]VideoResultModel, error) {
			recordLimit(maxSearchLimit)
			return []VideoResultModel{{ID: 5, Name: "Episode 01", Path: "/videos/episode-01.mkv", ParentPath: "/videos", Format: ".mkv"}}, nil
		},
		searchImagesFn: func(string, int) ([]ImageResultModel, error) {
			recordLimit(maxSearchLimit)
			return []ImageResultModel{{ID: 6, Name: "Vacation", Path: "/photos/vacation.jpg", ParentPath: "/photos", Format: ".jpg", Category: "photo", Context: "Canon"}}, nil
		},
	}, nil)

	response, err := service.SearchGlobal(" mix ", 99)
	if err != nil {
		t.Fatalf("SearchGlobal returned error: %v", err)
	}

	if response.Artists[0].Key != "ac/dc" {
		t.Fatalf("unexpected artist key: %+v", response.Artists[0])
	}
	if response.Albums[0].Key != "miles davis::kind of blue" {
		t.Fatalf("unexpected album key: %+v", response.Albums[0])
	}
	if len(response.Playlists) != 2 || response.Playlists[0].Scope != "music" || response.Playlists[1].Scope != "video" {
		t.Fatalf("unexpected playlists payload: %+v", response.Playlists)
	}
	if response.Images[0].Context != "Canon" || response.Files[0].Starred != true {
		t.Fatalf("unexpected mapped response: %+v", response)
	}
}

func TestSearchServiceStopsOnRepositoryError(t *testing.T) {
	service := NewService(&searchRepositoryMock{
		searchFilesFn: func(string, int) ([]FileResultModel, error) {
			return nil, errors.New("boom")
		},
		searchFoldersFn:        func(string, int) ([]FolderResultModel, error) { return nil, nil },
		searchArtistsFn:        func(string, int) ([]ArtistResultModel, error) { return nil, nil },
		searchAlbumsFn:         func(string, int) ([]AlbumResultModel, error) { return nil, nil },
		searchMusicPlaylistsFn: func(string, int) ([]MusicPlaylistResultModel, error) { return nil, nil },
		searchVideoPlaylistsFn: func(string, int) ([]VideoPlaylistResultModel, error) { return nil, nil },
		searchVideosFn:         func(string, int) ([]VideoResultModel, error) { return nil, nil },
		searchImagesFn:         func(string, int) ([]ImageResultModel, error) { return nil, nil },
	}, nil)

	if _, err := service.SearchGlobal("mix", 4); err == nil {
		t.Fatalf("expected repository error")
	}
}

func TestClampLimitAndNormalizeLookupKey(t *testing.T) {
	if got := clampLimit(-1); got != defaultSearchLimit {
		t.Fatalf("expected default limit, got %d", got)
	}
	if got := clampLimit(999); got != maxSearchLimit {
		t.Fatalf("expected max limit, got %d", got)
	}
	if got := normalizeLookupKey("  Álbum.Name-Test  "); got != "álbum name test" {
		t.Fatalf("unexpected normalized key %q", got)
	}
}

func TestSearchWithAIExpansionMergesResults(t *testing.T) {
	repo := emptyRepo()
	repo.searchFilesFn = func(query string, limit int) ([]FileResultModel, error) {
		if query == "my photos" {
			return []FileResultModel{{ID: 1, Name: "photo1.jpg"}}, nil
		}
		if query == "photos" {
			return []FileResultModel{{ID: 1, Name: "photo1.jpg"}, {ID: 2, Name: "photo2.jpg"}}, nil
		}
		return nil, nil
	}

	aiMock := &searchAIMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			return ai.Response{Content: `{"keywords": ["photos"], "suggestion": "Try searching by folder name"}`}, nil
		},
	}

	service := NewService(repo, aiMock)
	response, err := service.SearchGlobalWithAI("my photos", 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(response.Files) != 2 {
		t.Fatalf("expected 2 files (original + AI expanded), got %d", len(response.Files))
	}
	if response.Suggestion != "Try searching by folder name" {
		t.Fatalf("expected AI suggestion, got %q", response.Suggestion)
	}
}

func TestSearchWithAINilServiceSkipsExpansion(t *testing.T) {
	repo := emptyRepo()
	repo.searchFilesFn = func(string, int) ([]FileResultModel, error) {
		return []FileResultModel{{ID: 1, Name: "file.txt"}}, nil
	}

	service := NewService(repo, nil)
	response, err := service.SearchGlobalWithAI("my files", 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Suggestion != "" {
		t.Fatalf("expected no suggestion, got %q", response.Suggestion)
	}
}

func TestSearchWithAIErrorFallsBackGracefully(t *testing.T) {
	repo := emptyRepo()
	repo.searchFilesFn = func(string, int) ([]FileResultModel, error) {
		return []FileResultModel{{ID: 1, Name: "file.txt"}}, nil
	}

	aiMock := &searchAIMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			return ai.Response{}, errors.New("provider timeout")
		},
	}

	service := NewService(repo, aiMock)
	response, err := service.SearchGlobalWithAI("my files", 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(response.Files) != 1 {
		t.Fatalf("expected 1 file from original search, got %d", len(response.Files))
	}
}

func TestSearchWithAISingleWordSkipsExpansion(t *testing.T) {
	repo := emptyRepo()
	aiCalled := false
	aiMock := &searchAIMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			aiCalled = true
			return ai.Response{}, nil
		},
	}

	service := NewService(repo, aiMock)
	_, err := service.SearchGlobalWithAI("photo", 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if aiCalled {
		t.Fatalf("AI should not be called for single-word queries")
	}
}

func TestSearchWithAIInvalidJSONFallsBack(t *testing.T) {
	repo := emptyRepo()
	aiMock := &searchAIMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			return ai.Response{Content: "not json"}, nil
		},
	}

	service := NewService(repo, aiMock)
	response, err := service.SearchGlobalWithAI("my files", 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Suggestion != "" {
		t.Fatalf("expected no suggestion on parse error")
	}
}

func TestSearchGlobalBasePathNeverCallsAI(t *testing.T) {
	aiMock := &searchAIMock{
		executeFn: func(context.Context, ai.Request) (ai.Response, error) {
			t.Fatal("base search must not call the AI provider")
			return ai.Response{}, nil
		},
	}

	service := NewService(emptyRepo(), aiMock)
	if _, err := service.SearchGlobal("my photos from trip", 6); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearchWithAICachesExpansionPerNormalizedQuery(t *testing.T) {
	aiCalls := 0
	aiMock := &searchAIMock{
		executeFn: func(context.Context, ai.Request) (ai.Response, error) {
			aiCalls++
			return ai.Response{Content: `{"keywords": ["photos"], "suggestion": "tip"}`}, nil
		},
	}

	service := NewService(emptyRepo(), aiMock)
	for _, query := range []string{"my photos", "  My   Photos ", "my photos"} {
		response, err := service.SearchGlobalWithAI(query, 6)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if response.Suggestion != "tip" {
			t.Fatalf("expected cached suggestion, got %q", response.Suggestion)
		}
	}

	if aiCalls != 1 {
		t.Fatalf("expected a single AI call, got %d", aiCalls)
	}
}

func TestSearchWithAIDoesNotCacheFailures(t *testing.T) {
	aiCalls := 0
	aiMock := &searchAIMock{
		executeFn: func(context.Context, ai.Request) (ai.Response, error) {
			aiCalls++
			return ai.Response{}, errors.New("provider timeout")
		},
	}

	service := NewService(emptyRepo(), aiMock)
	_, _ = service.SearchGlobalWithAI("my files", 6)
	_, _ = service.SearchGlobalWithAI("my files", 6)

	if aiCalls != 2 {
		t.Fatalf("expected failures to be retried, got %d calls", aiCalls)
	}
}

func TestSearchWithAITrimsMergedGroupsToLimitAndKeywordCount(t *testing.T) {
	searchedKeywords := []string{}
	repo := emptyRepo()
	repo.searchFilesFn = func(query string, limit int) ([]FileResultModel, error) {
		searchedKeywords = append(searchedKeywords, query)
		switch query {
		case "my photos":
			return []FileResultModel{{ID: 1}}, nil
		default:
			return []FileResultModel{{ID: 10 + len(searchedKeywords)}, {ID: 20 + len(searchedKeywords)}}, nil
		}
	}
	aiMock := &searchAIMock{
		executeFn: func(context.Context, ai.Request) (ai.Response, error) {
			return ai.Response{Content: "```json\n{\"keywords\": [\"a\", \"b\", \"c\", \"d\"], \"suggestion\": \"tip\"}\n```"}, nil
		},
	}

	service := NewService(repo, aiMock)
	response, err := service.SearchGlobalWithAI("my photos", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(response.Files) != 3 {
		t.Fatalf("expected files trimmed to limit 3, got %d", len(response.Files))
	}
	if response.Files[0].ID != 1 {
		t.Fatalf("expected base results first, got %+v", response.Files)
	}
	if len(searchedKeywords) != 1+aiMaxKeywords {
		t.Fatalf("expected base query plus %d keywords, got %v", aiMaxKeywords, searchedKeywords)
	}
}

func TestExpansionCacheExpiresAndBoundsSize(t *testing.T) {
	currentTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cache := newExpansionCache(2, time.Minute)
	cache.now = func() time.Time { return currentTime }

	cache.put("one", aiSearchExpansion{Suggestion: "1"})
	currentTime = currentTime.Add(10 * time.Second)
	cache.put("two", aiSearchExpansion{Suggestion: "2"})
	currentTime = currentTime.Add(10 * time.Second)
	cache.put("three", aiSearchExpansion{Suggestion: "3"})

	if _, isCached := cache.get("one"); isCached {
		t.Fatal("expected oldest entry to be evicted at capacity")
	}
	if _, isCached := cache.get("three"); !isCached {
		t.Fatal("expected newest entry to be cached")
	}

	currentTime = currentTime.Add(2 * time.Minute)
	if _, isCached := cache.get("three"); isCached {
		t.Fatal("expected entry to expire after ttl")
	}
}
