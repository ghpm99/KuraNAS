INSERT INTO document_text (
    file_id,
    extracted_text,
    text_length,
    truncated,
    extracted_at,
    source_updated_at,
    error,
    state,
    indexed_bytes,
    failure_reason
)
VALUES ($1, $2, $3, $4, now(), $5, $6, $7, $8, $9)
ON CONFLICT (file_id) DO UPDATE SET
    extracted_text = EXCLUDED.extracted_text,
    text_length = EXCLUDED.text_length,
    truncated = EXCLUDED.truncated,
    extracted_at = EXCLUDED.extracted_at,
    source_updated_at = EXCLUDED.source_updated_at,
    error = EXCLUDED.error,
    state = EXCLUDED.state,
    indexed_bytes = EXCLUDED.indexed_bytes,
    failure_reason = EXCLUDED.failure_reason;
