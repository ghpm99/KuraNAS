-- Hard-deletes the soft-deleted row at a path and its soft-deleted subtree,
-- once their bytes were purged from the trash and can never be restored.
-- Active rows (a path recreated after the delete) are left untouched.
-- $1 = purged path, $2 = purged path + OS separator.
-- Literal prefix operator ^@, never LIKE (see update_descendant_paths.sql).
DELETE FROM home_file
WHERE
    deleted_at IS NOT NULL
    AND (
        path = $1
        OR path ^@ $2
    );
