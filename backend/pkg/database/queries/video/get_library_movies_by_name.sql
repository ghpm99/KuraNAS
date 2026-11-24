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
LEFT JOIN video_manual_override vmo ON vmo.file_id = hf.id
WHERE
    hf.format = ANY($1)
    AND hf.deleted_at IS NULL
    AND CASE
        WHEN vmo.classification IS NOT NULL THEN
            CASE vmo.classification WHEN 'other' THEN 'program' ELSE vmo.classification END = $2
        ELSE EXISTS (
            SELECT 1
            FROM video_metadata vm
            WHERE vm.file_id = hf.id
              AND vm.classification = $2
        )
    END
ORDER BY
    lower(hf.name), hf.id
LIMIT $3
OFFSET $4
