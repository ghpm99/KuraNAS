INSERT INTO
    home_file (
        NAME,
        PATH,
        parent_path,
        FORMAT,
        size,
        updated_at,
        created_at,
        last_interaction,
        last_backup,
        deleted_at,
        TYPE,
        checksum
    )
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING
    id;