INSERT INTO
    audio_metadata_override (
        file_id,
        title,
        artist,
        album,
        album_artist,
        genre,
        year,
        disc_number,
        track_number,
        updated_at
    )
SELECT DISTINCT
    am.file_id,
    $3::TEXT,
    $5::TEXT,
    $7::TEXT,
    $9::TEXT,
    $11::TEXT,
    $13::TEXT,
    $15::INTEGER,
    $17::INTEGER,
    now()
FROM
    audio_metadata am
    JOIN home_file hf ON hf.id = am.file_id
WHERE
    am.file_id = ANY ($1)
    AND hf.deleted_at IS NULL
ON CONFLICT (file_id) DO UPDATE
SET
    title = CASE WHEN $2::BOOLEAN THEN EXCLUDED.title ELSE audio_metadata_override.title END,
    artist = CASE WHEN $4::BOOLEAN THEN EXCLUDED.artist ELSE audio_metadata_override.artist END,
    album = CASE WHEN $6::BOOLEAN THEN EXCLUDED.album ELSE audio_metadata_override.album END,
    album_artist = CASE WHEN $8::BOOLEAN THEN EXCLUDED.album_artist ELSE audio_metadata_override.album_artist END,
    genre = CASE WHEN $10::BOOLEAN THEN EXCLUDED.genre ELSE audio_metadata_override.genre END,
    year = CASE WHEN $12::BOOLEAN THEN EXCLUDED.year ELSE audio_metadata_override.year END,
    disc_number = CASE WHEN $14::BOOLEAN THEN EXCLUDED.disc_number ELSE audio_metadata_override.disc_number END,
    track_number = CASE WHEN $16::BOOLEAN THEN EXCLUDED.track_number ELSE audio_metadata_override.track_number END,
    updated_at = now();
