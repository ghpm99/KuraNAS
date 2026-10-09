SELECT
    COUNT(*) FILTER (WHERE hf.type = $2) AS file_count,
    COUNT(*) FILTER (WHERE hf.type = $3) AS folder_count,
    COALESCE(SUM(hf.size) FILTER (WHERE hf.type = $2), 0) AS total_size_bytes
FROM
    home_file hf
WHERE
    hf.deleted_at IS NULL
    AND hf.path ^@ $1;
