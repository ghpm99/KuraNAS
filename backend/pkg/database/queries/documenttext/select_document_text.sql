SELECT
    CASE
        WHEN dt.file_id IS NULL OR dt.source_updated_at < hf.updated_at THEN ''
        ELSE COALESCE(dt.extracted_text, '')
    END AS extracted_text,
    COALESCE(dt.state = 'partial' AND dt.source_updated_at >= hf.updated_at, FALSE) AS is_partial
FROM
    home_file hf
    LEFT JOIN document_text dt ON dt.file_id = hf.id
WHERE
    hf.id = $1
    AND hf.deleted_at IS NULL
    AND hf.format = ANY($2);
