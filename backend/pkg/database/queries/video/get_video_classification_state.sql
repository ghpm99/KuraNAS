SELECT
    vmo.classification,
    (
        SELECT vm.classification
        FROM video_metadata vm
        WHERE vm.file_id = hf.id
          AND vm.classification IS NOT NULL
        ORDER BY vm.id DESC
        LIMIT 1
    )
FROM home_file hf
LEFT JOIN video_manual_override vmo ON vmo.file_id = hf.id
WHERE hf.id = $1;
