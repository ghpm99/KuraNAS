-- Items past the retention window (deleted before the cutoff) — purge feed.
SELECT
    id,
    original_path,
    trash_path,
    size,
    deleted_at
FROM
    trash_item
WHERE
    deleted_at < $1
ORDER BY
    deleted_at ASC,
    id ASC;
