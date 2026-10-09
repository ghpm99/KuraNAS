SELECT
    hf.id,
    hf.path,
    hf.format,
    hf.size,
    hf.updated_at
FROM
    home_file hf
    LEFT JOIN document_text dt ON dt.file_id = hf.id
WHERE
    hf.deleted_at IS NULL
    AND hf.format = ANY($1)
    AND hf.id > $2
    AND (
        dt.file_id IS NULL
        OR dt.source_updated_at < hf.updated_at
    )
ORDER BY
    hf.id
LIMIT
    $3;
