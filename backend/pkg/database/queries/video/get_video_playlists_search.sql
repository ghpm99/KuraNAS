SELECT
    vp.id,
    vp.type,
    vp.source_path,
    vp.name,
    vp.name_key,
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
WHERE vp.is_hidden = FALSE
  AND (
      $1::text = ''
      OR (kuranas_fold(vp.name) LIKE kuranas_fold($1::text) AND vp.name_key IS NULL)
      OR vp.name_key = ANY($4::text[])
  )
  AND (
      vp.is_auto = FALSE
      OR EXISTS (
          SELECT 1
          FROM video_playlist_item active_item
          JOIN home_file active_file ON active_file.id = active_item.video_id
          WHERE active_item.playlist_id = vp.id
            AND active_file.deleted_at IS NULL
      )
  )
GROUP BY vp.id, cover.video_id
ORDER BY vp.last_played_at DESC NULLS LAST, vp.updated_at DESC, vp.id DESC
LIMIT $2 OFFSET $3;
