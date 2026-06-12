-- Every item in the trash — empty-trash feed.
SELECT
    id,
    original_path,
    trash_path,
    size,
    deleted_at
FROM
    trash_item
ORDER BY
    deleted_at ASC,
    id ASC;
