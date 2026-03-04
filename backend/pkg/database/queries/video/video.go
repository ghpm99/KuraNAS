package queries

import (
	_ "embed"
)

//go:embed get_video_file_by_id.sql
var GetVideoFileByIDQuery string

//go:embed get_videos_by_parent_path.sql
var GetVideosByParentPathQuery string

//go:embed get_playlist_by_context.sql
var GetPlaylistByContextQuery string

//go:embed create_playlist.sql
var CreatePlaylistQuery string

//go:embed delete_playlist_items.sql
var DeletePlaylistItemsQuery string

//go:embed insert_playlist_items.sql
var InsertPlaylistItemsQuery string

//go:embed get_playlist_items.sql
var GetPlaylistItemsQuery string

//go:embed get_playback_state.sql
var GetPlaybackStateQuery string

//go:embed upsert_playback_state.sql
var UpsertPlaybackStateQuery string

//go:embed touch_playlist.sql
var TouchPlaylistQuery string

//go:embed get_catalog_videos.sql
var GetCatalogVideosQuery string

//go:embed get_recent_videos.sql
var GetRecentVideosQuery string
