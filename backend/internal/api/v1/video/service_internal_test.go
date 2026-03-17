package video

import (
	"database/sql"
	"errors"
	"nas-go/api/internal/api/v1/video/playlist"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
	"testing"
	"time"
)

type videoRepoMock struct {
	db *database.DbContext

	getVideoFileByIDFn         func(id int) (VideoFileModel, error)
	getVideosByParentPathFn    func(parentPath string) ([]VideoFileModel, error)
	getPlaylistByContextFn     func(contextType string, sourcePath string) (VideoPlaylistModel, error)
	createPlaylistFn           func(tx *sql.Tx, contextType string, sourcePath string) (VideoPlaylistModel, error)
	replacePlaylistItemsFn     func(tx *sql.Tx, playlistID int, videoIDs []int) error
	getPlaybackStateFn         func(clientID string) (VideoPlaybackStateModel, error)
	upsertPlaybackStateFn      func(tx *sql.Tx, state VideoPlaybackStateModel) (VideoPlaybackStateModel, error)
	touchPlaylistFn            func(tx *sql.Tx, playlistID int) error
	getCatalogVideosFn         func(limit int) ([]VideoFileModel, error)
	getRecentVideosFn          func(limit int) ([]VideoFileModel, error)
	checkVideoInPlaylistFn     func(playlistID int, videoID int) (bool, error)
	getUnassignedVideosFn      func(limit int) ([]VideoFileModel, error)
	getVideoPlaylistsFn        func(includeHidden bool) ([]VideoPlaylistModel, error)
	getVideoPlaylistMembersFn  func(includeHidden bool) ([]VideoPlaylistMembershipModel, error)
	getVideoPlaylistByIDFn     func(id int) (VideoPlaylistModel, error)
	getVideoPlaylistItemsFn    func(playlistID int) ([]VideoPlaylistItemModel, error)
	listLibraryVideosFn        func(page int, pageSize int, searchQuery string) (utils.PaginationResponse[VideoFileModel], error)
	setPlaylistHiddenFn        func(tx *sql.Tx, playlistID int, hidden bool) error
	addPlaylistVideoManualFn   func(tx *sql.Tx, playlistID int, videoID int) error
	deletePlaylistExclusionFn  func(tx *sql.Tx, playlistID int, videoID int) error
	removePlaylistVideoFn      func(tx *sql.Tx, playlistID int, videoID int) error
	upsertPlaylistExclusionFn  func(tx *sql.Tx, playlistID int, videoID int) error
	updatePlaylistNameFn       func(tx *sql.Tx, playlistID int, name string) error
	reorderPlaylistItemFn      func(tx *sql.Tx, playlistID int, videoID int, orderIndex int) error
	getAllVideosForGroupingFn  func() ([]VideoFileModel, error)
	getAllVideosWithMetadataFn func() ([]VideoWithMetadataModel, error)
	upsertAutoPlaylistFn       func(tx *sql.Tx, contextType, sourcePath, name, groupMode, classification string) (VideoPlaylistModel, error)
	getPlaylistExclusionsFn    func(playlistID int) (map[int]bool, error)
	deleteAutoPlaylistItemsFn  func(tx *sql.Tx, playlistID int) error
	insertPlaylistItemsSrcFn   func(tx *sql.Tx, playlistID int, videoIDs []int, sourceKind string) error
	insertBehaviorEventFn      func(tx *sql.Tx, event VideoBehaviorEventModel) (VideoBehaviorEventModel, error)
	getBehaviorEventsFn        func(clientID string, limit int) ([]VideoBehaviorEventModel, error)
	getAllBehaviorEventsFn     func(limit int) ([]VideoBehaviorEventModel, error)
}

func (m *videoRepoMock) GetDbContext() *database.DbContext { return m.db }
func (m *videoRepoMock) GetVideoFileByID(id int) (VideoFileModel, error) {
	if m.getVideoFileByIDFn != nil {
		return m.getVideoFileByIDFn(id)
	}
	return VideoFileModel{}, errors.New("not used")
}
func (m *videoRepoMock) GetVideosByParentPath(parentPath string) ([]VideoFileModel, error) {
	if m.getVideosByParentPathFn != nil {
		return m.getVideosByParentPathFn(parentPath)
	}
	return nil, errors.New("not used")
}
func (m *videoRepoMock) GetPlaylistByContext(contextType string, sourcePath string) (VideoPlaylistModel, error) {
	if m.getPlaylistByContextFn != nil {
		return m.getPlaylistByContextFn(contextType, sourcePath)
	}
	return VideoPlaylistModel{}, errors.New("not used")
}
func (m *videoRepoMock) CreatePlaylist(tx *sql.Tx, contextType string, sourcePath string) (VideoPlaylistModel, error) {
	if m.createPlaylistFn != nil {
		return m.createPlaylistFn(tx, contextType, sourcePath)
	}
	return VideoPlaylistModel{}, errors.New("not used")
}
func (m *videoRepoMock) ReplacePlaylistItems(tx *sql.Tx, playlistID int, videoIDs []int) error {
	if m.replacePlaylistItemsFn != nil {
		return m.replacePlaylistItemsFn(tx, playlistID, videoIDs)
	}
	return errors.New("not used")
}
func (m *videoRepoMock) GetPlaylistItems(playlistID int) ([]VideoPlaylistItemModel, error) {
	if m.getVideoPlaylistItemsFn != nil {
		return m.getVideoPlaylistItemsFn(playlistID)
	}
	return nil, nil
}
func (m *videoRepoMock) GetPlaybackState(clientID string) (VideoPlaybackStateModel, error) {
	if m.getPlaybackStateFn != nil {
		return m.getPlaybackStateFn(clientID)
	}
	return VideoPlaybackStateModel{}, errors.New("not used")
}
func (m *videoRepoMock) UpsertPlaybackState(tx *sql.Tx, state VideoPlaybackStateModel) (VideoPlaybackStateModel, error) {
	if m.upsertPlaybackStateFn != nil {
		return m.upsertPlaybackStateFn(tx, state)
	}
	return VideoPlaybackStateModel{}, errors.New("not used")
}
func (m *videoRepoMock) TouchPlaylist(tx *sql.Tx, playlistID int) error {
	if m.touchPlaylistFn != nil {
		return m.touchPlaylistFn(tx, playlistID)
	}
	return nil
}
func (m *videoRepoMock) GetCatalogVideos(limit int) ([]VideoFileModel, error) {
	if m.getCatalogVideosFn != nil {
		return m.getCatalogVideosFn(limit)
	}
	return nil, errors.New("not used")
}
func (m *videoRepoMock) GetRecentVideos(limit int) ([]VideoFileModel, error) {
	if m.getRecentVideosFn != nil {
		return m.getRecentVideosFn(limit)
	}
	return nil, errors.New("not used")
}
func (m *videoRepoMock) GetAllVideosForGrouping() ([]VideoFileModel, error) {
	if m.getAllVideosForGroupingFn != nil {
		return m.getAllVideosForGroupingFn()
	}
	return nil, nil
}
func (m *videoRepoMock) UpsertAutoPlaylist(tx *sql.Tx, contextType, sourcePath, name, groupMode, classification string) (VideoPlaylistModel, error) {
	if m.upsertAutoPlaylistFn != nil {
		return m.upsertAutoPlaylistFn(tx, contextType, sourcePath, name, groupMode, classification)
	}
	return VideoPlaylistModel{}, nil
}
func (m *videoRepoMock) DeleteAutoPlaylistItems(tx *sql.Tx, playlistID int) error {
	if m.deleteAutoPlaylistItemsFn != nil {
		return m.deleteAutoPlaylistItemsFn(tx, playlistID)
	}
	return nil
}
func (m *videoRepoMock) InsertPlaylistItemsWithSource(tx *sql.Tx, playlistID int, videoIDs []int, sourceKind string) error {
	if m.insertPlaylistItemsSrcFn != nil {
		return m.insertPlaylistItemsSrcFn(tx, playlistID, videoIDs, sourceKind)
	}
	return nil
}
func (m *videoRepoMock) GetPlaylistExclusions(playlistID int) (map[int]bool, error) {
	if m.getPlaylistExclusionsFn != nil {
		return m.getPlaylistExclusionsFn(playlistID)
	}
	return map[int]bool{}, nil
}
func (m *videoRepoMock) GetVideoPlaylists(includeHidden bool) ([]VideoPlaylistModel, error) {
	if m.getVideoPlaylistsFn != nil {
		return m.getVideoPlaylistsFn(includeHidden)
	}
	return nil, nil
}
func (m *videoRepoMock) GetVideoPlaylistMemberships(includeHidden bool) ([]VideoPlaylistMembershipModel, error) {
	if m.getVideoPlaylistMembersFn != nil {
		return m.getVideoPlaylistMembersFn(includeHidden)
	}
	return nil, nil
}
func (m *videoRepoMock) GetVideoPlaylistByID(id int) (VideoPlaylistModel, error) {
	if m.getVideoPlaylistByIDFn != nil {
		return m.getVideoPlaylistByIDFn(id)
	}
	return VideoPlaylistModel{}, nil
}
func (m *videoRepoMock) GetVideoPlaylistItemsDetailed(playlistID int) ([]VideoPlaylistItemModel, error) {
	if m.getVideoPlaylistItemsFn != nil {
		return m.getVideoPlaylistItemsFn(playlistID)
	}
	return nil, nil
}
func (m *videoRepoMock) ListLibraryVideos(page int, pageSize int, searchQuery string) (utils.PaginationResponse[VideoFileModel], error) {
	if m.listLibraryVideosFn != nil {
		return m.listLibraryVideosFn(page, pageSize, searchQuery)
	}
	return utils.PaginationResponse[VideoFileModel]{}, nil
}
func (m *videoRepoMock) SetPlaylistHidden(tx *sql.Tx, playlistID int, hidden bool) error {
	if m.setPlaylistHiddenFn != nil {
		return m.setPlaylistHiddenFn(tx, playlistID, hidden)
	}
	return nil
}
func (m *videoRepoMock) AddPlaylistVideoManual(tx *sql.Tx, playlistID int, videoID int) error {
	if m.addPlaylistVideoManualFn != nil {
		return m.addPlaylistVideoManualFn(tx, playlistID, videoID)
	}
	return nil
}
func (m *videoRepoMock) RemovePlaylistVideo(tx *sql.Tx, playlistID int, videoID int) error {
	if m.removePlaylistVideoFn != nil {
		return m.removePlaylistVideoFn(tx, playlistID, videoID)
	}
	return nil
}
func (m *videoRepoMock) UpsertPlaylistExclusion(tx *sql.Tx, playlistID int, videoID int) error {
	if m.upsertPlaylistExclusionFn != nil {
		return m.upsertPlaylistExclusionFn(tx, playlistID, videoID)
	}
	return nil
}
func (m *videoRepoMock) DeletePlaylistExclusion(tx *sql.Tx, playlistID int, videoID int) error {
	if m.deletePlaylistExclusionFn != nil {
		return m.deletePlaylistExclusionFn(tx, playlistID, videoID)
	}
	return nil
}
func (m *videoRepoMock) GetUnassignedVideos(limit int) ([]VideoFileModel, error) {
	if m.getUnassignedVideosFn != nil {
		return m.getUnassignedVideosFn(limit)
	}
	return nil, nil
}
func (m *videoRepoMock) CheckVideoInPlaylist(playlistID int, videoID int) (bool, error) {
	if m.checkVideoInPlaylistFn != nil {
		return m.checkVideoInPlaylistFn(playlistID, videoID)
	}
	return false, errors.New("not used")
}
func (m *videoRepoMock) UpdatePlaylistName(tx *sql.Tx, playlistID int, name string) error {
	if m.updatePlaylistNameFn != nil {
		return m.updatePlaylistNameFn(tx, playlistID, name)
	}
	return nil
}
func (m *videoRepoMock) ReorderPlaylistItem(tx *sql.Tx, playlistID int, videoID int, orderIndex int) error {
	if m.reorderPlaylistItemFn != nil {
		return m.reorderPlaylistItemFn(tx, playlistID, videoID, orderIndex)
	}
	return nil
}
func (m *videoRepoMock) GetAllVideosWithMetadata() ([]VideoWithMetadataModel, error) {
	if m.getAllVideosWithMetadataFn != nil {
		return m.getAllVideosWithMetadataFn()
	}
	return nil, nil
}
func (m *videoRepoMock) InsertBehaviorEvent(tx *sql.Tx, event VideoBehaviorEventModel) (VideoBehaviorEventModel, error) {
	if m.insertBehaviorEventFn != nil {
		return m.insertBehaviorEventFn(tx, event)
	}
	return event, nil
}
func (m *videoRepoMock) GetBehaviorEvents(clientID string, limit int) ([]VideoBehaviorEventModel, error) {
	if m.getBehaviorEventsFn != nil {
		return m.getBehaviorEventsFn(clientID, limit)
	}
	return nil, nil
}
func (m *videoRepoMock) GetAllBehaviorEvents(limit int) ([]VideoBehaviorEventModel, error) {
	if m.getAllBehaviorEventsFn != nil {
		return m.getAllBehaviorEventsFn(limit)
	}
	return nil, nil
}

func newVideoServiceForTest(t *testing.T, repo *videoRepoMock) *Service {
	t.Helper()
	repo.db = database.NewDbContext(nil)
	return &Service{Repository: repo, PlaylistEngine: playlist.NewPlaylistEngine()}
}

func TestVideoHelpersClassificationAndGrouping(t *testing.T) {
	classifier := playlist.NewVideoClassifier()

	seriesResult := classifier.Classify(playlist.VideoEntry{Name: "S01E02 episode", ParentPath: "/series/show", Path: "/series/show/S01E02 episode"})
	if seriesResult.Classification != playlist.ClassSeries {
		t.Fatalf("expected series classification, got %s", seriesResult.Classification)
	}

	movieResult := classifier.Classify(playlist.VideoEntry{Name: "Movie", ParentPath: "/movies", Path: "/movies/Movie"})
	if movieResult.Classification != playlist.ClassMovie {
		t.Fatalf("expected movie classification, got %s", movieResult.Classification)
	}

	personalResult := classifier.Classify(playlist.VideoEntry{Name: "Family Video", ParentPath: "/personal", Path: "/personal/Family Video"})
	if personalResult.Classification != playlist.ClassPersonal {
		t.Fatalf("expected personal classification, got %s", personalResult.Classification)
	}

	if got := playlist.InferTitlePrefix("My.Show.S01E02.mkv"); got == "" {
		t.Fatalf("expected inferred title prefix")
	}
	if !playlist.IsGenericFolderName("videos") || playlist.IsGenericFolderName("myfolder") {
		t.Fatalf("generic folder detection mismatch")
	}

	// "tutorial" classifica como course (mais especifico que program no novo classifier)
	courseResult := classifier.Classify(playlist.VideoEntry{Name: "tutorial", ParentPath: "/x", Path: "/x/tutorial"})
	if courseResult.Classification != playlist.ClassCourse {
		t.Fatalf("expected course classification, got %s", courseResult.Classification)
	}

	// "steam" keyword classifica como program
	programResult := classifier.Classify(playlist.VideoEntry{Name: "gameplay", ParentPath: "/steam", Path: "/steam/gameplay"})
	if programResult.Classification != playlist.ClassProgram {
		t.Fatalf("expected program classification, got %s", programResult.Classification)
	}

	engine := playlist.NewPlaylistEngine()
	result := engine.Build(playlist.BuildInput{
		Videos: []playlist.VideoEntry{
			{ID: 1, Name: "Show S01E01.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E01.mkv"},
			{ID: 2, Name: "Show S01E02.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E02.mkv"},
			{ID: 3, Name: "Movie.mkv", ParentPath: "/movies", Path: "/movies/Movie.mkv"},
		},
	})
	if len(result.Candidates) == 0 {
		t.Fatalf("expected smart groups to be built")
	}
}

func TestVideoServiceWrappersAndValidations(t *testing.T) {
	repo := &videoRepoMock{
		getUnassignedVideosFn: func(limit int) ([]VideoFileModel, error) {
			return []VideoFileModel{{ID: 1, Name: "v", ParentPath: "/p", Path: "/p/v.mp4", Format: ".mp4"}}, nil
		},
		getVideoPlaylistsFn: func(includeHidden bool) ([]VideoPlaylistModel, error) {
			return []VideoPlaylistModel{{ID: 10, Name: "p1"}}, nil
		},
		getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
			return VideoPlaylistModel{ID: id, Name: "playlist"}, nil
		},
		getVideoPlaylistItemsFn: func(playlistID int) ([]VideoPlaylistItemModel, error) {
			return []VideoPlaylistItemModel{
				{ID: 1, PlaylistID: playlistID, VideoID: 1, OrderIndex: 0, Video: VideoFileModel{ID: 1, Name: "a", Path: "/a", ParentPath: "/"}},
			}, nil
		},
		setPlaylistHiddenFn:      func(tx *sql.Tx, playlistID int, hidden bool) error { return nil },
		addPlaylistVideoManualFn: func(tx *sql.Tx, playlistID int, videoID int) error { return nil },
		deletePlaylistExclusionFn: func(tx *sql.Tx, playlistID int, videoID int) error {
			return nil
		},
		removePlaylistVideoFn:     func(tx *sql.Tx, playlistID int, videoID int) error { return nil },
		upsertPlaylistExclusionFn: func(tx *sql.Tx, playlistID int, videoID int) error { return nil },
		updatePlaylistNameFn:      func(tx *sql.Tx, playlistID int, name string) error { return nil },
		reorderPlaylistItemFn:     func(tx *sql.Tx, playlistID int, videoID int, orderIndex int) error { return nil },
	}
	svc := newVideoServiceForTest(t, repo)

	if _, err := svc.GetUnassignedVideos(0); err != nil {
		t.Fatalf("expected unassigned videos success, err=%v", err)
	}
	if _, err := svc.GetPlaylists(true); err != nil {
		t.Fatalf("expected get playlists success, err=%v", err)
	}
	if _, err := svc.GetPlaylistByID("client-1", 10); err != nil {
		t.Fatalf("expected get playlist by id success, err=%v", err)
	}
	if err := svc.SetPlaylistHidden(10, true); err != nil {
		t.Fatalf("expected set hidden success, err=%v", err)
	}
	if err := svc.AddVideoToPlaylist(10, 2); err != nil {
		t.Fatalf("expected add video success, err=%v", err)
	}
	if err := svc.RemoveVideoFromPlaylist(10, 2); err != nil {
		t.Fatalf("expected remove video success, err=%v", err)
	}
	if err := svc.UpdatePlaylistName(10, "  renamed  "); err != nil {
		t.Fatalf("expected update name success, err=%v", err)
	}
	if err := svc.ReorderPlaylistItems(10, []ReorderPlaylistItemRequest{{VideoID: 1, OrderIndex: 0}}); err != nil {
		t.Fatalf("expected reorder success, err=%v", err)
	}

	if err := svc.UpdatePlaylistName(10, "   "); err == nil {
		t.Fatalf("expected empty name validation error")
	}
	if err := svc.ReorderPlaylistItems(10, nil); err == nil {
		t.Fatalf("expected empty reorder payload error")
	}
	if err := svc.ReorderPlaylistItems(10, []ReorderPlaylistItemRequest{{VideoID: 1, OrderIndex: 0}, {VideoID: 1, OrderIndex: 1}}); err == nil {
		t.Fatalf("expected duplicated video id reorder error")
	}
	if err := svc.ReorderPlaylistItems(10, []ReorderPlaylistItemRequest{{VideoID: 1, OrderIndex: 0}, {VideoID: 2, OrderIndex: 0}}); err == nil {
		t.Fatalf("expected duplicated order index reorder error")
	}
}

func TestGetPlaylistByIDUsesPlaybackAndBehaviorProgress(t *testing.T) {
	repo := &videoRepoMock{
		getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
			return VideoPlaylistModel{ID: id, Name: "Show", Classification: "series"}, nil
		},
		getVideoPlaylistItemsFn: func(playlistID int) ([]VideoPlaylistItemModel, error) {
			return []VideoPlaylistItemModel{
				{
					ID:         1,
					PlaylistID: playlistID,
					VideoID:    10,
					OrderIndex: 0,
					SourceKind: "auto",
					Video:      VideoFileModel{ID: 10, Name: "Show S01E01.mkv", Path: "/series/show/Show S01E01.mkv", ParentPath: "/series/show"},
				},
				{
					ID:         2,
					PlaylistID: playlistID,
					VideoID:    11,
					OrderIndex: 1,
					SourceKind: "auto",
					Video:      VideoFileModel{ID: 11, Name: "Show S01E02.mkv", Path: "/series/show/Show S01E02.mkv", ParentPath: "/series/show"},
				},
			}, nil
		},
		getPlaybackStateFn: func(clientID string) (VideoPlaybackStateModel, error) {
			return VideoPlaybackStateModel{
				ClientID:    clientID,
				VideoID:     sql.NullInt64{Int64: 11, Valid: true},
				CurrentTime: 60,
				Duration:    120,
			}, nil
		},
		getBehaviorEventsFn: func(clientID string, limit int) ([]VideoBehaviorEventModel, error) {
			return []VideoBehaviorEventModel{
				{VideoID: 10, EventType: string(playlist.EventCompleted), WatchedPct: 100},
			}, nil
		},
	}

	svc := newVideoServiceForTest(t, repo)
	detail, err := svc.GetPlaylistByID("client-1", 7)
	if err != nil {
		t.Fatalf("expected playlist detail success, err=%v", err)
	}
	if len(detail.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(detail.Items))
	}

	if detail.Items[0].Status != "completed" || detail.Items[0].ProgressPct != 100 {
		t.Fatalf("expected first item completed, got status=%s pct=%.1f", detail.Items[0].Status, detail.Items[0].ProgressPct)
	}
	if detail.Items[1].Status != "in_progress" || detail.Items[1].ProgressPct != 50 {
		t.Fatalf("expected second item in progress at 50%%, got status=%s pct=%.1f", detail.Items[1].Status, detail.Items[1].ProgressPct)
	}
}

func TestVideoServiceRebuildSmartPlaylists(t *testing.T) {
	repo := &videoRepoMock{
		getAllVideosWithMetadataFn: func() ([]VideoWithMetadataModel, error) {
			return []VideoWithMetadataModel{
				{VideoFileModel: VideoFileModel{ID: 1, Name: "Show S01E01.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E01.mkv"}},
				{VideoFileModel: VideoFileModel{ID: 2, Name: "Show S01E02.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E02.mkv"}},
			}, nil
		},
		upsertAutoPlaylistFn: func(tx *sql.Tx, contextType, sourcePath, name, groupMode, classification string) (VideoPlaylistModel, error) {
			return VideoPlaylistModel{ID: 100, Name: name}, nil
		},
		getPlaylistExclusionsFn: func(playlistID int) (map[int]bool, error) {
			return map[int]bool{2: true}, nil
		},
		deleteAutoPlaylistItemsFn: func(tx *sql.Tx, playlistID int) error { return nil },
		insertPlaylistItemsSrcFn:  func(tx *sql.Tx, playlistID int, videoIDs []int, sourceKind string) error { return nil },
	}
	svc := newVideoServiceForTest(t, repo)
	if err := svc.RebuildSmartPlaylists(); err != nil {
		t.Fatalf("expected rebuild smart playlists success, err=%v", err)
	}
}

func TestToCatalogItem(t *testing.T) {
	now := time.Now()
	svc := &Service{}
	video := VideoFileModel{ID: 1, Name: "v", Path: "/v", ParentPath: "/", Format: ".mp4", UpdatedAt: now, CreatedAt: now}

	state := VideoPlaybackStateModel{
		VideoID:     sql.NullInt64{Int64: 1, Valid: true},
		CurrentTime: 10,
		Duration:    20,
	}
	item := svc.toCatalogItem(video, state)
	if item.Status != "in_progress" || item.ProgressPct <= 0 {
		t.Fatalf("expected in progress item with pct")
	}

	state.Completed = true
	item = svc.toCatalogItem(video, state)
	if item.Status != "completed" || item.ProgressPct != 100 {
		t.Fatalf("expected completed item")
	}
}

func TestVideoServicePlaybackFlows(t *testing.T) {
	var playbackState VideoPlaybackStateModel
	repo := &videoRepoMock{
		getVideoFileByIDFn: func(id int) (VideoFileModel, error) {
			return VideoFileModel{ID: id, Name: "v", ParentPath: "/series/show", Path: "/series/show/v.mp4", Format: ".mp4"}, nil
		},
		getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
			return VideoPlaylistModel{ID: id, Name: "p"}, nil
		},
		checkVideoInPlaylistFn: func(playlistID int, videoID int) (bool, error) { return true, nil },
		getPlaylistByContextFn: func(contextType string, sourcePath string) (VideoPlaylistModel, error) {
			return VideoPlaylistModel{}, errors.New("missing")
		},
		getVideosByParentPathFn: func(parentPath string) ([]VideoFileModel, error) {
			return []VideoFileModel{
				{ID: 1, Name: "v1", ParentPath: parentPath, Path: parentPath + "/v1.mp4"},
				{ID: 2, Name: "v2", ParentPath: parentPath, Path: parentPath + "/v2.mp4"},
			}, nil
		},
		createPlaylistFn: func(tx *sql.Tx, contextType string, sourcePath string) (VideoPlaylistModel, error) {
			return VideoPlaylistModel{ID: 20, Name: "ctx"}, nil
		},
		replacePlaylistItemsFn: func(tx *sql.Tx, playlistID int, videoIDs []int) error { return nil },
		getPlaybackStateFn: func(clientID string) (VideoPlaybackStateModel, error) {
			if playbackState.ClientID == "" {
				return VideoPlaybackStateModel{}, errors.New("none")
			}
			return playbackState, nil
		},
		upsertPlaybackStateFn: func(tx *sql.Tx, state VideoPlaybackStateModel) (VideoPlaybackStateModel, error) {
			state.ID = 1
			playbackState = state
			return state, nil
		},
		touchPlaylistFn: func(tx *sql.Tx, playlistID int) error { return nil },
		getVideoPlaylistItemsFn: func(playlistID int) ([]VideoPlaylistItemModel, error) {
			return []VideoPlaylistItemModel{
				{ID: 1, PlaylistID: playlistID, VideoID: 1, OrderIndex: 0, Video: VideoFileModel{ID: 1, Name: "v1", ParentPath: "/series/show", Path: "/series/show/v1.mp4"}},
				{ID: 2, PlaylistID: playlistID, VideoID: 2, OrderIndex: 1, Video: VideoFileModel{ID: 2, Name: "v2", ParentPath: "/series/show", Path: "/series/show/v2.mp4"}},
			}, nil
		},
	}
	svc := newVideoServiceForTest(t, repo)

	pid := 20
	session, err := svc.StartPlayback("c1", 1, &pid)
	if err != nil {
		t.Fatalf("StartPlayback with explicit playlist failed: %v", err)
	}
	if session.PlaybackState.VideoID == nil || *session.PlaybackState.VideoID != 1 {
		t.Fatalf("expected started video ID 1")
	}

	session, err = svc.StartPlayback("c2", 2, nil)
	if err != nil {
		t.Fatalf("StartPlayback with context playlist failed: %v", err)
	}
	if session.Playlist.ID == 0 {
		t.Fatalf("expected created context playlist")
	}

	updated, err := svc.UpdatePlaybackState("c2", UpdatePlaybackStateRequest{
		CurrentTime: ptrFloat(15),
		Duration:    ptrFloat(30),
		IsPaused:    ptrBool(false),
		Completed:   ptrBool(false),
	})
	if err != nil {
		t.Fatalf("UpdatePlaybackState failed: %v", err)
	}
	if updated.CurrentTime != 15 {
		t.Fatalf("expected updated current time")
	}

	next, err := svc.NextVideo("c2")
	if err != nil {
		t.Fatalf("NextVideo failed: %v", err)
	}
	if next.PlaybackState.VideoID == nil || *next.PlaybackState.VideoID != 2 {
		t.Fatalf("expected next video to be 2")
	}

	prev, err := svc.PreviousVideo("c2")
	if err != nil {
		t.Fatalf("PreviousVideo failed: %v", err)
	}
	if prev.PlaybackState.VideoID == nil || *prev.PlaybackState.VideoID != 1 {
		t.Fatalf("expected previous video to be 1")
	}

	got, err := svc.GetPlaybackState("c2")
	if err != nil {
		t.Fatalf("GetPlaybackState failed: %v", err)
	}
	if got.Playlist.ID == 0 {
		t.Fatalf("expected active playlist")
	}
}

func TestVideoServiceHomeCatalog(t *testing.T) {
	repo := &videoRepoMock{
		getCatalogVideosFn: func(limit int) ([]VideoFileModel, error) {
			return []VideoFileModel{
				{ID: 1, Name: "Show S01E01", ParentPath: "/series/show", Path: "/series/show/Show S01E01.mkv"},
				{ID: 2, Name: "Movie", ParentPath: "/movies", Path: "/movies/Movie.mkv"},
				{ID: 3, Name: "Personal clip", ParentPath: "/personal", Path: "/personal/clip.mp4"},
			}, nil
		},
		getRecentVideosFn: func(limit int) ([]VideoFileModel, error) {
			return []VideoFileModel{
				{ID: 2, Name: "Movie", ParentPath: "/movies", Path: "/movies/Movie.mkv"},
			}, nil
		},
		getPlaybackStateFn: func(clientID string) (VideoPlaybackStateModel, error) {
			return VideoPlaybackStateModel{
				ClientID:    clientID,
				VideoID:     sql.NullInt64{Int64: 1, Valid: true},
				CurrentTime: 5,
				Duration:    10,
			}, nil
		},
	}
	svc := newVideoServiceForTest(t, repo)
	catalog, err := svc.GetHomeCatalog("c", 2)
	if err != nil {
		t.Fatalf("GetHomeCatalog failed: %v", err)
	}
	if len(catalog.Sections) == 0 {
		t.Fatalf("expected non-empty catalog sections")
	}
}

func TestVideoService_ErrorBranchesAndContextPlaylistVariants(t *testing.T) {
	t.Run("ensureContextPlaylist returns existing playlist with items", func(t *testing.T) {
		repo := &videoRepoMock{
			getPlaylistByContextFn: func(contextType string, sourcePath string) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{ID: 70, Name: "ctx"}, nil
			},
			getVideoPlaylistItemsFn: func(playlistID int) ([]VideoPlaylistItemModel, error) {
				return []VideoPlaylistItemModel{{ID: 1, PlaylistID: playlistID, VideoID: 1}}, nil
			},
		}
		svc := newVideoServiceForTest(t, repo)
		playlist, err := svc.ensureContextPlaylist(string(ContextFolder), "/videos")
		if err != nil {
			t.Fatalf("expected existing context playlist, got %v", err)
		}
		if playlist.ID != 70 {
			t.Fatalf("expected playlist id 70, got %d", playlist.ID)
		}
	})

	t.Run("ensureContextPlaylist updates existing playlist when empty", func(t *testing.T) {
		replaced := false
		repo := &videoRepoMock{
			getPlaylistByContextFn: func(contextType string, sourcePath string) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{ID: 80, Name: "ctx"}, nil
			},
			getVideoPlaylistItemsFn: func(playlistID int) ([]VideoPlaylistItemModel, error) {
				return []VideoPlaylistItemModel{}, nil
			},
			getVideosByParentPathFn: func(parentPath string) ([]VideoFileModel, error) {
				return []VideoFileModel{{ID: 1, Name: "v1", ParentPath: parentPath, Path: parentPath + "/v1.mp4"}}, nil
			},
			replacePlaylistItemsFn: func(tx *sql.Tx, playlistID int, videoIDs []int) error {
				replaced = true
				return nil
			},
		}
		svc := newVideoServiceForTest(t, repo)
		playlist, err := svc.ensureContextPlaylist(string(ContextFolder), "/videos")
		if err != nil {
			t.Fatalf("expected update existing context playlist, got %v", err)
		}
		if playlist.ID != 80 || !replaced {
			t.Fatalf("expected existing playlist replacement path")
		}
	})

	t.Run("ensureContextPlaylist reports no videos found", func(t *testing.T) {
		repo := &videoRepoMock{
			getPlaylistByContextFn: func(contextType string, sourcePath string) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{}, errors.New("missing")
			},
			getVideosByParentPathFn: func(parentPath string) ([]VideoFileModel, error) {
				return []VideoFileModel{}, nil
			},
		}
		svc := newVideoServiceForTest(t, repo)
		if _, err := svc.ensureContextPlaylist(string(ContextFolder), "/empty"); err == nil {
			t.Fatalf("expected no videos found error")
		}
	})

	t.Run("getPlaybackState fails without active playlist", func(t *testing.T) {
		repo := &videoRepoMock{
			getPlaybackStateFn: func(clientID string) (VideoPlaybackStateModel, error) {
				return VideoPlaybackStateModel{ClientID: clientID}, nil
			},
		}
		svc := newVideoServiceForTest(t, repo)
		if _, err := svc.GetPlaybackState("c1"); err == nil {
			t.Fatalf("expected missing playlist error")
		}
	})

	t.Run("removeVideoFromPlaylist upserts exclusion for auto playlist", func(t *testing.T) {
		upsertExclusionCalled := false
		repo := &videoRepoMock{
			getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{ID: id, Name: "auto", IsAuto: true}, nil
			},
			removePlaylistVideoFn: func(tx *sql.Tx, playlistID int, videoID int) error {
				return nil
			},
			upsertPlaylistExclusionFn: func(tx *sql.Tx, playlistID int, videoID int) error {
				upsertExclusionCalled = true
				return nil
			},
		}
		svc := newVideoServiceForTest(t, repo)
		if err := svc.RemoveVideoFromPlaylist(10, 3); err != nil {
			t.Fatalf("expected remove from auto playlist success, got %v", err)
		}
		if !upsertExclusionCalled {
			t.Fatalf("expected exclusion upsert for auto playlist")
		}
	})
}

func TestVideoService_MoreErrorBranches(t *testing.T) {
	t.Run("AddVideoToPlaylist propagates add error", func(t *testing.T) {
		repo := &videoRepoMock{
			addPlaylistVideoManualFn: func(tx *sql.Tx, playlistID int, videoID int) error {
				return errors.New("add failed")
			},
		}
		svc := newVideoServiceForTest(t, repo)
		if err := svc.AddVideoToPlaylist(1, 2); err == nil {
			t.Fatalf("expected add error")
		}
	})

	t.Run("AddVideoToPlaylist propagates exclusion delete error", func(t *testing.T) {
		repo := &videoRepoMock{
			addPlaylistVideoManualFn: func(tx *sql.Tx, playlistID int, videoID int) error { return nil },
			deletePlaylistExclusionFn: func(tx *sql.Tx, playlistID int, videoID int) error {
				return errors.New("delete exclusion failed")
			},
		}
		svc := newVideoServiceForTest(t, repo)
		if err := svc.AddVideoToPlaylist(1, 2); err == nil {
			t.Fatalf("expected exclusion delete error")
		}
	})

	t.Run("RemoveVideoFromPlaylist propagates playlist fetch error", func(t *testing.T) {
		repo := &videoRepoMock{
			getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{}, errors.New("playlist not found")
			},
		}
		svc := newVideoServiceForTest(t, repo)
		if err := svc.RemoveVideoFromPlaylist(1, 2); err == nil {
			t.Fatalf("expected playlist fetch error")
		}
	})

	t.Run("RemoveVideoFromPlaylist propagates remove error", func(t *testing.T) {
		repo := &videoRepoMock{
			getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{ID: id, IsAuto: false}, nil
			},
			removePlaylistVideoFn: func(tx *sql.Tx, playlistID int, videoID int) error {
				return errors.New("remove failed")
			},
		}
		svc := newVideoServiceForTest(t, repo)
		if err := svc.RemoveVideoFromPlaylist(1, 2); err == nil {
			t.Fatalf("expected remove error")
		}
	})

	t.Run("StartPlayback rejects video not in selected playlist", func(t *testing.T) {
		repo := &videoRepoMock{
			getVideoFileByIDFn: func(id int) (VideoFileModel, error) {
				return VideoFileModel{ID: id, ParentPath: "/videos", Path: "/videos/v.mp4"}, nil
			},
			getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{ID: id}, nil
			},
			checkVideoInPlaylistFn: func(playlistID int, videoID int) (bool, error) {
				return false, nil
			},
		}
		svc := newVideoServiceForTest(t, repo)
		pid := 10
		if _, err := svc.StartPlayback("c1", 1, &pid); err == nil {
			t.Fatalf("expected not-in-playlist error")
		}
	})

	t.Run("StartPlayback propagates selected-playlist lookup error", func(t *testing.T) {
		repo := &videoRepoMock{
			getVideoFileByIDFn: func(id int) (VideoFileModel, error) {
				return VideoFileModel{ID: id, ParentPath: "/videos", Path: "/videos/v.mp4"}, nil
			},
			getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{}, errors.New("playlist lookup failed")
			},
		}
		svc := newVideoServiceForTest(t, repo)
		pid := 10
		if _, err := svc.StartPlayback("c1", 1, &pid); err == nil {
			t.Fatalf("expected playlist lookup error")
		}
	})

	t.Run("StartPlayback propagates in-playlist check error", func(t *testing.T) {
		repo := &videoRepoMock{
			getVideoFileByIDFn: func(id int) (VideoFileModel, error) {
				return VideoFileModel{ID: id, ParentPath: "/videos", Path: "/videos/v.mp4"}, nil
			},
			getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{ID: id}, nil
			},
			checkVideoInPlaylistFn: func(playlistID int, videoID int) (bool, error) {
				return false, errors.New("membership check failed")
			},
		}
		svc := newVideoServiceForTest(t, repo)
		pid := 10
		if _, err := svc.StartPlayback("c1", 1, &pid); err == nil {
			t.Fatalf("expected membership check error")
		}
	})

	t.Run("StartPlayback propagates context videos error", func(t *testing.T) {
		repo := &videoRepoMock{
			getVideoFileByIDFn: func(id int) (VideoFileModel, error) {
				return VideoFileModel{ID: id, ParentPath: "/videos", Path: "/videos/v.mp4"}, nil
			},
			getPlaylistByContextFn: func(contextType string, sourcePath string) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{}, errors.New("missing")
			},
			getVideosByParentPathFn: func(parentPath string) ([]VideoFileModel, error) {
				return nil, errors.New("list videos failed")
			},
		}
		svc := newVideoServiceForTest(t, repo)
		if _, err := svc.StartPlayback("c1", 1, nil); err == nil {
			t.Fatalf("expected context videos error")
		}
	})

	t.Run("StartPlayback propagates upsert and touch errors", func(t *testing.T) {
		repoUpsert := &videoRepoMock{
			getVideoFileByIDFn: func(id int) (VideoFileModel, error) {
				return VideoFileModel{ID: id, ParentPath: "/videos", Path: "/videos/v.mp4"}, nil
			},
			getPlaylistByContextFn: func(contextType string, sourcePath string) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{ID: 1}, nil
			},
			getVideoPlaylistItemsFn: func(playlistID int) ([]VideoPlaylistItemModel, error) {
				return []VideoPlaylistItemModel{{ID: 1, PlaylistID: playlistID, VideoID: 1}}, nil
			},
			getPlaybackStateFn: func(clientID string) (VideoPlaybackStateModel, error) {
				return VideoPlaybackStateModel{}, errors.New("none")
			},
			upsertPlaybackStateFn: func(tx *sql.Tx, state VideoPlaybackStateModel) (VideoPlaybackStateModel, error) {
				return VideoPlaybackStateModel{}, errors.New("upsert failed")
			},
		}
		svcUpsert := newVideoServiceForTest(t, repoUpsert)
		if _, err := svcUpsert.StartPlayback("c1", 1, nil); err == nil {
			t.Fatalf("expected upsert error")
		}

		repoTouch := &videoRepoMock{
			getVideoFileByIDFn: func(id int) (VideoFileModel, error) {
				return VideoFileModel{ID: id, ParentPath: "/videos", Path: "/videos/v.mp4"}, nil
			},
			getPlaylistByContextFn: func(contextType string, sourcePath string) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{ID: 1}, nil
			},
			getVideoPlaylistItemsFn: func(playlistID int) ([]VideoPlaylistItemModel, error) {
				return []VideoPlaylistItemModel{{ID: 1, PlaylistID: playlistID, VideoID: 1}}, nil
			},
			getPlaybackStateFn: func(clientID string) (VideoPlaybackStateModel, error) {
				return VideoPlaybackStateModel{}, errors.New("none")
			},
			upsertPlaybackStateFn: func(tx *sql.Tx, state VideoPlaybackStateModel) (VideoPlaybackStateModel, error) {
				state.ID = 1
				return state, nil
			},
			touchPlaylistFn: func(tx *sql.Tx, playlistID int) error { return errors.New("touch failed") },
		}
		svcTouch := newVideoServiceForTest(t, repoTouch)
		if _, err := svcTouch.StartPlayback("c1", 1, nil); err == nil {
			t.Fatalf("expected touch error")
		}
	})

	t.Run("RebuildSmartPlaylists propagates repository errors", func(t *testing.T) {
		cases := []struct {
			name string
			repo *videoRepoMock
		}{
			{
				name: "get grouped videos",
				repo: &videoRepoMock{
					getAllVideosWithMetadataFn: func() ([]VideoWithMetadataModel, error) {
						return nil, errors.New("grouping failed")
					},
				},
			},
			{
				name: "upsert auto playlist",
				repo: &videoRepoMock{
					getAllVideosWithMetadataFn: func() ([]VideoWithMetadataModel, error) {
						return []VideoWithMetadataModel{
							{VideoFileModel: VideoFileModel{ID: 1, Name: "Show S01E01.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E01.mkv"}},
							{VideoFileModel: VideoFileModel{ID: 2, Name: "Show S01E02.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E02.mkv"}},
						}, nil
					},
					upsertAutoPlaylistFn: func(tx *sql.Tx, contextType, sourcePath, name, groupMode, classification string) (VideoPlaylistModel, error) {
						return VideoPlaylistModel{}, errors.New("upsert failed")
					},
				},
			},
			{
				name: "playlist exclusions",
				repo: &videoRepoMock{
					getAllVideosWithMetadataFn: func() ([]VideoWithMetadataModel, error) {
						return []VideoWithMetadataModel{
							{VideoFileModel: VideoFileModel{ID: 1, Name: "Show S01E01.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E01.mkv"}},
							{VideoFileModel: VideoFileModel{ID: 2, Name: "Show S01E02.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E02.mkv"}},
						}, nil
					},
					upsertAutoPlaylistFn: func(tx *sql.Tx, contextType, sourcePath, name, groupMode, classification string) (VideoPlaylistModel, error) {
						return VideoPlaylistModel{ID: 1}, nil
					},
					getPlaylistExclusionsFn: func(playlistID int) (map[int]bool, error) {
						return nil, errors.New("exclusions failed")
					},
				},
			},
			{
				name: "delete auto items",
				repo: &videoRepoMock{
					getAllVideosWithMetadataFn: func() ([]VideoWithMetadataModel, error) {
						return []VideoWithMetadataModel{
							{VideoFileModel: VideoFileModel{ID: 1, Name: "Show S01E01.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E01.mkv"}},
							{VideoFileModel: VideoFileModel{ID: 2, Name: "Show S01E02.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E02.mkv"}},
						}, nil
					},
					upsertAutoPlaylistFn: func(tx *sql.Tx, contextType, sourcePath, name, groupMode, classification string) (VideoPlaylistModel, error) {
						return VideoPlaylistModel{ID: 1}, nil
					},
					getPlaylistExclusionsFn:   func(playlistID int) (map[int]bool, error) { return map[int]bool{}, nil },
					deleteAutoPlaylistItemsFn: func(tx *sql.Tx, playlistID int) error { return errors.New("delete auto failed") },
				},
			},
			{
				name: "insert playlist items",
				repo: &videoRepoMock{
					getAllVideosWithMetadataFn: func() ([]VideoWithMetadataModel, error) {
						return []VideoWithMetadataModel{
							{VideoFileModel: VideoFileModel{ID: 1, Name: "Show S01E01.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E01.mkv"}},
							{VideoFileModel: VideoFileModel{ID: 2, Name: "Show S01E02.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E02.mkv"}},
						}, nil
					},
					upsertAutoPlaylistFn: func(tx *sql.Tx, contextType, sourcePath, name, groupMode, classification string) (VideoPlaylistModel, error) {
						return VideoPlaylistModel{ID: 1}, nil
					},
					getPlaylistExclusionsFn:   func(playlistID int) (map[int]bool, error) { return map[int]bool{}, nil },
					deleteAutoPlaylistItemsFn: func(tx *sql.Tx, playlistID int) error { return nil },
					insertPlaylistItemsSrcFn: func(tx *sql.Tx, playlistID int, videoIDs []int, sourceKind string) error {
						return errors.New("insert auto failed")
					},
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc := newVideoServiceForTest(t, tc.repo)
				if err := svc.RebuildSmartPlaylists(); err == nil {
					t.Fatalf("expected error for case %s", tc.name)
				}
			})
		}
	})

	t.Run("shiftPlayback validates active state and playlist/items errors", func(t *testing.T) {
		repoNoActive := &videoRepoMock{
			getPlaybackStateFn: func(clientID string) (VideoPlaybackStateModel, error) {
				return VideoPlaybackStateModel{ClientID: clientID}, nil
			},
		}
		svcNoActive := newVideoServiceForTest(t, repoNoActive)
		if _, err := svcNoActive.NextVideo("c1"); err == nil {
			t.Fatalf("expected no-active-video error")
		}

		repoPlaylistErr := &videoRepoMock{
			getPlaybackStateFn: func(clientID string) (VideoPlaybackStateModel, error) {
				return VideoPlaybackStateModel{
					ClientID:   clientID,
					PlaylistID: sql.NullInt64{Int64: 10, Valid: true},
					VideoID:    sql.NullInt64{Int64: 1, Valid: true},
				}, nil
			},
			getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{}, errors.New("playlist failed")
			},
		}
		svcPlaylistErr := newVideoServiceForTest(t, repoPlaylistErr)
		if _, err := svcPlaylistErr.NextVideo("c1"); err == nil {
			t.Fatalf("expected playlist lookup error")
		}

		repoNoItems := &videoRepoMock{
			getPlaybackStateFn: func(clientID string) (VideoPlaybackStateModel, error) {
				return VideoPlaybackStateModel{
					ClientID:   clientID,
					PlaylistID: sql.NullInt64{Int64: 10, Valid: true},
					VideoID:    sql.NullInt64{Int64: 1, Valid: true},
				}, nil
			},
			getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
				return VideoPlaylistModel{ID: id}, nil
			},
			getVideoPlaylistItemsFn: func(playlistID int) ([]VideoPlaylistItemModel, error) {
				return []VideoPlaylistItemModel{}, nil
			},
		}
		svcNoItems := newVideoServiceForTest(t, repoNoItems)
		if _, err := svcNoItems.NextVideo("c1"); err == nil {
			t.Fatalf("expected empty playlist error")
		}
	})
}

func ptrFloat(v float64) *float64 { return &v }
func ptrBool(v bool) *bool        { return &v }
