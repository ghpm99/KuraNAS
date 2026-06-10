SELECT
    hf.id,
    hf.name,
    hf.size,
    hf.path
FROM
    home_file hf
ORDER BY
    hf.size DESC
LIMIT
    $1;