SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.size,
    hf.created_at,
    hf.updated_at
FROM
    home_file hf
WHERE
    hf.parent_path = $2
    AND hf.format = ANY($1)
    AND hf.deleted_at IS NULL
ORDER BY
    ARRAY(
        SELECT CASE WHEN piece[1] ~ '^[0-9]' THEN lpad(piece[1], 20, '0') ELSE piece[1] END
        FROM regexp_matches(lower(hf.name), '([0-9]+|[^0-9]+)', 'g') AS piece
    ),
    hf.id
LIMIT $3
OFFSET $4
