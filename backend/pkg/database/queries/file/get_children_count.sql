SELECT
	COUNT(*)
FROM
	home_file hf
WHERE
	hf.parent_path = $1
	AND hf.id != $2;