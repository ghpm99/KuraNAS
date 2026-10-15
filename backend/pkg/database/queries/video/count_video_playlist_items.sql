SELECT COUNT(*)
FROM video_playlist_item vpi
JOIN home_file hf ON hf.id = vpi.video_id
WHERE vpi.playlist_id = $1
  AND hf.deleted_at IS NULL;
