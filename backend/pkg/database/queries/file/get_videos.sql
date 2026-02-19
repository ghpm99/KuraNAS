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
    vm.id,
    vm.file_id,
    vm.path,
    vm.format_name,
    vm.size,
    vm.duration,
    vm.width,
    vm.height,
    vm.frame_rate,
    vm.nb_frames,
    vm.bit_rate,
    vm.codec_name,
    vm.codec_long_name,
    vm.pix_fmt,
    vm.level,
    vm.profile,
    vm.aspect_ratio,
    vm.audio_codec,
    vm.audio_channels,
    vm.audio_sample_rate,
    vm.audio_bit_rate,
    vm.created_at
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
