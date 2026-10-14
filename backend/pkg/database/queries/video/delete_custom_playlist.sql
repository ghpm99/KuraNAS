DELETE FROM video_playlist
WHERE id = $1
  AND is_auto = FALSE;
