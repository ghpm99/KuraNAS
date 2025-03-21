SELECT
    *
FROM
    home_file hf
WHERE
    1 = 1
    AND hf.format = $3
LIMIT
    $1 OFFSET $2;