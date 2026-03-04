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
}

type ServiceInterface interface {
	StartPlayback(clientID string, videoID int) (PlaybackSessionDto, error)
	GetPlaybackState(clientID string) (PlaybackSessionDto, error)
	UpdatePlaybackState(clientID string, req UpdatePlaybackStateRequest) (VideoPlaybackStateDto, error)
	NextVideo(clientID string) (PlaybackSessionDto, error)
	PreviousVideo(clientID string) (PlaybackSessionDto, error)
	GetHomeCatalog(clientID string, limit int) (VideoHomeCatalogDto, error)
}
