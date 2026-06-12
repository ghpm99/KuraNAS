-- Paginated trash listing, newest deletions first.
SELECT
    id,
    original_path,
    trash_path,
    size,
    deleted_at
FROM
    trash_item
ORDER BY
    deleted_at DESC,
    id DESC
LIMIT
    $1
OFFSET
    $2;
