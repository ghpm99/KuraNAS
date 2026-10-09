package music

import (
	"context"
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
	ReplacePlayerQueue(tx *sql.Tx, clientID string, fileIDs []int, currentIndex int) error
	GetPlayerQueue(clientID string) ([]MusicQueueEntryModel, error)
	GetPlayerQueueCurrentIndex(clientID string) (int, error)
	UpsertPlayerState(tx *sql.Tx, state PlayerStateModel) (PlayerStateModel, error)
	GetLibraryTracks(page int, pageSize int) (utils.PaginationResponse[files.FileModel], error)
	SearchLibraryTracks(searchText string, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error)
	GetLibrarySummary() (MusicLibrarySummaryDto, error)
	GetLibraryArtistGroups(page int, pageSize int, sort CatalogSort) (utils.PaginationResponse[MusicArtistGroupDto], error)
	GetLibraryAlbumGroups(page int, pageSize int, sort CatalogSort) (utils.PaginationResponse[MusicAlbumGroupDto], error)
	GetLibraryGenreGroups(page int, pageSize int, sort CatalogSort) (utils.PaginationResponse[MusicGenreGroupDto], error)
	GetLibraryFolderGroups(page int, pageSize int, sort CatalogSort) (utils.PaginationResponse[MusicFolderGroupDto], error)
	GetLibraryTrackIDsByArtist(artistKey string, page int, pageSize int) (utils.PaginationResponse[int], error)
	GetLibraryTrackIDsByAlbum(albumKey string, page int, pageSize int) (utils.PaginationResponse[int], error)
	GetLibraryTrackIDsByGenre(genreKey string, page int, pageSize int) (utils.PaginationResponse[int], error)
	GetLibraryTrackIDsByFolder(folderPath string, page int, pageSize int) (utils.PaginationResponse[int], error)
	GetLibraryQueueByArtist(artistKey string, limit int) ([]MusicQueueEntryModel, error)
	GetLibraryQueueByAlbum(albumKey string, limit int) ([]MusicQueueEntryModel, error)
	GetLibraryQueueByGenre(genreKey string, limit int) ([]MusicQueueEntryModel, error)
	GetLibraryQueueByFolder(folderPath string, limit int) ([]MusicQueueEntryModel, error)
	GetPlaylistQueue(playlistID int, limit int) ([]MusicQueueEntryModel, error)
	GetLibraryQueueByFileIDs(fileIDs []int) ([]MusicQueueEntryModel, error)
	GetRecentLibraryFileIDs(limit int) ([]int, error)
	GetFavoriteLibraryFileIDs(limit int) ([]int, error)
	GetArtistClusterInputs() ([]artistClusterInput, error)
	GetLibraryFileIDsByArtistKeys(artistKeys []string) ([]int, error)
	GetLibraryFilesByIDs(fileIDs []int) ([]files.FileModel, error)
	GetArtistClusters() ([]ArtistClusterModel, error)
	UpsertArtistCluster(tx *sql.Tx, cluster ArtistClusterModel) error
	DeleteArtistClustersExcept(tx *sql.Tx, artistKeys []string) error
	GetAIPlaylists() ([]PlaylistModel, error)
	CreateAIPlaylist(tx *sql.Tx, name string, description string) (PlaylistModel, error)
	ReplacePlaylistTracks(tx *sql.Tx, playlistID int, fileIDs []int) error
	// Browse queries (moved from files)
	GetMusic(page int, pageSize int) (utils.PaginationResponse[files.FileModel], error)
	GetMusicArtists(page int, pageSize int) (utils.PaginationResponse[MusicArtistDto], error)
	GetMusicByArtist(artist string, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error)
	GetMusicAlbums(page int, pageSize int) (utils.PaginationResponse[MusicAlbumDto], error)
	GetMusicByAlbum(album string, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error)
	GetMusicGenres(page int, pageSize int) (utils.PaginationResponse[MusicGenreDto], error)
	GetMusicByGenre(genre string, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error)
	GetMusicFolders(page int, pageSize int) (utils.PaginationResponse[MusicFolderDto], error)
}

// AudioMetadataRepositoryInterface is the write-side for audio complement metadata.
type AudioMetadataRepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetAudioMetadataByID(id int) (AudioMetadataModel, error)
	UpsertAudioMetadata(tx *sql.Tx, metadata AudioMetadataModel) (AudioMetadataModel, error)
	DeleteAudioMetadata(id int) error
	ListAudioWithoutMetadata(afterFileID int, limit int) ([]AudioWithoutMetadata, error)
	ListAudioWithStaleTags(afterFileID int, limit int) ([]AudioWithStaleTags, error)
	ListAudioWithoutCatalogKeys(afterAudioMetadataID int, limit int) ([]AudioCatalogKeySource, error)
	UpdateAudioCatalogKeys(tx *sql.Tx, audioMetadataID int, groupingKeys CatalogGroupingKeys) error
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
	GetHomeCatalog(clientID string, limit int, sort CatalogSort) (MusicHomeCatalogDto, error)
	GetLibraryTracks(page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	SearchLibraryTracks(searchText string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetLibraryArtists(page int, pageSize int, sort CatalogSort) (utils.PaginationResponse[MusicArtistGroupDto], error)
	GetLibraryTracksByArtist(artistKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetLibraryAlbums(page int, pageSize int, sort CatalogSort) (utils.PaginationResponse[MusicAlbumGroupDto], error)
	GetLibraryTracksByAlbum(albumKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetLibraryGenres(page int, pageSize int, sort CatalogSort) (utils.PaginationResponse[MusicGenreGroupDto], error)
	GetLibraryTracksByGenre(genreKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetLibraryFolders(page int, pageSize int, sort CatalogSort) (utils.PaginationResponse[MusicFolderGroupDto], error)
	GetLibraryTracksByFolder(folderPath string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetLibraryQueueByArtist(artistKey string) (MusicQueueDto, error)
	GetLibraryQueueByAlbum(albumKey string) (MusicQueueDto, error)
	GetLibraryQueueByGenre(genreKey string) (MusicQueueDto, error)
	GetLibraryQueueByFolder(folderPath string) (MusicQueueDto, error)
	GetPlaylistQueue(clientID string, playlistID int) (MusicQueueDto, error)
	GetPlayerState(clientID string) (PlayerStateDto, error)
	ReplacePlayerQueue(clientID string, request ReplacePlayerQueueRequest) error
	GetPlayerQueue(clientID string) (PlayerQueueDto, error)
	UpdatePlayerState(clientID string, req UpdatePlayerStateRequest) (PlayerStateDto, error)
	RebuildAIClusters(ctx context.Context) error
	// Browse methods (moved from files)
	GetMusic(page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetMusicArtists(page int, pageSize int) (utils.PaginationResponse[MusicArtistDto], error)
	GetMusicByArtist(artist string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetMusicAlbums(page int, pageSize int) (utils.PaginationResponse[MusicAlbumDto], error)
	GetMusicByAlbum(album string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetMusicGenres(page int, pageSize int) (utils.PaginationResponse[MusicGenreDto], error)
	GetMusicByGenre(genre string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetMusicFolders(page int, pageSize int) (utils.PaginationResponse[MusicFolderDto], error)
}
