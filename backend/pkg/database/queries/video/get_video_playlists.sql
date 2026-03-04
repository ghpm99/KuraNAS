SELECT
    vp.id,
    vp.type,
    vp.source_path,
    vp.name,
    vp.is_hidden,
    vp.is_auto,
    vp.group_mode,
    vp.classification,
    vp.created_at,
    vp.updated_at,
    vp.last_played_at,
    COUNT(vpi.video_id) AS item_count
FROM video_playlist vp
LEFT JOIN video_playlist_item vpi ON vpi.playlist_id = vp.id
WHERE ($1 = TRUE OR vp.is_hidden = FALSE)
GROUP BY vp.id
ORDER BY vp.last_played_at DESC NULLS LAST, vp.updated_at DESC;
