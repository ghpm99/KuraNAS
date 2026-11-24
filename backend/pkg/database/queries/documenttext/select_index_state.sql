SELECT
    CASE
        WHEN dt.file_id IS NULL OR dt.source_updated_at < hf.updated_at THEN 'queued'
        ELSE dt.state
    END AS state,
    CASE
        WHEN dt.file_id IS NULL OR dt.source_updated_at < hf.updated_at THEN 0
        ELSE dt.indexed_bytes
    END AS indexed_bytes,
    CASE
        WHEN dt.file_id IS NULL OR dt.source_updated_at < hf.updated_at THEN NULL
        ELSE dt.failure_reason
    END AS failure_reason
FROM
    home_file hf
    LEFT JOIN document_text dt ON dt.file_id = hf.id
WHERE
    hf.id = $1
    AND hf.deleted_at IS NULL
    AND hf.format = ANY($2);
