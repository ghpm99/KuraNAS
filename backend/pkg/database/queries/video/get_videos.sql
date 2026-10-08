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
    COALESCE(vm.id, 0),
    COALESCE(vm.file_id, 0),
    COALESCE(vm.path, ''),
    COALESCE(vm.format_name, ''),
    COALESCE(vm.size, ''),
    COALESCE(vm.duration, ''),
    COALESCE(vm.width, 0),
    COALESCE(vm.height, 0),
    COALESCE(vm.frame_rate, 0),
    COALESCE(vm.nb_frames, 0),
    COALESCE(vm.bit_rate, ''),
    COALESCE(vm.codec_name, ''),
    COALESCE(vm.codec_long_name, ''),
    COALESCE(vm.pix_fmt, ''),
    COALESCE(vm.level, 0),
    COALESCE(vm.profile, ''),
    COALESCE(vm.aspect_ratio, ''),
    COALESCE(vm.audio_codec, ''),
    COALESCE(vm.audio_channels, 0),
    COALESCE(vm.audio_sample_rate, ''),
    COALESCE(vm.audio_bit_rate, ''),
    COALESCE(vm.created_at, TIMESTAMPTZ '0001-01-01 00:00:00+00')
FROM
    home_file hf
    LEFT JOIN video_metadata vm ON hf.id = vm.file_id
WHERE
    hf.format = ANY($1)
    AND hf.deleted_at IS NULL
ORDER BY
    hf.TYPE,
    hf.NAME,
    hf.id DESC
LIMIT
    $2
OFFSET
    $3;
