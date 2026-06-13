SELECT COUNT(*)
FROM home_file hf
WHERE hf.type = 2
  AND hf.deleted_at IS NULL
  AND (hf.last_backup IS NULL OR hf.last_backup < hf.updated_at);
