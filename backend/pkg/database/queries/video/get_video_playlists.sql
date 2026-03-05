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
    COUNT(vpi.video_id) AS item_count,
    cover.video_id AS cover_video_id
FROM video_playlist vp
LEFT JOIN video_playlist_item vpi ON vpi.playlist_id = vp.id
LEFT JOIN LATERAL (
    SELECT vpi2.video_id
    FROM video_playlist_item vpi2
    WHERE vpi2.playlist_id = vp.id
    ORDER BY vpi2.order_index
    LIMIT 1
) cover ON TRUE
WHERE ($1 = TRUE OR vp.is_hidden = FALSE)
GROUP BY vp.id, cover.video_id
ORDER BY vp.last_played_at DESC NULLS LAST, vp.updated_at DESC;
