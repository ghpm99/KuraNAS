-- Soft-deletes a path and its whole subtree in one statement.
-- $1 = deleted path, $2 = deletion timestamp, $3 = deleted path + OS separator.
-- Literal starts_with(), never LIKE (see update_descendant_paths.sql).
UPDATE home_file
SET
    deleted_at = $2
WHERE
    deleted_at IS NULL
    AND (
        path = $1
        OR starts_with(path, $3)
    );
