SELECT
    duration,
    width,
    height,
    codec_name,
    frame_rate,
    bit_rate,
    audio_codec,
    format_name
FROM
    video_metadata
WHERE
    file_id = $1
ORDER BY
    id DESC
LIMIT
    1;
