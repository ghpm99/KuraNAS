SELECT
    hf.id,
    hf.path,
    hf.format,
    hf.size,
    hf.updated_at
FROM
    home_file hf
WHERE
    hf.id = $1
    AND hf.deleted_at IS NULL
    AND hf.format = ANY($2);
