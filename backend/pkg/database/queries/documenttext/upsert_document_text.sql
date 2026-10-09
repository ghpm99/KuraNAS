INSERT INTO document_text (
    file_id,
    extracted_text,
    text_length,
    truncated,
    extracted_at,
    source_updated_at,
    error
)
VALUES ($1, $2, $3, $4, now(), $5, $6)
ON CONFLICT (file_id) DO UPDATE SET
    extracted_text = EXCLUDED.extracted_text,
    text_length = EXCLUDED.text_length,
    truncated = EXCLUDED.truncated,
    extracted_at = EXCLUDED.extracted_at,
    source_updated_at = EXCLUDED.source_updated_at,
    error = EXCLUDED.error;
