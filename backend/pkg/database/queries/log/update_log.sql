UPDATE log
SET
    name = ?,
    description = ?,
    level = ?,
    ip_address = ?,
    start_time = ?,
    end_time = ?,
    status = ?,
    extra_data = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE
    id = ?;