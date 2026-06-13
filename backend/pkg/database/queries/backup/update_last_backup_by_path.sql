UPDATE home_file
SET last_backup = $2
WHERE path = $1
  AND deleted_at IS NULL;
