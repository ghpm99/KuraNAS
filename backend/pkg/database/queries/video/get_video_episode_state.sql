SELECT
    hf.name,
    vmo.season,
    vmo.episode,
    auto_item.season,
    auto_item.episode
FROM home_file hf
LEFT JOIN video_manual_override vmo ON vmo.file_id = hf.id
LEFT JOIN LATERAL (
    SELECT vpi.season, vpi.episode
    FROM video_playlist_item vpi
    WHERE vpi.video_id = hf.id
      AND vpi.source_kind = 'auto'
      AND (vpi.season IS NOT NULL OR vpi.episode IS NOT NULL)
    ORDER BY vpi.id DESC
    LIMIT 1
) auto_item ON TRUE
WHERE hf.id = $1;
