UPDATE video_playlist_item
SET order_index = $3
WHERE playlist_id = $1
  AND video_id = $2;
