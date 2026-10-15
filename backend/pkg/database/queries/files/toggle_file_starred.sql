UPDATE home_file
SET starred = NOT starred
WHERE id = $1 AND deleted_at IS NULL
RETURNING starred;
