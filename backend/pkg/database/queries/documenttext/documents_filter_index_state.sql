(
    CASE
        WHEN dt.file_id IS NULL OR dt.source_updated_at < hf.updated_at THEN 'queued'
        ELSE dt.state
    END
) = @1
