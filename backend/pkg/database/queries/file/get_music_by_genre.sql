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
    am.id,
    am.file_id,
    am.PATH,
    am.mime,
    am.LENGTH,
    am.bitrate,
    am.sample_rate,
    am.channels,
    am.bitrate_mode,
    am.encoder_info,
    am.bit_depth,
    am.title,
    am.artist,
    am.album,
    am.album_artist,
    am.track_number,
    am.genre,
    am.composer,
    am.YEAR,
    am.recording_date,
    am.encoder,
    am.publisher,
    am.original_release_date,
    am.original_artist,
    am.lyricist,
    am.lyrics,
    am.created_at
FROM
    home_file hf
    INNER JOIN audio_metadata am ON hf.id = am.file_id
WHERE
    hf.format = ANY ($1)
    AND hf.deleted_at IS NULL
    AND am.genre = $2
ORDER BY
    am.artist,
    am.album,
    am.track_number,
    hf.NAME
LIMIT
    $3
OFFSET
    $4;
