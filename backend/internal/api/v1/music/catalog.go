package music

import (
	"database/sql"
	"time"
)

const (
	AutoPlaylistContinueListeningID = -1
	AutoPlaylistRecentlyAddedID     = -2
	AutoPlaylistFavoritesID         = -3
)

const (
	PlaylistKindManual    = "manual"
	PlaylistKindSystem    = "system"
	PlaylistKindAutomatic = "automatic"
)

const (
	autoPlaylistContinueListeningKey = "continue-listening"
	autoPlaylistRecentlyAddedKey     = "recently-added"
	autoPlaylistFavoritesKey         = "favorites"
)

type MusicLibraryIndexEntryModel struct {
	FileID          int
	FileName        string
	FilePath        string
	ParentPath      string
	Starred         bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastInteraction sql.NullTime
	Title           string
	Artist          string
	AlbumArtist     string
	Album           string
	Genre           string
	Year            string
	TrackNumber     string
}

type MusicArtistGroupDto struct {
	Key        string `json:"key"`
	Artist     string `json:"artist"`
	TrackCount int    `json:"track_count"`
	AlbumCount int    `json:"album_count"`
}

type MusicAlbumGroupDto struct {
	Key        string `json:"key"`
	Album      string `json:"album"`
	Artist     string `json:"artist"`
	Year       string `json:"year"`
	TrackCount int    `json:"track_count"`
}

type MusicGenreGroupDto struct {
	Key        string `json:"key"`
	Genre      string `json:"genre"`
	TrackCount int    `json:"track_count"`
}

type MusicFolderGroupDto struct {
	Folder     string `json:"folder"`
	TrackCount int    `json:"track_count"`
}

type MusicLibrarySummaryDto struct {
	TotalTracks  int `json:"total_tracks"`
	TotalArtists int `json:"total_artists"`
	TotalAlbums  int `json:"total_albums"`
	TotalGenres  int `json:"total_genres"`
	TotalFolders int `json:"total_folders"`
}

type MusicHomeCatalogDto struct {
	Summary   MusicLibrarySummaryDto `json:"summary"`
	Playlists []PlaylistDto          `json:"playlists"`
	Artists   []MusicArtistGroupDto  `json:"artists"`
	Albums    []MusicAlbumGroupDto   `json:"albums"`
}
