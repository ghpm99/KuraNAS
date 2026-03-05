SELECT
	name,
	description,
	created_at
FROM log
WHERE level = 'ERROR'
ORDER BY created_at DESC
LIMIT $1;
