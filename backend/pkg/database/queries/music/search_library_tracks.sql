SELECT
    hf.id,
    hf."name",
    hf."path",
    hf.parent_path,
    hf.format,
    hf."size",
    hf.updated_at,
    hf.created_at,
    hf.last_interaction,
    hf.last_backup,
    hf."type",
    hf.checksum,
    hf.deleted_at,
    hf.starred,
    COALESCE(am.id, 0),
    COALESCE(am.file_id, 0),
    COALESCE(am.PATH, ''),
    COALESCE(am.mime, ''),
    COALESCE(am.LENGTH, 0),
    COALESCE(am.bitrate, 0),
    COALESCE(am.sample_rate, 0),
    COALESCE(am.channels, 0),
    COALESCE(am.bitrate_mode, 0),
    COALESCE(am.encoder_info, ''),
    COALESCE(am.bit_depth, 0),
    COALESCE(am.title, ''),
    COALESCE(am.artist, ''),
    COALESCE(am.album, ''),
    COALESCE(am.album_artist, ''),
    COALESCE(am.track_number, ''),
    COALESCE(am.genre, ''),
    COALESCE(am.composer, ''),
    COALESCE(am.YEAR, ''),
    COALESCE(am.recording_date, ''),
    COALESCE(am.encoder, ''),
    COALESCE(am.publisher, ''),
    COALESCE(am.original_release_date, ''),
    COALESCE(am.original_artist, ''),
    COALESCE(am.lyricist, ''),
    COALESCE(am.lyrics, ''),
    COALESCE(am.created_at, TIMESTAMPTZ '0001-01-01 00:00:00+00')
FROM
    home_file hf
    LEFT JOIN audio_metadata am ON hf.id = am.file_id
WHERE
    hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
    AND hf.id IN (
        SELECT title_match.file_id
        FROM audio_metadata title_match
        WHERE lower(title_match.title) LIKE $4
        UNION
        SELECT artist_match.file_id
        FROM audio_metadata artist_match
        WHERE lower(artist_match.artist) LIKE $4
        UNION
        SELECT album_match.file_id
        FROM audio_metadata album_match
        WHERE lower(album_match.album) LIKE $4
        UNION
        SELECT name_match.id
        FROM home_file name_match
        WHERE name_match.deleted_at IS NULL
            AND lower(name_match.name) LIKE $4
    )
    AND lower(concat_ws(' ', am.title, am.artist, am.album, hf.name)) LIKE ALL ($5::text[])
ORDER BY
    hf.starred DESC,
    hf.NAME,
    hf.id DESC
LIMIT
    $2
OFFSET
    $3;
