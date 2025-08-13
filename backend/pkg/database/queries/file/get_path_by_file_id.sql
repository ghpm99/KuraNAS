SELECT
    hf.path
FROM
    home_file hf
WHERE
    hf.id = $1
    AND hf.deleted_at IS NULL;