SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.size,
    hf.created_at,
    hf.updated_at
FROM
    home_file hf
WHERE
    hf.format = ANY($1)
    AND hf.deleted_at IS NULL
    AND EXISTS (
        SELECT 1
        FROM video_metadata vm
        WHERE vm.file_id = hf.id
          AND vm.classification = $2
    )
ORDER BY
    lower(hf.name), hf.id
LIMIT $3
OFFSET $4
