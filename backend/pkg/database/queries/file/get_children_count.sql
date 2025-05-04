SELECT
	count(*)
FROM
	home_file hf
WHERE
	1 = 1
	AND hf.parent_path = $1
	AND hf.id != $2