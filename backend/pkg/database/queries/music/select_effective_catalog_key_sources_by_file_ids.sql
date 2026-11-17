SELECT
    ame.id,
    COALESCE(ame.artist, ''),
    COALESCE(ame.album_artist, ''),
    COALESCE(ame.album, ''),
    COALESCE(ame.genre, '')
FROM
    audio_metadata_effective ame
WHERE
    ame.file_id = ANY ($1)
ORDER BY
    ame.id;
