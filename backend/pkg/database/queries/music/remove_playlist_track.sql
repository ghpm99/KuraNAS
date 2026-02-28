DELETE FROM playlist_track
WHERE playlist_id = $1 AND file_id = $2;
