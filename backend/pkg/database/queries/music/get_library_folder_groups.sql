WITH folder_groups AS (
    SELECT
        COALESCE(NULLIF(TRIM(hf.parent_path), ''), '/') AS folder,
        COUNT(*) AS track_count
    FROM
        home_file hf
    WHERE
        hf.format = ANY ($1)
        AND hf.deleted_at IS NULL
    GROUP BY
        1
)
SELECT
    folder,
    track_count
FROM
    folder_groups
ORDER BY
    track_count DESC,
    folder COLLATE "C" ASC
LIMIT
    $2 OFFSET $3;
