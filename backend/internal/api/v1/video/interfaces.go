package video

import (
	"database/sql"
	"nas-go/api/pkg/database"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetVideoFileByID(id int) (VideoFileModel, error)
	GetVideosByParentPath(parentPath string) ([]VideoFileModel, error)
	GetPlaylistByContext(contextType string, sourcePath string) (VideoPlaylistModel, error)
	CreatePlaylist(tx *sql.Tx, contextType string, sourcePath string) (VideoPlaylistModel, error)
	ReplacePlaylistItems(tx *sql.Tx, playlistID int, videoIDs []int) error
	GetPlaylistItems(playlistID int) ([]VideoPlaylistItemModel, error)
	GetPlaybackState(clientID string) (VideoPlaybackStateModel, error)
	UpsertPlaybackState(tx *sql.Tx, state VideoPlaybackStateModel) (VideoPlaybackStateModel, error)
	TouchPlaylist(tx *sql.Tx, playlistID int) error
	GetCatalogVideos(limit int) ([]VideoFileModel, error)
	GetRecentVideos(limit int) ([]VideoFileModel, error)
	GetAllVideosForGrouping() ([]VideoFileModel, error)
	UpsertAutoPlaylist(tx *sql.Tx, contextType, sourcePath, name, groupMode, classification string) (VideoPlaylistModel, error)
	DeleteAutoPlaylistItems(tx *sql.Tx, playlistID int) error
	InsertPlaylistItemsWithSource(tx *sql.Tx, playlistID int, videoIDs []int, sourceKind string) error
	GetPlaylistExclusions(playlistID int) (map[int]bool, error)
	GetVideoPlaylists(includeHidden bool) ([]VideoPlaylistModel, error)
	GetVideoPlaylistByID(id int) (VideoPlaylistModel, error)
	GetVideoPlaylistItemsDetailed(playlistID int) ([]VideoPlaylistItemModel, error)
	SetPlaylistHidden(tx *sql.Tx, playlistID int, hidden bool) error
	AddPlaylistVideoManual(tx *sql.Tx, playlistID int, videoID int) error
	RemovePlaylistVideo(tx *sql.Tx, playlistID int, videoID int) error
	UpsertPlaylistExclusion(tx *sql.Tx, playlistID int, videoID int) error
	DeletePlaylistExclusion(tx *sql.Tx, playlistID int, videoID int) error
	GetUnassignedVideos(limit int) ([]VideoFileModel, error)
}

type ServiceInterface interface {
	StartPlayback(clientID string, videoID int) (PlaybackSessionDto, error)
	GetPlaybackState(clientID string) (PlaybackSessionDto, error)
	UpdatePlaybackState(clientID string, req UpdatePlaybackStateRequest) (VideoPlaybackStateDto, error)
	NextVideo(clientID string) (PlaybackSessionDto, error)
	PreviousVideo(clientID string) (PlaybackSessionDto, error)
	GetHomeCatalog(clientID string, limit int) (VideoHomeCatalogDto, error)
	RebuildSmartPlaylists() error
	GetPlaylists(includeHidden bool) ([]VideoPlaylistDto, error)
	GetPlaylistByID(id int) (VideoPlaylistDto, error)
	SetPlaylistHidden(playlistID int, hidden bool) error
	AddVideoToPlaylist(playlistID int, videoID int) error
	RemoveVideoFromPlaylist(playlistID int, videoID int) error
	GetUnassignedVideos(limit int) ([]VideoFileDto, error)
}
