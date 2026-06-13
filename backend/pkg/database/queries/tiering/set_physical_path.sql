-- $2 NULL promotes the file back to hot (bytes at the logical path); a value
-- demotes it. updated_at is deliberately left untouched: physical_path is a
-- storage-location change, not a content change, so the scanner's size+mtime
-- diff and the backup must not treat the file as modified.
UPDATE home_file
SET physical_path = $2
WHERE id = $1;
