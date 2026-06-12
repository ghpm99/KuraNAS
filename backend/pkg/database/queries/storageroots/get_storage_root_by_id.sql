-- Point lookup of one storage root.
SELECT
    id,
    path,
    label,
    enabled,
    created_at
FROM
    storage_root
WHERE
    id = $1;
