-- Every registered root, registration order (the first enabled is primary).
SELECT
    id,
    path,
    label,
    enabled,
    created_at
FROM
    storage_root
ORDER BY
    id ASC;
