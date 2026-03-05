INSERT INTO
    LOG(NAME, description, LEVEL, ip_address, start_time, end_time, status, extra_data)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING
    id;