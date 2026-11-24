SELECT
    p.id,
    p.name,
    p.description,
    p.is_system,
    p.created_at,
    p.updated_at,
    (SELECT COUNT(*) FROM playlist_track pt WHERE pt.playlist_id = p.id) AS track_count,
    p.is_ai_generated,
    p.kind,
    p.rules,
    COALESCE((
        SELECT SUM(am.length)
        FROM playlist_track pt
            INNER JOIN home_file hf ON hf.id = pt.file_id AND hf.deleted_at IS NULL
            INNER JOIN audio_metadata_effective am ON am.file_id = hf.id
        WHERE pt.playlist_id = p.id
    ), 0)::float8 AS total_length_seconds
FROM
    playlist p
WHERE
    p.is_ai_generated = TRUE
ORDER BY
    p.name;
