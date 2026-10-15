UPDATE home_file
SET starred = $2,
    starred_at = CASE WHEN $2 THEN COALESCE(starred_at, now()) ELSE NULL END
WHERE id = $1 AND deleted_at IS NULL
RETURNING starred;
