SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.size,
    hf.created_at,
    hf.updated_at,
    vm.duration,
    vm.width,
    vm.height,
    vm.frame_rate,
    vm.codec_name,
    vm.aspect_ratio,
    vm.audio_channels,
    vm.audio_codec,
    vm.audio_sample_rate
FROM home_file hf
LEFT JOIN video_metadata vm ON hf.id = vm.file_id
WHERE hf.deleted_at IS NULL
  AND hf.format = ANY($1)
ORDER BY LOWER(hf.parent_path), LOWER(hf.name), hf.id;
