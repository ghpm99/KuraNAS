WITH folder_groups AS (
    SELECT
        COALESCE(NULLIF(TRIM(hf.parent_path), ''), '/') AS folder,
        COUNT(*) AS track_count,
        MAX(hf.created_at) AS latest_added_at
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
    (
        CASE
            WHEN $2::TEXT = 'tracks' THEN track_count
            WHEN $2::TEXT = 'recent' THEN EXTRACT(EPOCH FROM latest_added_at)::BIGINT
        END
    ) * (CASE WHEN $3::BOOLEAN THEN -1 ELSE 1 END) ASC NULLS LAST,
    (CASE WHEN $2::TEXT = 'name' AND NOT $3::BOOLEAN THEN folder END) COLLATE "C" ASC,
    (CASE WHEN $2::TEXT = 'name' AND $3::BOOLEAN THEN folder END) COLLATE "C" DESC,
    track_count DESC,
    folder COLLATE "C" ASC
LIMIT
    $4 OFFSET $5;
