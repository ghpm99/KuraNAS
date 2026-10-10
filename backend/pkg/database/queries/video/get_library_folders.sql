WITH scope AS (
    SELECT prefix, is_whole_root, label
    FROM unnest($2::text[], $3::boolean[], $4::text[]) AS scope(prefix, is_whole_root, label)
),
video_in_scope AS (
    SELECT
        hf.id,
        hf.name,
        scope.prefix,
        scope.is_whole_root,
        scope.label,
        substr(hf.path, length(scope.prefix) + 1) AS relative_path
    FROM
        home_file hf
        JOIN scope ON hf.path ^@ scope.prefix
    WHERE
        hf.deleted_at IS NULL
        AND hf.format = ANY($1)
),
video_in_child_folder AS (
    SELECT
        id,
        name,
        CASE WHEN is_whole_root THEN left(prefix, -1) ELSE prefix || split_part(relative_path, $5, 1) END AS folder_path,
        CASE WHEN is_whole_root THEN label ELSE split_part(relative_path, $5, 1) END AS folder_name
    FROM
        video_in_scope
    WHERE
        is_whole_root OR strpos(relative_path, $5) > 0
)
SELECT
    folder_path,
    folder_name,
    count(*) AS video_count,
    (array_agg(id ORDER BY lower(name), id))[1] AS cover_file_id
FROM
    video_in_child_folder
GROUP BY
    folder_path, folder_name
ORDER BY
    lower(folder_name), folder_path
LIMIT $6 OFFSET $7
