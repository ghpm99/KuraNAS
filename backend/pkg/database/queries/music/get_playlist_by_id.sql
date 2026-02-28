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
    p.id = $1;
