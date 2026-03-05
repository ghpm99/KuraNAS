SELECT
	status,
	start_time,
	end_time
FROM log
WHERE name IN ('UpdateFiles', 'ScanFiles', 'ScanDir')
ORDER BY start_time DESC
LIMIT 1;
