DELETE FROM video_playlist_exclusion
WHERE playlist_id = $1
  AND video_id = $2;
