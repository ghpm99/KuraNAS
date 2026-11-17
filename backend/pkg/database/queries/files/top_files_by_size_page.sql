SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.type,
    hf.size,
    hf.updated_at,
    hf.created_at,
    hf.physical_path
FROM
    home_file hf
WHERE
    hf.type = 2
    AND hf.deleted_at IS NULL
ORDER BY
    hf.size DESC,
    hf.id ASC
LIMIT
    $1
OFFSET
    $2;
