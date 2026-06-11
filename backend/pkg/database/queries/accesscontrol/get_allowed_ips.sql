SELECT
    id,
    cidr,
    label,
    enabled,
    created_at
FROM
    allowed_ip
ORDER BY
    id;
