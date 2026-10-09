package search

import "nas-go/api/internal/api/v1/documenttext"

type RepositoryInterface interface {
	SearchFiles(query string, limit int) ([]FileResultModel, error)
	SearchFolders(query string, limit int) ([]FolderResultModel, error)
	SearchArtists(query string, limit int) ([]ArtistResultModel, error)
	SearchAlbums(query string, limit int) ([]AlbumResultModel, error)
	SearchMusicPlaylists(query string, limit int) ([]MusicPlaylistResultModel, error)
	SearchVideoPlaylists(query string, limit int) ([]VideoPlaylistResultModel, error)
	SearchVideos(query string, limit int) ([]VideoResultModel, error)
	SearchImages(query string, limit int) ([]ImageResultModel, error)
	SearchTracks(query string, limit int) ([]TrackResultModel, error)
	IsFuzzySearchAvailable() bool
	SearchFilesFuzzy(query string, limit int) ([]FileResultModel, error)
	SearchFoldersFuzzy(query string, limit int) ([]FolderResultModel, error)
}

type ServiceInterface interface {
	SearchGlobal(query string, limit int) (GlobalSearchResponseDto, error)
	SearchGlobalWithAI(query string, limit int) (GlobalSearchResponseDto, error)
}

type DocumentSearcher interface {
	SearchTopDocuments(query string, limit int) ([]documenttext.DocumentSearchResultDto, error)
}
