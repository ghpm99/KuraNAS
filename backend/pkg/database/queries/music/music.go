package queries

import (
	_ "embed"
)

//go:embed get_playlists.sql
var GetPlaylistsQuery string

//go:embed get_playlist_by_id.sql
var GetPlaylistByIDQuery string

//go:embed create_playlist.sql
var CreatePlaylistQuery string

//go:embed update_playlist.sql
var UpdatePlaylistQuery string

//go:embed delete_playlist.sql
var DeletePlaylistQuery string

//go:embed get_playlist_tracks.sql
var GetPlaylistTracksQuery string

//go:embed add_playlist_track.sql
var AddPlaylistTrackQuery string

//go:embed remove_playlist_track.sql
var RemovePlaylistTrackQuery string

//go:embed reorder_playlist_track.sql
var ReorderPlaylistTrackQuery string

//go:embed get_now_playing.sql
var GetNowPlayingQuery string

//go:embed get_player_state.sql
var GetPlayerStateQuery string

//go:embed upsert_player_state.sql
var UpsertPlayerStateQuery string
