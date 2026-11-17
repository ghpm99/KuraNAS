SELECT
    ame.file_id,
    COALESCE(ame.genre, '')
FROM
    audio_metadata_effective ame
    JOIN home_file hf ON hf.id = ame.file_id
WHERE
    ame.catalog_genre_keys @> ARRAY[$1::TEXT]
    AND hf.format = ANY ($2)
    AND hf.deleted_at IS NULL
ORDER BY
    ame.file_id ASC;
