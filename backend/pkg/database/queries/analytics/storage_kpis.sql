SELECT
	COALESCE(SUM(CASE WHEN type = 2 AND deleted_at IS NULL THEN size ELSE 0 END), 0) AS used_bytes,
	COALESCE(SUM(CASE WHEN type = 2 AND deleted_at IS NULL AND created_at >= NOW() - $1::interval THEN size ELSE 0 END), 0) AS growth_bytes,
	COALESCE(SUM(CASE WHEN type = 2 AND deleted_at IS NULL AND created_at >= NOW() - $1::interval THEN 1 ELSE 0 END), 0) AS files_added,
	COALESCE(SUM(CASE WHEN type = 2 AND deleted_at IS NULL THEN 1 ELSE 0 END), 0) AS files_total,
	COALESCE(SUM(CASE WHEN type = 1 AND deleted_at IS NULL THEN 1 ELSE 0 END), 0) AS folders_total
FROM home_file;
