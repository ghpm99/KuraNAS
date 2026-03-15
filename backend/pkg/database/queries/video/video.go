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

//go:embed get_all_videos_for_grouping.sql
var GetAllVideosForGroupingQuery string

//go:embed upsert_auto_playlist.sql
var UpsertAutoPlaylistQuery string

//go:embed delete_auto_playlist_items.sql
var DeleteAutoPlaylistItemsQuery string

//go:embed insert_playlist_items_with_source.sql
var InsertPlaylistItemsWithSourceQuery string

//go:embed get_playlist_exclusions.sql
var GetPlaylistExclusionsQuery string

//go:embed get_video_playlists.sql
var GetVideoPlaylistsQuery string

//go:embed get_video_playlist_memberships.sql
var GetVideoPlaylistMembershipsQuery string

//go:embed get_video_playlist_by_id.sql
var GetVideoPlaylistByIDQuery string

//go:embed get_video_playlist_items_detailed.sql
var GetVideoPlaylistItemsDetailedQuery string

//go:embed set_playlist_hidden.sql
var SetPlaylistHiddenQuery string

//go:embed add_playlist_video_manual.sql
var AddPlaylistVideoManualQuery string

//go:embed remove_playlist_video.sql
var RemovePlaylistVideoQuery string

//go:embed upsert_playlist_exclusion.sql
var UpsertPlaylistExclusionQuery string

//go:embed delete_playlist_exclusion.sql
var DeletePlaylistExclusionQuery string

//go:embed get_unassigned_videos.sql
var GetUnassignedVideosQuery string

//go:embed check_video_in_playlist.sql
var CheckVideoInPlaylistQuery string

//go:embed update_playlist_name.sql
var UpdatePlaylistNameQuery string

//go:embed reorder_playlist_item.sql
var ReorderPlaylistItemQuery string

//go:embed insert_behavior_event.sql
var InsertBehaviorEventQuery string

//go:embed get_behavior_events.sql
var GetBehaviorEventsQuery string

//go:embed get_all_behavior_events.sql
var GetAllBehaviorEventsQuery string

//go:embed get_all_videos_with_metadata.sql
var GetAllVideosWithMetadataQuery string

//go:embed get_library_videos.sql
var GetLibraryVideosQuery string
