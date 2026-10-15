UPDATE home_file
SET starred = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING starred;
