package video

import (
	"database/sql"
	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
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
	GetVideoWatchProgress(clientID string, videoID int) (VideoWatchProgressModel, error)
	GetVideoWatchProgressByVideos(clientID string, videoIDs []int) ([]VideoWatchProgressModel, error)
	UpsertVideoWatchProgress(tx *sql.Tx, progress VideoWatchProgressModel) (VideoWatchProgressModel, error)
	GetContinueWatchingVideos(clientID string, limit int) ([]ContinueWatchingModel, error)
	TouchPlaylist(tx *sql.Tx, playlistID int) error
	GetCatalogVideos(limit int) ([]VideoFileModel, error)
	GetRecentVideos(limit int) ([]VideoFileModel, error)
	GetAllVideosForGrouping() ([]VideoFileModel, error)
	GetAllVideosWithMetadata() ([]VideoWithMetadataModel, error)
	UpsertAutoPlaylist(tx *sql.Tx, contextType, sourcePath, name, groupMode, classification string) (VideoPlaylistModel, error)
	DeleteAutoPlaylistItems(tx *sql.Tx, playlistID int) error
	DeleteStaleAutoPlaylists(tx *sql.Tx, keptPlaylistIDs []int) error
	RenumberPlaylistItems(tx *sql.Tx, playlistID int) error
	InsertPlaylistItemsWithSource(tx *sql.Tx, playlistID int, videoIDs []int, sourceKind string) error
	GetPlaylistExclusions(playlistID int) (map[int]bool, error)
	GetVideoPlaylists(includeHidden bool) ([]VideoPlaylistModel, error)
	GetVideoPlaylistsBySection(filter PlaylistSectionFilter, limit int, offset int) ([]VideoPlaylistModel, error)
	GetVideoPlaylistMemberships(includeHidden bool) ([]VideoPlaylistMembershipModel, error)
	GetVideoPlaylistsByVideo(videoID int) ([]VideoPlaylistOfVideoModel, error)
	GetVideoPlaylistByID(id int) (VideoPlaylistModel, error)
	GetVideoPlaylistItemsDetailed(playlistID int) ([]VideoPlaylistItemModel, error)
	GetVideoPlaylistItemsPage(playlistID int, limit int, offset int) ([]VideoPlaylistItemModel, error)
	ListLibraryVideos(page int, pageSize int, searchQuery string) (utils.PaginationResponse[VideoFileModel], error)
	ListLibraryFolders(query LibraryFolderQuery) ([]LibraryFolderModel, error)
	ListLibraryFolderVideos(folderPath string, limit int, offset int) ([]VideoFileModel, error)
	ListLibraryMovies(sort LibraryMovieSort, limit int, offset int) ([]VideoFileModel, error)
	SetPlaylistHidden(tx *sql.Tx, playlistID int, hidden bool) error
	AddPlaylistVideoManual(tx *sql.Tx, playlistID int, videoID int) error
	RemovePlaylistVideo(tx *sql.Tx, playlistID int, videoID int) error
	UpsertPlaylistExclusion(tx *sql.Tx, playlistID int, videoID int) error
	DeletePlaylistExclusion(tx *sql.Tx, playlistID int, videoID int) error
	GetUnassignedVideos(limit int) ([]VideoFileModel, error)
	CheckVideoInPlaylist(playlistID int, videoID int) (bool, error)
	UpdatePlaylistName(tx *sql.Tx, playlistID int, name string) error
	ReorderPlaylistItems(tx *sql.Tx, playlistID int, videoIDs []int, orderIndices []int) error
	InsertBehaviorEvent(tx *sql.Tx, event VideoBehaviorEventModel) (VideoBehaviorEventModel, error)
	GetBehaviorEvents(clientID string, limit int) ([]VideoBehaviorEventModel, error)
	GetAllBehaviorEvents(limit int) ([]VideoBehaviorEventModel, error)
	// Browse query (moved from files)
	GetVideos(page int, pageSize int) (utils.PaginationResponse[files.FileModel], error)
}

// VideoMetadataRepositoryInterface is the write-side for the video_metadata complement table.
type VideoMetadataRepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetVideoMetadataByID(id int) (VideoMetadataModel, error)
	UpsertVideoMetadata(tx *sql.Tx, metadata VideoMetadataModel) (VideoMetadataModel, error)
	DeleteVideoMetadata(id int) error
	ListVideosWithoutMetadata(afterFileID int, limit int) ([]VideoWithoutMetadata, error)
	ListVideosPendingClassification(afterMetadataID int, limit int) ([]VideoPendingClassification, error)
	UpdateVideoClassification(metadataID int, classification string) error
}

type ServiceInterface interface {
	StartPlayback(clientID string, videoID int, playlistID *int) (PlaybackSessionDto, error)
	GetPlaybackState(clientID string) (PlaybackSessionDto, error)
	UpdatePlaybackState(clientID string, req UpdatePlaybackStateRequest) (VideoPlaybackStateDto, error)
	NextVideo(clientID string) (PlaybackSessionDto, error)
	PreviousVideo(clientID string) (PlaybackSessionDto, error)
	GetHomeCatalog(clientID string, limit int) (VideoHomeCatalogDto, error)
	GetContinueWatching(clientID string, limit int) ([]ContinueWatchingItemDto, error)
	SetVideoWatched(clientID string, videoID int, isWatched bool) error
	RebuildSmartPlaylists() error
	GetPlaylists(includeHidden bool) ([]VideoPlaylistDto, error)
	GetPlaylistsBySection(request PlaylistSectionRequest) (utils.PaginationResponse[VideoPlaylistDto], error)
	GetPlaylistMemberships(includeHidden bool) ([]VideoPlaylistMembershipDto, error)
	GetPlaylistsByVideo(videoID int) ([]VideoPlaylistOfVideoDto, error)
	GetPlaylistByID(clientID string, id int) (VideoPlaylistDto, error)
	GetPlaylistItemsPage(clientID string, playlistID int, page int, pageSize int) (utils.PaginationResponse[VideoPlaylistItemDto], error)
	ListLibraryVideos(page int, pageSize int, searchQuery string) (utils.PaginationResponse[VideoFileDto], error)
	ListLibraryFolders(request LibraryFolderRequest) (utils.PaginationResponse[LibraryFolderDto], error)
	ListLibraryFolderVideos(request LibraryFolderVideosRequest) (utils.PaginationResponse[VideoFileDto], error)
	ListLibraryMovies(request LibraryMoviesRequest) (utils.PaginationResponse[VideoFileDto], error)
	SetPlaylistHidden(playlistID int, hidden bool) error
	AddVideoToPlaylist(playlistID int, videoID int) error
	RemoveVideoFromPlaylist(playlistID int, videoID int) error
	GetUnassignedVideos(limit int) ([]VideoFileDto, error)
	UpdatePlaylistName(playlistID int, name string) error
	ReorderPlaylistItems(playlistID int, items []ReorderPlaylistItemRequest) error
	TrackBehaviorEvent(clientID string, req TrackBehaviorEventRequest) error
	// Browse/streaming support (moved from files)
	GetVideos(page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetVideoThumbnail(fileDto files.FileDto, width, height int) ([]byte, error)
	GetVideoPreviewGif(fileDto files.FileDto, width, height int) ([]byte, error)
}
