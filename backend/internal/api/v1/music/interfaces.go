package music

import (
	"database/sql"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetPlaylists(page int, pageSize int) (utils.PaginationResponse[PlaylistModel], error)
	GetPlaylistByID(id int) (PlaylistModel, error)
	CreatePlaylist(tx *sql.Tx, name string, description string, isSystem bool) (PlaylistModel, error)
	UpdatePlaylist(tx *sql.Tx, id int, name string, description string) (PlaylistModel, error)
	DeletePlaylist(tx *sql.Tx, id int) error
	GetPlaylistTracks(playlistID int, page int, pageSize int) (utils.PaginationResponse[PlaylistTrackModel], error)
	AddPlaylistTrack(tx *sql.Tx, playlistID int, fileID int) (PlaylistTrackModel, error)
	RemovePlaylistTrack(tx *sql.Tx, playlistID int, fileID int) error
	ReorderPlaylistTrack(tx *sql.Tx, playlistID int, fileID int, position int) error
	GetNowPlaying() (PlaylistModel, error)
	GetPlayerState(clientID string) (PlayerStateModel, error)
	UpsertPlayerState(tx *sql.Tx, state PlayerStateModel) (PlayerStateModel, error)
	GetLibraryTracks(page int, pageSize int) (utils.PaginationResponse[files.FileModel], error)
	GetLibraryIndexEntries() ([]MusicLibraryIndexEntryModel, error)
	GetLibraryFilesByIDs(fileIDs []int) ([]files.FileModel, error)
}

type ServiceInterface interface {
	GetPlaylists(page int, pageSize int) (utils.PaginationResponse[PlaylistDto], error)
	GetAutomaticPlaylists(clientID string) ([]PlaylistDto, error)
	GetPlaylistByID(id int) (PlaylistDto, error)
	CreatePlaylist(req CreatePlaylistRequest) (PlaylistDto, error)
	UpdatePlaylist(id int, req UpdatePlaylistRequest) (PlaylistDto, error)
	DeletePlaylist(id int) error
	GetPlaylistTracks(clientID string, playlistID int, page int, pageSize int) (utils.PaginationResponse[PlaylistTrackDto], error)
	AddPlaylistTrack(playlistID int, fileID int) (PlaylistTrackDto, error)
	RemovePlaylistTrack(playlistID int, fileID int) error
	ReorderPlaylistTracks(playlistID int, tracks []ReorderTrackItem) error
	GetOrCreateNowPlaying() (PlaylistDto, error)
	GetHomeCatalog(clientID string, limit int) (MusicHomeCatalogDto, error)
	GetLibraryTracks(page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetLibraryArtists(page int, pageSize int) (utils.PaginationResponse[MusicArtistGroupDto], error)
	GetLibraryTracksByArtist(artistKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetLibraryAlbums(page int, pageSize int) (utils.PaginationResponse[MusicAlbumGroupDto], error)
	GetLibraryTracksByAlbum(albumKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetLibraryGenres(page int, pageSize int) (utils.PaginationResponse[MusicGenreGroupDto], error)
	GetLibraryTracksByGenre(genreKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetLibraryFolders(page int, pageSize int) (utils.PaginationResponse[MusicFolderGroupDto], error)
	GetLibraryTracksByFolder(folderPath string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetPlayerState(clientID string) (PlayerStateDto, error)
	UpdatePlayerState(clientID string, req UpdatePlayerStateRequest) (PlayerStateDto, error)
}
