SELECT
    hf."path"
FROM
    home_file hf
WHERE
    1 = 1
    AND hf.id = $1
    AND hf.deleted_at IS NULL