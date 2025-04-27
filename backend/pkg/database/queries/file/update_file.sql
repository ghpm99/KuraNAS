UPDATE home_file
SET
    name = $1,
    "path" = $2,
    parent_path = $3,
    format = $4,
    "size" = $5,
    updated_at = $6,
    created_at = $7,
    last_interaction = $8,
    last_backup = $9,
    "type"=$10,
    checksum=$11,
    deleted_at=$12
WHERE
    id = $13;
