SELECT
    COUNT(*)
FROM
    home_file hf
WHERE
    hf.type = $1;