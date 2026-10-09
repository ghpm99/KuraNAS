SELECT
	hf.parent_path,
	COUNT(*)
FROM
	home_file hf
WHERE
	hf.parent_path = ANY($1)
	AND hf.path != hf.parent_path
	AND hf.deleted_at IS NULL
GROUP BY
	hf.parent_path;
