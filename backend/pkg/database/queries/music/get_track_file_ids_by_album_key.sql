SELECT DISTINCT
    hf.id
FROM
    audio_metadata_effective am
    JOIN home_file hf ON hf.id = am.file_id
WHERE
    am.catalog_album_key = $1
    AND hf.deleted_at IS NULL
ORDER BY
    hf.id ASC;
