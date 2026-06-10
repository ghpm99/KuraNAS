SELECT
    hf.format,
    COUNT(*) AS total_itens,
    SUM(hf.size) AS total_size
FROM
    home_file hf
WHERE
    hf.type = $1
GROUP BY
    hf.format;