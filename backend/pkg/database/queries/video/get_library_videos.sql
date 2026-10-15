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
    hf.format = ANY($1)
    AND hf.deleted_at IS NULL
    AND (
        $2 = ''
        OR kuranas_fold(hf.name) LIKE kuranas_fold($2)
        OR kuranas_fold(hf.parent_path) LIKE kuranas_fold($2)
    )
    AND ($5::text[] IS NULL OR hf.format = ANY($5))
    AND ($6::int IS NULL OR tech.height >= $6)
    AND ($7::int IS NULL OR tech.height < $7)
    AND ($8::double precision IS NULL OR tech.duration_seconds >= $8)
    AND ($9::double precision IS NULL OR tech.duration_seconds < $9)
