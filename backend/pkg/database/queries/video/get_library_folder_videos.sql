SELECT
    hf.id,
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
FROM
    home_file hf
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
WHERE
    hf.parent_path = $2
    AND hf.format = ANY($1)
    AND hf.deleted_at IS NULL
ORDER BY
    ARRAY(
        SELECT CASE WHEN piece[1] ~ '^[0-9]' THEN lpad(piece[1], 20, '0') ELSE piece[1] END
        FROM regexp_matches(lower(hf.name), '([0-9]+|[^0-9]+)', 'g') AS piece
    ),
    hf.id
LIMIT $3
OFFSET $4
