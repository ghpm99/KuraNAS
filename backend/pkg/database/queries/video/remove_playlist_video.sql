DELETE FROM video_playlist_item
WHERE playlist_id = $1
  AND video_id = $2;
