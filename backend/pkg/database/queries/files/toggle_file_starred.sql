UPDATE home_file
SET starred = NOT COALESCE(starred, FALSE),
    starred_at = CASE WHEN NOT COALESCE(starred, FALSE) THEN now() ELSE NULL END
WHERE id = $1 AND deleted_at IS NULL
RETURNING starred;
