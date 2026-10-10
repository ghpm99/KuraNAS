SELECT
    vm.id,
    hf.name,
    hf.path,
    hf.parent_path,
    COALESCE(vm.duration, ''),
    COALESCE(vm.height, 0)
FROM
    video_metadata vm
    JOIN home_file hf ON hf.id = vm.file_id
WHERE
    hf.deleted_at IS NULL
    AND vm.id > $1
    AND (
        vm.classification IS NULL
        OR vm.classification_version < $2
    )
ORDER BY
    vm.id
LIMIT
    $3;
