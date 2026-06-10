INSERT INTO
    recent_file (ip_address, file_id, accessed_at)
VALUES
    ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT (ip_address, file_id)
DO
UPDATE
SET
    accessed_at = CURRENT_TIMESTAMP;