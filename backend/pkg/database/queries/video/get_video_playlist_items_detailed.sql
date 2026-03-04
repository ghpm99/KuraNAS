SELECT
    vpi.id,
    vpi.playlist_id,
    vpi.video_id,
    vpi.order_index,
    vpi.source_kind,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.size,
    hf.created_at,
    hf.updated_at
FROM video_playlist_item vpi
JOIN home_file hf ON hf.id = vpi.video_id
WHERE vpi.playlist_id = $1
  AND hf.deleted_at IS NULL
ORDER BY vpi.order_index, vpi.id;
