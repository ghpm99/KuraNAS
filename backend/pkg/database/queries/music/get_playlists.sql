SELECT
    p.id,
    p.name,
    p.description,
    p.is_system,
    p.created_at,
    p.updated_at,
    (SELECT COUNT(*) FROM playlist_track pt WHERE pt.playlist_id = p.id) AS track_count
FROM
    playlist p
WHERE
    p.is_system = FALSE
ORDER BY
    p.updated_at DESC
LIMIT
    $1
OFFSET
    $2;
