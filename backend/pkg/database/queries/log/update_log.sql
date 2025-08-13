UPDATE LOG
SET
    NAME = $1,
    description = $2,
    LEVEL = $3,
    ip_address = $4,
    start_time = $5,
    end_time = $6,
    status = $7,
    extra_data = $8,
    updated_at = CURRENT_TIMESTAMP
WHERE
    id = $9;