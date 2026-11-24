SELECT
    hf.id,
    hf.name,
    hf.format,
    hf.parent_path,
    hf.size,
    hf.updated_at,
    hf.created_at,
    hf.starred,
    CASE
        WHEN dt.file_id IS NULL OR dt.source_updated_at < hf.updated_at THEN 'queued'
        ELSE dt.state
    END AS index_state
