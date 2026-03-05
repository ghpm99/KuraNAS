SELECT COUNT(*)
FROM log
WHERE level = 'ERROR'
	AND created_at >= NOW() - INTERVAL '24 hours';
