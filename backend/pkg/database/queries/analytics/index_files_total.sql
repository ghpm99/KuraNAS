SELECT COUNT(*)
FROM home_file
WHERE type = 2
	AND deleted_at IS NULL;
