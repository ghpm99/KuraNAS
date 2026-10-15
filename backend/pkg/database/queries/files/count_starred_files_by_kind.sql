SELECT
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE hf.type = 1) AS folders,
    COUNT(*) FILTER (WHERE hf.type = 2 AND NOT (lower(hf.format) = ANY ($1::text[]))) AS files,
    COUNT(*) FILTER (WHERE hf.type = 2 AND lower(hf.format) = ANY ($1::text[])) AS media
FROM
    home_file hf
WHERE
    hf.deleted_at IS NULL
    AND hf.starred = TRUE;
