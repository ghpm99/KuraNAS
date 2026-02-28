UPDATE playlist_track
SET position = $1
WHERE playlist_id = $2 AND file_id = $3;
