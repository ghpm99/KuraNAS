-- Counts the trash items whose original path is the given path, lives under
-- it, or contains it — any of them may still restore rows of that subtree.
-- $1 = path, $2 = OS separator.
-- Literal prefix operator ^@, never LIKE.
SELECT COUNT(*)
FROM trash_item
WHERE
    original_path = $1
    OR original_path ^@ ($1 || $2)
    OR $1 ^@ (original_path || $2);
