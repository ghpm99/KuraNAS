-- Point lookup of one trash item.
SELECT
    id,
    original_path,
    trash_path,
    size,
    deleted_at
FROM
    trash_item
WHERE
    id = $1;
