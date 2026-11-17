SELECT
    COUNT(*)
FROM
    audio_metadata am
    JOIN home_file hf ON hf.id = am.file_id
WHERE
    am.catalog_genre_keys @> ARRAY[$1::TEXT]
    AND hf.format = ANY ($2)
    AND hf.deleted_at IS NULL;
