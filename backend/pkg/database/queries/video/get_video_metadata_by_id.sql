SELECT
    id,
    file_id,
    PATH,
    COALESCE(format_name, ''),
    COALESCE(size, ''),
    COALESCE(duration, ''),
    COALESCE(WIDTH, 0),
    COALESCE(HEIGHT, 0),
    COALESCE(frame_rate, 0),
    COALESCE(nb_frames, 0),
    COALESCE(bit_rate, ''),
    COALESCE(codec_name, ''),
    COALESCE(codec_long_name, ''),
    COALESCE(pix_fmt, ''),
    COALESCE(LEVEL, 0),
    COALESCE(profile, ''),
    COALESCE(aspect_ratio, ''),
    COALESCE(audio_codec, ''),
    COALESCE(audio_channels, 0),
    COALESCE(audio_sample_rate, ''),
    COALESCE(audio_bit_rate, ''),
    created_at,
    COALESCE(classification, ''),
    classification_version
FROM
    video_metadata
WHERE
    id = $1;
