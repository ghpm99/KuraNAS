SELECT hf.parent_path
FROM video_playlist_item vpi
JOIN home_file hf ON hf.id = vpi.video_id
WHERE vpi.playlist_id = $1
  AND hf.deleted_at IS NULL
GROUP BY hf.parent_path
ORDER BY COUNT(*) DESC, hf.parent_path
LIMIT 1;
