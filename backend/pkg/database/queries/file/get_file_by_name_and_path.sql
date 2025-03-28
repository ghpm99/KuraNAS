SELECT
    hf.id,
    hf.name,
    hf."path",
    hf.format,
    hf."size",
    hf.updated_at,
    hf.created_at,
    hf.last_interaction,
    hf.last_backup
FROM
    home_file hf
where
    1 = 1
    AND hf.name = $1
    AND hf."path" = $2
LIMIT
    1