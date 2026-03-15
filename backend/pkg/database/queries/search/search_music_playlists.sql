SELECT
    p.id,
    p.name,
    p.description,
    p.is_system,
    p.updated_at,
    (SELECT COUNT(*) FROM playlist_track pt WHERE pt.playlist_id = p.id) AS track_count
FROM
    playlist p
WHERE
    p.name ILIKE '%' || $1 || '%'
    OR p.description ILIKE '%' || $1 || '%'
ORDER BY
    CASE
        WHEN LOWER(p.name) = LOWER($1) THEN 0
        WHEN p.name ILIKE $1 || '%' THEN 1
        ELSE 2
    END,
    p.updated_at DESC,
    p.name ASC
LIMIT
    $2;
