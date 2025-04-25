UPDATE home_file
SET
    name = $1,
    "path" = $2,
    format = $3,
    "size" = $4,
    updated_at = $5,
    created_at = $6,
    last_interaction = $7,
    last_backup = $8,
    "type"=$9,
    checksum=$10,
    deleted_at=$11
WHERE
    id = $12;
