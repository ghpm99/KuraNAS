SELECT
    hf.id
FROM
    home_file hf
WHERE
    hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
    AND hf.starred = TRUE
ORDER BY
    COALESCE(hf.updated_at, hf.created_at) DESC,
    hf.id DESC
LIMIT
    $2;
