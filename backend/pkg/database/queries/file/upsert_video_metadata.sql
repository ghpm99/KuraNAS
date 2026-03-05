INSERT INTO
    video_metadata (
        file_id,
        PATH,
        format_name,
        size,
        duration,
        WIDTH,
        HEIGHT,
        frame_rate,
        nb_frames,
        bit_rate,
        codec_name,
        codec_long_name,
        pix_fmt,
        LEVEL,
        profile,
        aspect_ratio,
        audio_codec,
        audio_channels,
        audio_sample_rate,
        audio_bit_rate,
        created_at
    )
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21) ON CONFLICT (file_id, PATH)
DO
UPDATE
SET
    format_name = EXCLUDED.format_name,
    size = EXCLUDED.size,
    duration = EXCLUDED.duration,
    WIDTH = EXCLUDED.width,
    HEIGHT = EXCLUDED.height,
    frame_rate = EXCLUDED.frame_rate,
    nb_frames = EXCLUDED.nb_frames,
    bit_rate = EXCLUDED.bit_rate,
    codec_name = EXCLUDED.codec_name,
    codec_long_name = EXCLUDED.codec_long_name,
    pix_fmt = EXCLUDED.pix_fmt,
    LEVEL = EXCLUDED.level,
    profile = EXCLUDED.profile,
    aspect_ratio = EXCLUDED.aspect_ratio,
    audio_codec = EXCLUDED.audio_codec,
    audio_channels = EXCLUDED.audio_channels,
    audio_sample_rate = EXCLUDED.audio_sample_rate,
    audio_bit_rate = EXCLUDED.audio_bit_rate
RETURNING
    id,
    created_at;