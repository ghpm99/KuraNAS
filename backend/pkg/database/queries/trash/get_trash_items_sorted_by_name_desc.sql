-- Paginated trash listing ordered by name DESC.
-- $1 = limit, $2 = offset, $3 = lowercase search text ('' lists everything).
-- Literal substring match (position), never LIKE: Windows paths carry '\'.
SELECT
    id,
    original_path,
    trash_path,
    size,
    deleted_at,
    is_directory
FROM
    trash_item
WHERE
    $3 = ''
    OR position($3 IN lower(original_path)) > 0
ORDER BY
    lower(regexp_replace(original_path, '^.*[/\\]', '')) DESC,
    id DESC
LIMIT
    $1
OFFSET
    $2;
