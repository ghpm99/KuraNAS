-- Revives the soft-deleted row at a path and its whole subtree — the inverse
-- of mark_deleted_subtree, used when a trash restore puts the bytes back.
-- The restored entry itself lands on the hot volume, so its cold location is
-- cleared; descendants keep theirs because their cold copies never moved.
-- $1 = restored path, $2 = restored path + OS separator.
-- Literal prefix operator ^@, never LIKE (see update_descendant_paths.sql).
UPDATE home_file
SET
    deleted_at = NULL,
    physical_path = CASE WHEN path = $1 THEN NULL ELSE physical_path END
WHERE
    deleted_at IS NOT NULL
    AND (
        path = $1
        OR path ^@ $2
    );
