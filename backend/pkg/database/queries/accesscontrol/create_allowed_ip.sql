INSERT INTO
    allowed_ip (cidr, label, enabled)
VALUES
    ($1, $2, $3)
RETURNING
    id,
    cidr,
    label,
    enabled,
    created_at;
