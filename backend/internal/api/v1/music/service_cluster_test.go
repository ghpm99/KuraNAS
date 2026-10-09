package music

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	"nas-go/api/pkg/ai"
)

type fakeAIService struct {
	executeFn func(ctx context.Context, req ai.Request) (ai.Response, error)
	calls     int
}

func (f *fakeAIService) Execute(ctx context.Context, req ai.Request) (ai.Response, error) {
	f.calls++
	if f.executeFn != nil {
		return f.executeFn(ctx, req)
	}
	return ai.Response{}, nil
}

func clusterTestInputs() []artistClusterInput {
	return []artistClusterInput{
		{Key: "the beatles", Artist: "The Beatles", GenreHint: "Rock", TrackCount: 2},
		{Key: "johnny cash", Artist: "Johnny Cash", GenreHint: "Country", TrackCount: 1},
		{Key: "nightwish", Artist: "Nightwish", GenreHint: "Rock", TrackCount: 1},
	}
}

func clusterTestFileIDsByArtistKeys(artistKeys []string) ([]int, error) {
	fileIDsByArtistKey := map[string][]int{
		"the beatles": {1, 2},
		"nightwish":   {3},
		"johnny cash": {4},
	}
	sortedKeys := append([]string(nil), artistKeys...)
	sort.Strings(sortedKeys)
	fileIDs := []int{}
	for _, artistKey := range sortedKeys {
		fileIDs = append(fileIDs, fileIDsByArtistKey[artistKey]...)
	}
	return fileIDs, nil
}

func TestLoadClusterTrackIDs(t *testing.T) {
	repo := &musicRepoMock{getFileIDsByArtistsFn: clusterTestFileIDsByArtistKeys}
	svc := newMusicServiceForTest(t, repo)
	mapping := map[string]string{
		"the beatles": "Classic Rock",
		"nightwish":   "Metal",
		"johnny cash": "",
	}

	tracks, err := svc.loadClusterTrackIDs(mapping)
	if err != nil {
		t.Fatalf("loadClusterTrackIDs failed: %v", err)
	}
	if _, ok := tracks[""]; ok {
		t.Fatalf("empty cluster name must be skipped")
	}
	if !reflect.DeepEqual(tracks["Classic Rock"], []int{1, 2}) {
		t.Fatalf("unexpected Classic Rock tracks: %v", tracks["Classic Rock"])
	}
	if !reflect.DeepEqual(tracks["Metal"], []int{3}) {
		t.Fatalf("unexpected Metal tracks: %v", tracks["Metal"])
	}
}

func TestLoadClusterTrackIDsPropagatesError(t *testing.T) {
	repo := &musicRepoMock{getFileIDsByArtistsFn: func([]string) ([]int, error) { return nil, errors.New("boom") }}
	svc := newMusicServiceForTest(t, repo)
	if _, err := svc.loadClusterTrackIDs(map[string]string{"a": "Rock"}); err == nil {
		t.Fatalf("expected error to propagate")
	}
}

func TestUniqueFileIDsKeepsFirstOccurrenceOrder(t *testing.T) {
	if got := uniqueFileIDs([]int{3, 1, 3, 2, 1}); !reflect.DeepEqual(got, []int{3, 1, 2}) {
		t.Fatalf("unexpected unique ids: %v", got)
	}
}

func TestDistinctClusterNames(t *testing.T) {
	names := distinctClusterNames(map[string]string{
		"a": "Rock", "b": "rock", "c": "Metal", "d": "",
	})
	// Case-insensitive de-dup of "Rock"/"rock" leaves two names, empty dropped.
	if len(names) != 2 {
		t.Fatalf("expected 2 distinct names, got %v", names)
	}
	lowered := []string{strings.ToLower(names[0]), strings.ToLower(names[1])}
	if !reflect.DeepEqual(lowered, []string{"metal", "rock"}) {
		t.Fatalf("unexpected distinct names: %v", names)
	}
}

func TestRebuildAIClustersEndToEnd(t *testing.T) {
	var upserts []ArtistClusterModel
	var prunedKeys []string
	created := map[string]int{}
	replaced := map[int][]int{}
	nextID := 100

	repo := &musicRepoMock{
		getClusterInputsFn:    func() ([]artistClusterInput, error) { return clusterTestInputs(), nil },
		getFileIDsByArtistsFn: clusterTestFileIDsByArtistKeys,
		getArtistClustersFn: func() ([]ArtistClusterModel, error) {
			return []ArtistClusterModel{}, nil
		},
		upsertArtistClusterFn: func(_ *sql.Tx, cluster ArtistClusterModel) error {
			upserts = append(upserts, cluster)
			return nil
		},
		deleteArtistClustersExceptFn: func(_ *sql.Tx, keys []string) error {
			prunedKeys = keys
			return nil
		},
		getAIPlaylistsFn: func() ([]PlaylistModel, error) {
			return []PlaylistModel{}, nil
		},
		createAIPlaylistFn: func(_ *sql.Tx, name string, _ string) (PlaylistModel, error) {
			nextID++
			created[name] = nextID
			return PlaylistModel{ID: nextID, Name: name, IsAIGenerated: true}, nil
		},
		replacePlaylistTracksFn: func(_ *sql.Tx, playlistID int, fileIDs []int) error {
			replaced[playlistID] = fileIDs
			return nil
		},
	}

	svc := newMusicServiceForTest(t, repo)
	svc.AIService = &fakeAIService{executeFn: func(_ context.Context, _ ai.Request) (ai.Response, error) {
		return ai.Response{Content: `{"clusters":[
			{"name":"Classic Rock","artists":["The Beatles"]},
			{"name":"Symphonic Metal","artists":["Nightwish"]},
			{"name":"Country","artists":["Johnny Cash"]}
		]}`}, nil
	}}

	if err := svc.RebuildAIClusters(context.Background()); err != nil {
		t.Fatalf("RebuildAIClusters failed: %v", err)
	}

	if len(upserts) != 3 {
		t.Fatalf("expected 3 artist upserts, got %d", len(upserts))
	}
	if len(created) != 3 {
		t.Fatalf("expected 3 AI playlists created, got %d", len(created))
	}
	if got := replaced[created["Classic Rock"]]; !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("Classic Rock should contain both Beatles tracks, got %v", got)
	}

	sort.Strings(prunedKeys)
	want := []string{normalizeLookupKey("Johnny Cash"), normalizeLookupKey("Nightwish"), normalizeLookupKey("The Beatles")}
	sort.Strings(want)
	if !reflect.DeepEqual(prunedKeys, want) {
		t.Fatalf("prune should keep all current artists, got %v", prunedKeys)
	}
}

func TestRebuildAIClustersIncrementalSkipsKnownArtists(t *testing.T) {
	deletedPlaylists := []int{}
	repo := &musicRepoMock{
		getClusterInputsFn:    func() ([]artistClusterInput, error) { return clusterTestInputs(), nil },
		getFileIDsByArtistsFn: clusterTestFileIDsByArtistKeys,
		getArtistClustersFn: func() ([]ArtistClusterModel, error) {
			// Every current artist is already mapped, so no AI call is needed.
			return []ArtistClusterModel{
				{ArtistKey: normalizeLookupKey("The Beatles"), Artist: "The Beatles", ClusterName: "Classic Rock"},
				{ArtistKey: normalizeLookupKey("Nightwish"), Artist: "Nightwish", ClusterName: "Metal"},
				{ArtistKey: normalizeLookupKey("Johnny Cash"), Artist: "Johnny Cash", ClusterName: "Country"},
			}, nil
		},
		getAIPlaylistsFn: func() ([]PlaylistModel, error) {
			return []PlaylistModel{
				{ID: 1, Name: "Classic Rock", IsAIGenerated: true},
				{ID: 2, Name: "Metal", IsAIGenerated: true},
				{ID: 3, Name: "Country", IsAIGenerated: true},
				{ID: 4, Name: "Stale Cluster", IsAIGenerated: true},
			}, nil
		},
		createAIPlaylistFn: func(_ *sql.Tx, name string, _ string) (PlaylistModel, error) {
			t.Fatalf("should not create playlists when all exist (%s)", name)
			return PlaylistModel{}, nil
		},
		replacePlaylistTracksFn: func(_ *sql.Tx, _ int, _ []int) error { return nil },
		deletePlaylistFn: func(_ *sql.Tx, id int) error {
			deletedPlaylists = append(deletedPlaylists, id)
			return nil
		},
	}

	ai := &fakeAIService{}
	svc := newMusicServiceForTest(t, repo)
	svc.AIService = ai

	if err := svc.RebuildAIClusters(context.Background()); err != nil {
		t.Fatalf("RebuildAIClusters failed: %v", err)
	}
	if ai.calls != 0 {
		t.Fatalf("expected no AI calls when nothing is new, got %d", ai.calls)
	}
	if !reflect.DeepEqual(deletedPlaylists, []int{4}) {
		t.Fatalf("stale AI playlist should be deleted, got %v", deletedPlaylists)
	}
}

func TestRebuildAIClustersWithoutAIServiceIsNoop(t *testing.T) {
	repo := &musicRepoMock{
		getClusterInputsFn:    func() ([]artistClusterInput, error) { return clusterTestInputs(), nil },
		getFileIDsByArtistsFn: clusterTestFileIDsByArtistKeys,
		getArtistClustersFn:   func() ([]ArtistClusterModel, error) { return []ArtistClusterModel{}, nil },
		createAIPlaylistFn: func(_ *sql.Tx, name string, _ string) (PlaylistModel, error) {
			t.Fatalf("no playlist should be created without an AI service")
			return PlaylistModel{}, nil
		},
	}

	svc := newMusicServiceForTest(t, repo) // AIService stays nil
	if err := svc.RebuildAIClusters(context.Background()); err != nil {
		t.Fatalf("expected no-op, got error: %v", err)
	}
}

func TestRebuildAIClustersPropagatesIndexError(t *testing.T) {
	repo := &musicRepoMock{
		getClusterInputsFn: func() ([]artistClusterInput, error) {
			return nil, errors.New("boom")
		},
	}
	svc := newMusicServiceForTest(t, repo)
	if err := svc.RebuildAIClusters(context.Background()); err == nil {
		t.Fatalf("expected index error to propagate")
	}
}
