SELECT
    vp.id,
    vp.name,
    vp.name_key,
    vp.type,
    vp.classification,
    vp.source_path,
    vp.is_auto,
    vp.updated_at,
    COUNT(vpi.video_id) AS item_count
FROM
    video_playlist vp
    LEFT JOIN video_playlist_item vpi ON vpi.playlist_id = vp.id
WHERE
    vp.is_hidden = FALSE
    AND (
        (vp.name ILIKE '%' || $1 || '%' AND vp.name_key IS NULL)
        OR vp.name_key = ANY($3::text[])
        OR vp.source_path ILIKE '%' || $1 || '%'
    )
GROUP BY
    vp.id
ORDER BY
    CASE
        WHEN vp.name_key IS NULL AND LOWER(vp.name) = LOWER($1) THEN 0
        WHEN vp.name_key = ANY($4::text[]) THEN 0
        WHEN vp.name_key IS NULL AND vp.name ILIKE $1 || '%' THEN 1
        WHEN vp.name_key = ANY($5::text[]) THEN 1
        ELSE 2
    END,
    vp.updated_at DESC,
    vp.name ASC
LIMIT
    $2;
