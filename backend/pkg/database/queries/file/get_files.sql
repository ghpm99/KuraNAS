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
ORDER BY
    - id;