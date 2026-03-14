package music

import (
	"database/sql"
	"errors"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type musicRepoMock struct {
	db *database.DbContext

	getPlaylistsFn         func(page int, pageSize int) (utils.PaginationResponse[PlaylistModel], error)
	getPlaylistByIDFn      func(id int) (PlaylistModel, error)
	createPlaylistFn       func(tx *sql.Tx, name string, description string, isSystem bool) (PlaylistModel, error)
	updatePlaylistFn       func(tx *sql.Tx, id int, name string, description string) (PlaylistModel, error)
	deletePlaylistFn       func(tx *sql.Tx, id int) error
	getPlaylistTracksFn    func(playlistID int, page int, pageSize int) (utils.PaginationResponse[PlaylistTrackModel], error)
	addPlaylistTrackFn     func(tx *sql.Tx, playlistID int, fileID int) (PlaylistTrackModel, error)
	removePlaylistTrackFn  func(tx *sql.Tx, playlistID int, fileID int) error
	reorderPlaylistTrackFn func(tx *sql.Tx, playlistID int, fileID int, position int) error
	getNowPlayingFn        func() (PlaylistModel, error)
	getPlayerStateFn       func(clientID string) (PlayerStateModel, error)
	upsertPlayerStateFn    func(tx *sql.Tx, state PlayerStateModel) (PlayerStateModel, error)
	getLibraryTracksFn     func(page int, pageSize int) (utils.PaginationResponse[files.FileModel], error)
	getLibraryIndexFn      func() ([]MusicLibraryIndexEntryModel, error)
	getLibraryFilesByIDsFn func(fileIDs []int) ([]files.FileModel, error)
}

func (m *musicRepoMock) GetDbContext() *database.DbContext { return m.db }
func (m *musicRepoMock) GetPlaylists(page int, pageSize int) (utils.PaginationResponse[PlaylistModel], error) {
	if m.getPlaylistsFn != nil {
		return m.getPlaylistsFn(page, pageSize)
	}
	return utils.PaginationResponse[PlaylistModel]{Items: []PlaylistModel{}}, nil
}
func (m *musicRepoMock) GetPlaylistByID(id int) (PlaylistModel, error) {
	if m.getPlaylistByIDFn != nil {
		return m.getPlaylistByIDFn(id)
	}
	return PlaylistModel{}, nil
}
func (m *musicRepoMock) CreatePlaylist(tx *sql.Tx, name string, description string, isSystem bool) (PlaylistModel, error) {
	if m.createPlaylistFn != nil {
		return m.createPlaylistFn(tx, name, description, isSystem)
	}
	return PlaylistModel{}, nil
}
func (m *musicRepoMock) UpdatePlaylist(tx *sql.Tx, id int, name string, description string) (PlaylistModel, error) {
	if m.updatePlaylistFn != nil {
		return m.updatePlaylistFn(tx, id, name, description)
	}
	return PlaylistModel{}, nil
}
func (m *musicRepoMock) DeletePlaylist(tx *sql.Tx, id int) error {
	if m.deletePlaylistFn != nil {
		return m.deletePlaylistFn(tx, id)
	}
	return nil
}
func (m *musicRepoMock) GetPlaylistTracks(playlistID int, page int, pageSize int) (utils.PaginationResponse[PlaylistTrackModel], error) {
	if m.getPlaylistTracksFn != nil {
		return m.getPlaylistTracksFn(playlistID, page, pageSize)
	}
	return utils.PaginationResponse[PlaylistTrackModel]{Items: []PlaylistTrackModel{}}, nil
}
func (m *musicRepoMock) AddPlaylistTrack(tx *sql.Tx, playlistID int, fileID int) (PlaylistTrackModel, error) {
	if m.addPlaylistTrackFn != nil {
		return m.addPlaylistTrackFn(tx, playlistID, fileID)
	}
	return PlaylistTrackModel{}, nil
}
func (m *musicRepoMock) RemovePlaylistTrack(tx *sql.Tx, playlistID int, fileID int) error {
	if m.removePlaylistTrackFn != nil {
		return m.removePlaylistTrackFn(tx, playlistID, fileID)
	}
	return nil
}
func (m *musicRepoMock) ReorderPlaylistTrack(tx *sql.Tx, playlistID int, fileID int, position int) error {
	if m.reorderPlaylistTrackFn != nil {
		return m.reorderPlaylistTrackFn(tx, playlistID, fileID, position)
	}
	return nil
}
func (m *musicRepoMock) GetNowPlaying() (PlaylistModel, error) {
	if m.getNowPlayingFn != nil {
		return m.getNowPlayingFn()
	}
	return PlaylistModel{}, nil
}
func (m *musicRepoMock) GetPlayerState(clientID string) (PlayerStateModel, error) {
	if m.getPlayerStateFn != nil {
		return m.getPlayerStateFn(clientID)
	}
	return PlayerStateModel{}, nil
}
func (m *musicRepoMock) UpsertPlayerState(tx *sql.Tx, state PlayerStateModel) (PlayerStateModel, error) {
	if m.upsertPlayerStateFn != nil {
		return m.upsertPlayerStateFn(tx, state)
	}
	return state, nil
}
func (m *musicRepoMock) GetLibraryTracks(page int, pageSize int) (utils.PaginationResponse[files.FileModel], error) {
	if m.getLibraryTracksFn != nil {
		return m.getLibraryTracksFn(page, pageSize)
	}
	return utils.PaginationResponse[files.FileModel]{Items: []files.FileModel{}}, nil
}
func (m *musicRepoMock) GetLibraryIndexEntries() ([]MusicLibraryIndexEntryModel, error) {
	if m.getLibraryIndexFn != nil {
		return m.getLibraryIndexFn()
	}
	return []MusicLibraryIndexEntryModel{}, nil
}
func (m *musicRepoMock) GetLibraryFilesByIDs(fileIDs []int) ([]files.FileModel, error) {
	if m.getLibraryFilesByIDsFn != nil {
		return m.getLibraryFilesByIDsFn(fileIDs)
	}
	return []files.FileModel{}, nil
}

func newMusicServiceForTest(t *testing.T, repo *musicRepoMock) *Service {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo.db = database.NewDbContext(db)
	return &Service{Repository: repo}
}

func TestMusicService_PlaylistsCRUD(t *testing.T) {
	now := time.Now()
	repo := &musicRepoMock{
		getPlaylistsFn: func(page int, pageSize int) (utils.PaginationResponse[PlaylistModel], error) {
			return utils.PaginationResponse[PlaylistModel]{
				Items: []PlaylistModel{{ID: 1, Name: "p1", CreatedAt: now, UpdatedAt: now}},
			}, nil
		},
		getPlaylistByIDFn: func(id int) (PlaylistModel, error) {
			return PlaylistModel{ID: id, Name: "by-id"}, nil
		},
		createPlaylistFn: func(tx *sql.Tx, name string, description string, isSystem bool) (PlaylistModel, error) {
			return PlaylistModel{ID: 10, Name: name, Description: description, IsSystem: isSystem}, nil
		},
		updatePlaylistFn: func(tx *sql.Tx, id int, name string, description string) (PlaylistModel, error) {
			return PlaylistModel{ID: id, Name: name, Description: description}, nil
		},
	}
	svc := newMusicServiceForTest(t, repo)

	playlists, err := svc.GetPlaylists(1, 10)
	if err != nil || len(playlists.Items) != 1 {
		t.Fatalf("expected playlists success, err=%v", err)
	}
	playlist, err := svc.GetPlaylistByID(5)
	if err != nil || playlist.ID != 5 {
		t.Fatalf("expected playlist by id success, err=%v", err)
	}
	created, err := svc.CreatePlaylist(CreatePlaylistRequest{Name: "new", Description: "d"})
	if err != nil || created.Name != "new" {
		t.Fatalf("expected create success, err=%v", err)
	}
	updated, err := svc.UpdatePlaylist(10, UpdatePlaylistRequest{Name: "up", Description: "desc"})
	if err != nil || updated.Name != "up" {
		t.Fatalf("expected update success, err=%v", err)
	}
}

func TestMusicService_TrackOperations(t *testing.T) {
	repo := &musicRepoMock{
		getPlaylistTracksFn: func(playlistID int, page int, pageSize int) (utils.PaginationResponse[PlaylistTrackModel], error) {
			return utils.PaginationResponse[PlaylistTrackModel]{
				Items: []PlaylistTrackModel{
					{
						ID:         1,
						PlaylistID: playlistID,
						FileID:     99,
						Position:   1,
						AddedAt:    time.Now(),
						FileName:   "song",
						FilePath:   "/tmp/song.mp3",
						FileType:   2,
					},
				},
			}, nil
		},
		addPlaylistTrackFn: func(tx *sql.Tx, playlistID int, fileID int) (PlaylistTrackModel, error) {
			return PlaylistTrackModel{
				ID:         2,
				PlaylistID: playlistID,
				FileID:     fileID,
				Position:   2,
				AddedAt:    time.Now(),
				FileName:   "new-song",
				FilePath:   "/tmp/new-song.mp3",
				FileType:   2,
			}, nil
		},
		removePlaylistTrackFn: func(tx *sql.Tx, playlistID int, fileID int) error {
			return nil
		},
		reorderPlaylistTrackFn: func(tx *sql.Tx, playlistID int, fileID int, position int) error {
			if position < 0 {
				return errors.New("invalid position")
			}
			return nil
		},
	}
	svc := newMusicServiceForTest(t, repo)

	tracks, err := svc.GetPlaylistTracks("client-1", 1, 1, 10)
	if err != nil || len(tracks.Items) != 1 {
		t.Fatalf("expected tracks success, err=%v", err)
	}
	added, err := svc.AddPlaylistTrack(1, 100)
	if err != nil || added.File.ID != 100 {
		t.Fatalf("expected add track success, err=%v", err)
	}
	if err := svc.RemovePlaylistTrack(1, 100); err != nil {
		t.Fatalf("expected remove track success, err=%v", err)
	}
	if err := svc.ReorderPlaylistTracks(1, []ReorderTrackItem{{FileID: 100, Position: 0}}); err != nil {
		t.Fatalf("expected reorder success, err=%v", err)
	}
}

func TestMusicService_NowPlayingAndPlayerState(t *testing.T) {
	repo := &musicRepoMock{
		getNowPlayingFn: func() (PlaylistModel, error) {
			return PlaylistModel{}, errors.New("not found")
		},
		createPlaylistFn: func(tx *sql.Tx, name string, description string, isSystem bool) (PlaylistModel, error) {
			return PlaylistModel{ID: 7, Name: name, IsSystem: isSystem}, nil
		},
		getPlayerStateFn: func(clientID string) (PlayerStateModel, error) {
			return PlayerStateModel{ID: 1, ClientID: clientID, Volume: 0.5}, nil
		},
		upsertPlayerStateFn: func(tx *sql.Tx, state PlayerStateModel) (PlayerStateModel, error) {
			state.ID = 2
			return state, nil
		},
	}
	svc := newMusicServiceForTest(t, repo)

	nowPlaying, err := svc.GetOrCreateNowPlaying()
	if err != nil || nowPlaying.ID != 7 || !nowPlaying.IsSystem {
		t.Fatalf("expected now playing creation success, err=%v", err)
	}

	state, err := svc.GetPlayerState("client-1")
	if err != nil || state.ClientID != "client-1" {
		t.Fatalf("expected get player state success, err=%v", err)
	}

	pid := 5
	fid := 9
	updated, err := svc.UpdatePlayerState("client-1", UpdatePlayerStateRequest{
		PlaylistID:      &pid,
		CurrentFileID:   &fid,
		CurrentPosition: 12.5,
		Volume:          0.9,
		Shuffle:         true,
		RepeatMode:      "all",
	})
	if err != nil || updated.ClientID != "client-1" {
		t.Fatalf("expected update player state success, err=%v", err)
	}
	if updated.PlaylistID == nil || *updated.PlaylistID != 5 {
		t.Fatalf("expected playlist id propagated")
	}
	if updated.CurrentFileID == nil || *updated.CurrentFileID != 9 {
		t.Fatalf("expected current file id propagated")
	}
}

func TestMusicService_ErrorPaths(t *testing.T) {
	repo := &musicRepoMock{
		getPlaylistsFn: func(page int, pageSize int) (utils.PaginationResponse[PlaylistModel], error) {
			return utils.PaginationResponse[PlaylistModel]{}, errors.New("repo error")
		},
		deletePlaylistFn: func(tx *sql.Tx, id int) error { return errors.New("delete error") },
		addPlaylistTrackFn: func(tx *sql.Tx, playlistID int, fileID int) (PlaylistTrackModel, error) {
			return PlaylistTrackModel{}, errors.New("add error")
		},
		reorderPlaylistTrackFn: func(tx *sql.Tx, playlistID int, fileID int, position int) error {
			return errors.New("reorder error")
		},
		getNowPlayingFn: func() (PlaylistModel, error) {
			return PlaylistModel{}, errors.New("not found")
		},
		createPlaylistFn: func(tx *sql.Tx, name string, description string, isSystem bool) (PlaylistModel, error) {
			return PlaylistModel{}, errors.New("create error")
		},
		upsertPlayerStateFn: func(tx *sql.Tx, state PlayerStateModel) (PlayerStateModel, error) {
			return PlayerStateModel{}, errors.New("upsert error")
		},
	}
	svc := newMusicServiceForTest(t, repo)

	if _, err := svc.GetPlaylists(1, 10); err == nil {
		t.Fatalf("expected get playlists error")
	}
	if err := svc.DeletePlaylist(1); err == nil {
		t.Fatalf("expected delete playlist error")
	}
	if _, err := svc.AddPlaylistTrack(1, 1); err == nil {
		t.Fatalf("expected add track error")
	}
	if err := svc.ReorderPlaylistTracks(1, []ReorderTrackItem{{FileID: 1, Position: 0}}); err == nil {
		t.Fatalf("expected reorder error")
	}
	if _, err := svc.GetOrCreateNowPlaying(); err == nil {
		t.Fatalf("expected now playing create error")
	}
	if _, err := svc.UpdatePlayerState("c", UpdatePlayerStateRequest{}); err == nil {
		t.Fatalf("expected update player state error")
	}
}

func TestMusicService_AdditionalErrorPaths(t *testing.T) {
	repo := &musicRepoMock{
		getPlaylistByIDFn: func(id int) (PlaylistModel, error) {
			return PlaylistModel{}, errors.New("playlist fetch failed")
		},
		getPlaylistTracksFn: func(playlistID int, page int, pageSize int) (utils.PaginationResponse[PlaylistTrackModel], error) {
			return utils.PaginationResponse[PlaylistTrackModel]{}, errors.New("tracks fetch failed")
		},
		getPlayerStateFn: func(clientID string) (PlayerStateModel, error) {
			return PlayerStateModel{}, errors.New("player state failed")
		},
	}
	svc := newMusicServiceForTest(t, repo)

	if _, err := svc.GetPlaylistByID(1); err == nil {
		t.Fatalf("expected GetPlaylistByID error")
	}
	if _, err := svc.GetPlaylistTracks("client-1", 1, 1, 10); err == nil {
		t.Fatalf("expected GetPlaylistTracks error")
	}
	if _, err := svc.GetPlayerState("c1"); err == nil {
		t.Fatalf("expected GetPlayerState error")
	}
}

func TestMusicNewService(t *testing.T) {
	repo := &musicRepoMock{}
	svc := NewService(repo)
	typed, ok := svc.(*Service)
	if !ok || typed.Repository != repo {
		t.Fatalf("expected concrete service with repository")
	}
}
