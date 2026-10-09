SELECT
    hf.id
FROM
    home_file hf
    LEFT JOIN audio_metadata am ON am.file_id = hf.id
WHERE
    hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
    AND (
        hf.parent_path = $2
        OR hf.parent_path LIKE $3
    )
ORDER BY
    hf.parent_path COLLATE "C" ASC,
    COALESCE(am.disc_number, 1) ASC,
    am.track_no ASC NULLS LAST,
    hf."name" COLLATE "C" ASC,
    hf.id ASC
LIMIT
    $4 OFFSET $5;
