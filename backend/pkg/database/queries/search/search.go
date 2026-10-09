package queries

import (
	_ "embed"
)

//go:embed search_files.sql
var SearchFilesQuery string

//go:embed search_folders.sql
var SearchFoldersQuery string

//go:embed search_artists.sql
var SearchArtistsQuery string

//go:embed search_albums.sql
var SearchAlbumsQuery string

//go:embed search_music_playlists.sql
var SearchMusicPlaylistsQuery string

//go:embed search_video_playlists.sql
var SearchVideoPlaylistsQuery string

//go:embed search_videos.sql
var SearchVideosQuery string

//go:embed search_images.sql
var SearchImagesQuery string

//go:embed search_tracks.sql
var SearchTracksQuery string

//go:embed search_files_fuzzy.sql
var SearchFilesFuzzyQuery string

//go:embed search_folders_fuzzy.sql
var SearchFoldersFuzzyQuery string

//go:embed check_fuzzy_search_support.sql
var CheckFuzzySearchSupportQuery string
