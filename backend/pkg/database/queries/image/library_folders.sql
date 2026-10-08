WITH scope AS (
    SELECT prefix, is_whole_root, label
    FROM unnest($2::text[], $3::boolean[], $4::text[]) AS scope(prefix, is_whole_root, label)
),
image_in_scope AS (
    SELECT
        hf.id,
        im.taken_at,
        scope.prefix,
        scope.is_whole_root,
        scope.label,
        substr(hf.path, length(scope.prefix) + 1) AS relative_path
    FROM
        image_metadata im
        JOIN home_file hf ON hf.id = im.file_id
        JOIN scope ON hf.path ^@ scope.prefix
    WHERE
        hf.deleted_at IS NULL
        AND hf.format = ANY($1)
        AND NOT EXISTS (
            SELECT 1 FROM image_metadata newer
            WHERE newer.file_id = im.file_id AND newer.id > im.id
        )
),
image_in_child_folder AS (
    SELECT
        id,
        taken_at,
        CASE WHEN is_whole_root THEN left(prefix, -1) ELSE prefix || split_part(relative_path, $5, 1) END AS folder_path,
        CASE WHEN is_whole_root THEN label ELSE split_part(relative_path, $5, 1) END AS folder_name
    FROM
        image_in_scope
    WHERE
        is_whole_root OR strpos(relative_path, $5) > 0
)
SELECT
    folder_path,
    folder_name,
    count(*) AS image_count,
    (array_agg(id ORDER BY taken_at DESC NULLS LAST, id DESC))[1] AS cover_file_id
FROM
    image_in_child_folder
GROUP BY
    folder_path, folder_name
ORDER BY
    lower(folder_name), folder_path
LIMIT $6 OFFSET $7
