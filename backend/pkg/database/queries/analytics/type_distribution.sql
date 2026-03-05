SELECT
	CASE
		WHEN lower(COALESCE(format, '')) IN ('.mp4', '.mkv', '.avi', '.mov', '.wmv', '.webm', '.m4v') THEN 'video'
		WHEN lower(COALESCE(format, '')) IN ('.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp', '.svg', '.heic') THEN 'image'
		WHEN lower(COALESCE(format, '')) IN ('.mp3', '.flac', '.wav', '.ogg', '.aac', '.m4a') THEN 'audio'
		WHEN lower(COALESCE(format, '')) IN ('.pdf', '.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx', '.txt', '.md', '.csv') THEN 'document'
		WHEN lower(COALESCE(format, '')) IN ('.zip', '.rar', '.7z', '.tar', '.gz', '.bz2', '.xz') THEN 'archive'
		ELSE 'other'
	END AS category,
	COUNT(*) AS total_count,
	COALESCE(SUM(size), 0) AS total_bytes
FROM home_file
WHERE type = 2
	AND deleted_at IS NULL
GROUP BY category
ORDER BY total_bytes DESC;
