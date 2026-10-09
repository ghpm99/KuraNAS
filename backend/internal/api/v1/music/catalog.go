package music

const (
	AutoPlaylistContinueListeningID = -1
	AutoPlaylistRecentlyAddedID     = -2
	AutoPlaylistFavoritesID         = -3
)

const (
	PlaylistKindManual    = "manual"
	PlaylistKindSystem    = "system"
	PlaylistKindAutomatic = "automatic"
	PlaylistKindAI        = "ai"
)

const (
	autoPlaylistContinueListeningKey = "continue-listening"
	autoPlaylistRecentlyAddedKey     = "recently-added"
	autoPlaylistFavoritesKey         = "favorites"
)

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

type MusicAlbumSummaryDto struct {
	Key                string `json:"key"`
	Name               string `json:"name"`
	Artist             string `json:"artist"`
	Year               string `json:"year"`
	TrackCount         int    `json:"track_count"`
	TotalLengthSeconds int64  `json:"total_length_seconds"`
	DiscCount          int    `json:"disc_count"`
}

type MusicArtistSummaryDto struct {
	Key                string `json:"key"`
	Name               string `json:"name"`
	TrackCount         int    `json:"track_count"`
	AlbumCount         int    `json:"album_count"`
	TotalLengthSeconds int64  `json:"total_length_seconds"`
}

type MusicGroupSummaryDto struct {
	Key                string `json:"key"`
	Name               string `json:"name"`
	TrackCount         int    `json:"track_count"`
	TotalLengthSeconds int64  `json:"total_length_seconds"`
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
