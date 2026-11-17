SELECT DISTINCT
    hf.id
FROM
    audio_metadata_effective am
    JOIN home_file hf ON hf.id = am.file_id
WHERE
    hf.deleted_at IS NULL
    AND (
        hf.parent_path = $1
        OR hf.parent_path LIKE $2
    )
ORDER BY
    hf.id ASC;
