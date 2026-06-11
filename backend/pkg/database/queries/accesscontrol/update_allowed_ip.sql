UPDATE
    allowed_ip
SET
    cidr = $2,
    label = $3,
    enabled = $4
WHERE
    id = $1
RETURNING
    id,
    cidr,
    label,
    enabled,
    created_at;
