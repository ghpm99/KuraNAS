UPDATE home_file
SET
    name = $2,
    "path" = $3,
    format = $4,
    "size" = $5,
    updated_at = $6,
    created_at = $7,
    last_interaction = $8,
    last_backup = $9
WHERE
    id = $1;