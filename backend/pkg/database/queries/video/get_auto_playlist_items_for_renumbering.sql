SELECT
    vpi.id,
    vpi.video_id,
    vpi.order_index,
    hf.name,
    vpi.season,
    vpi.episode,
    mo.season,
    mo.episode
FROM video_playlist_item vpi
JOIN home_file hf ON hf.id = vpi.video_id
LEFT JOIN video_manual_override mo ON mo.file_id = vpi.video_id
WHERE vpi.playlist_id = $1
  AND vpi.source_kind = 'auto';
