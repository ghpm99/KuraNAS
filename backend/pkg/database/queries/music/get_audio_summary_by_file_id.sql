SELECT
    title,
    artist,
    album,
    genre,
    year,
    track_number,
    length,
    bitrate,
    sample_rate,
    channels
FROM
    audio_metadata_effective
WHERE
    file_id = $1
ORDER BY
    id DESC
LIMIT
    1;
