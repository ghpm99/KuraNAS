SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.starred,
    hf.size,
    hf.updated_at,
    COALESCE(hf.physical_path, '') <> '' AS is_cold
FROM
    home_file hf
WHERE
    hf.deleted_at IS NULL
    AND hf.type = 1
    AND kuranas_fold(hf.name) % kuranas_fold($1)
    AND similarity(kuranas_fold(hf.name), kuranas_fold($1)) > 0.3
ORDER BY
    similarity(kuranas_fold(hf.name), kuranas_fold($1)) DESC,
    hf.starred DESC,
    hf.name ASC
LIMIT
    $2;
