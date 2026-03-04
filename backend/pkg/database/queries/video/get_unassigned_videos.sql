SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.size,
    hf.created_at,
    hf.updated_at
FROM home_file hf
WHERE hf.deleted_at IS NULL
  AND hf.format = ANY($1)
  AND NOT EXISTS (
      SELECT 1 FROM video_playlist_item vpi WHERE vpi.video_id = hf.id
  )
ORDER BY hf.updated_at DESC, hf.id DESC
LIMIT $2;
