SELECT
    vpi.id,
    vpi.playlist_id,
    vpi.video_id,
    vpi.order_index,
    vpi.source_kind,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.size,
    hf.created_at,
    hf.updated_at,
    COALESCE(hf.starred, FALSE),
    tech.duration_seconds,
    tech.width,
    tech.height,
    tech.codec_name
FROM video_playlist_item vpi
JOIN home_file hf ON hf.id = vpi.video_id
LEFT JOIN LATERAL (
    SELECT
        CASE
            WHEN vm.duration ~ '^[0-9]+(\.[0-9]+)?$' THEN vm.duration::double precision
        END AS duration_seconds,
        vm.width,
        vm.height,
        vm.codec_name
    FROM
        video_metadata vm
    WHERE
        vm.id = (
            SELECT
                MAX(latest.id)
            FROM
                video_metadata latest
            WHERE
                latest.file_id = hf.id
        )
) tech ON TRUE
WHERE vpi.playlist_id = $1
  AND hf.deleted_at IS NULL
ORDER BY vpi.order_index, vpi.id
LIMIT $2 OFFSET $3;
