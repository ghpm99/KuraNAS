UPDATE home_file
SET
    NAME = $1,
    PATH = $2,
    parent_path = $3,
    FORMAT = $4,
    size = $5,
    updated_at = $6,
    created_at = $7,
    last_interaction = $8,
    last_backup = $9,
TYPE = $10,
checksum = $11,
deleted_at = $12,
starred = $13
WHERE
    id = $14;