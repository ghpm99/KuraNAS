DELETE FROM audio_metadata_override
WHERE
    file_id = ANY ($1)
    AND title IS NULL
    AND artist IS NULL
    AND album IS NULL
    AND album_artist IS NULL
    AND genre IS NULL
    AND year IS NULL
    AND disc_number IS NULL
    AND track_number IS NULL;
