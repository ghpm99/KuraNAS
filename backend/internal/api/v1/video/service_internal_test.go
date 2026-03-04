package video

import (
	"database/sql"
	"errors"
	"nas-go/api/pkg/database"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type videoRepoMock struct {
	db *database.DbContext

	getUnassignedVideosFn     func(limit int) ([]VideoFileModel, error)
	getVideoPlaylistsFn       func(includeHidden bool) ([]VideoPlaylistModel, error)
	getVideoPlaylistByIDFn    func(id int) (VideoPlaylistModel, error)
	getVideoPlaylistItemsFn   func(playlistID int) ([]VideoPlaylistItemModel, error)
	setPlaylistHiddenFn       func(tx *sql.Tx, playlistID int, hidden bool) error
	addPlaylistVideoManualFn  func(tx *sql.Tx, playlistID int, videoID int) error
	deletePlaylistExclusionFn func(tx *sql.Tx, playlistID int, videoID int) error
	removePlaylistVideoFn     func(tx *sql.Tx, playlistID int, videoID int) error
	upsertPlaylistExclusionFn func(tx *sql.Tx, playlistID int, videoID int) error
	updatePlaylistNameFn      func(tx *sql.Tx, playlistID int, name string) error
	reorderPlaylistItemFn     func(tx *sql.Tx, playlistID int, videoID int, orderIndex int) error
	getAllVideosForGroupingFn func() ([]VideoFileModel, error)
	upsertAutoPlaylistFn      func(tx *sql.Tx, contextType, sourcePath, name, groupMode, classification string) (VideoPlaylistModel, error)
	getPlaylistExclusionsFn   func(playlistID int) (map[int]bool, error)
	deleteAutoPlaylistItemsFn func(tx *sql.Tx, playlistID int) error
	insertPlaylistItemsSrcFn  func(tx *sql.Tx, playlistID int, videoIDs []int, sourceKind string) error
}

func (m *videoRepoMock) GetDbContext() *database.DbContext { return m.db }
func (m *videoRepoMock) GetVideoFileByID(id int) (VideoFileModel, error) {
	return VideoFileModel{}, errors.New("not used")
}
func (m *videoRepoMock) GetVideosByParentPath(parentPath string) ([]VideoFileModel, error) {
	return nil, errors.New("not used")
}
func (m *videoRepoMock) GetPlaylistByContext(contextType string, sourcePath string) (VideoPlaylistModel, error) {
	return VideoPlaylistModel{}, errors.New("not used")
}
func (m *videoRepoMock) CreatePlaylist(tx *sql.Tx, contextType string, sourcePath string) (VideoPlaylistModel, error) {
	return VideoPlaylistModel{}, errors.New("not used")
}
func (m *videoRepoMock) ReplacePlaylistItems(tx *sql.Tx, playlistID int, videoIDs []int) error {
	return errors.New("not used")
}
func (m *videoRepoMock) GetPlaylistItems(playlistID int) ([]VideoPlaylistItemModel, error) {
	if m.getVideoPlaylistItemsFn != nil {
		return m.getVideoPlaylistItemsFn(playlistID)
	}
	return nil, nil
}
func (m *videoRepoMock) GetPlaybackState(clientID string) (VideoPlaybackStateModel, error) {
	return VideoPlaybackStateModel{}, errors.New("not used")
}
func (m *videoRepoMock) UpsertPlaybackState(tx *sql.Tx, state VideoPlaybackStateModel) (VideoPlaybackStateModel, error) {
	return VideoPlaybackStateModel{}, errors.New("not used")
}
func (m *videoRepoMock) TouchPlaylist(tx *sql.Tx, playlistID int) error { return nil }
func (m *videoRepoMock) GetCatalogVideos(limit int) ([]VideoFileModel, error) {
	return nil, errors.New("not used")
}
func (m *videoRepoMock) GetRecentVideos(limit int) ([]VideoFileModel, error) {
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

func newVideoServiceForTest(t *testing.T, repo *videoRepoMock) *Service {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo.db = database.NewDbContext(db)
	return &Service{Repository: repo}
}

func TestVideoHelpersClassificationAndGrouping(t *testing.T) {
	if got := classifyVideo(VideoFileModel{Name: "S01E02 episode", ParentPath: "/series/show"}); got != "series" {
		t.Fatalf("expected series classification, got %s", got)
	}
	if got := classifyVideo(VideoFileModel{Name: "Movie", ParentPath: "/movies"}); got != "movie" {
		t.Fatalf("expected movie classification, got %s", got)
	}
	if got := classifyVideo(VideoFileModel{Name: "Clip", ParentPath: "/personal"}); got != "personal" {
		t.Fatalf("expected personal classification, got %s", got)
	}

	if got := inferTitlePrefix("My.Show.S01E02.mkv"); got == "" {
		t.Fatalf("expected inferred title prefix")
	}
	if !isGenericFolderName("videos") || isGenericFolderName("myfolder") {
		t.Fatalf("generic folder detection mismatch")
	}
	if got := classifySmartVideo(VideoFileModel{Name: "tutorial", ParentPath: "/x"}); got != "program" {
		t.Fatalf("expected program smart classification, got %s", got)
	}

	groups := buildSmartGroups([]VideoFileModel{
		{ID: 1, Name: "Show S01E01.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E01.mkv"},
		{ID: 2, Name: "Show S01E02.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E02.mkv"},
		{ID: 3, Name: "Movie.mkv", ParentPath: "/movies", Path: "/movies/Movie.mkv"},
	})
	if len(groups) == 0 {
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
	if _, err := svc.GetPlaylistByID(10); err != nil {
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

func TestVideoServiceRebuildSmartPlaylists(t *testing.T) {
	repo := &videoRepoMock{
		getAllVideosForGroupingFn: func() ([]VideoFileModel, error) {
			return []VideoFileModel{
				{ID: 1, Name: "Show S01E01.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E01.mkv"},
				{ID: 2, Name: "Show S01E02.mkv", ParentPath: "/series/show", Path: "/series/show/Show S01E02.mkv"},
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
