SELECT
    id,
    cidr,
    label,
    enabled,
    created_at
FROM
    allowed_ip
WHERE
    id = $1;
