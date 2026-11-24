ALTER TABLE document_text
    ADD COLUMN IF NOT EXISTS state TEXT NOT NULL DEFAULT 'indexed',
    ADD COLUMN IF NOT EXISTS indexed_bytes BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS failure_reason TEXT NULL;

UPDATE document_text
SET
    state = CASE
        WHEN error IN ('encrypted', 'too_large', 'unsupported_format') THEN 'name_only'
        WHEN error IS NOT NULL THEN 'failed'
        WHEN extracted_text = '' THEN 'name_only'
        WHEN truncated THEN 'partial'
        ELSE 'indexed'
    END,
    indexed_bytes = octet_length(extracted_text),
    failure_reason = CASE
        WHEN error IS NOT NULL THEN error
        WHEN extracted_text = '' THEN 'no_text'
        ELSE NULL
    END;

UPDATE document_text
SET source_updated_at = 'epoch'
WHERE truncated AND error IS NULL;

ALTER TABLE document_text
    ALTER COLUMN state DROP DEFAULT;

ALTER TABLE document_text
    DROP CONSTRAINT IF EXISTS document_text_state_check;

ALTER TABLE document_text
    ADD CONSTRAINT document_text_state_check
    CHECK (state IN ('indexed', 'partial', 'name_only', 'failed'));
