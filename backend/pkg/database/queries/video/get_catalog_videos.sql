SELECT
    id,
    name,
    path,
    parent_path,
    format,
    size,
    created_at,
    updated_at,
    COALESCE(starred, FALSE),
    COALESCE((
        SELECT CASE vmo.classification WHEN 'other' THEN 'program' ELSE vmo.classification END
        FROM video_manual_override vmo
        WHERE vmo.file_id = home_file.id
          AND vmo.classification IS NOT NULL
    ), (
        SELECT vm.classification
        FROM video_metadata vm
        WHERE vm.file_id = home_file.id
          AND vm.classification IS NOT NULL
        ORDER BY vm.id DESC
        LIMIT 1
    ), '')
FROM home_file
WHERE deleted_at IS NULL
  AND format = ANY($1)
ORDER BY updated_at DESC, id DESC
LIMIT $2;
