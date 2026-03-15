SELECT
	COALESCE(
		(
			SELECT COUNT(*)
			FROM (
				SELECT am.file_id
				FROM audio_metadata am
				INNER JOIN home_file hf ON hf.id = am.file_id
				WHERE hf.type = 2
					AND hf.deleted_at IS NULL
				UNION
				SELECT vm.file_id
				FROM video_metadata vm
				INNER JOIN home_file hf ON hf.id = vm.file_id
				WHERE hf.type = 2
					AND hf.deleted_at IS NULL
				UNION
				SELECT im.file_id
				FROM image_metadata im
				INNER JOIN home_file hf ON hf.id = im.file_id
				WHERE hf.type = 2
					AND hf.deleted_at IS NULL
			) AS categorized_media
		),
		0
	) AS categorized_media,
	COALESCE(
		(
			SELECT COUNT(DISTINCT am.file_id)
			FROM audio_metadata am
			INNER JOIN home_file hf ON hf.id = am.file_id
			WHERE hf.type = 2
				AND hf.deleted_at IS NULL
		),
		0
	) AS audio_with_metadata,
	COALESCE(
		(
			SELECT COUNT(DISTINCT vm.file_id)
			FROM video_metadata vm
			INNER JOIN home_file hf ON hf.id = vm.file_id
			WHERE hf.type = 2
				AND hf.deleted_at IS NULL
		),
		0
	) AS video_with_metadata,
	COALESCE(
		(
			SELECT COUNT(DISTINCT im.file_id)
			FROM image_metadata im
			INNER JOIN home_file hf ON hf.id = im.file_id
			WHERE hf.type = 2
				AND hf.deleted_at IS NULL
		),
		0
	) AS image_with_metadata,
	COALESCE(
		(
			SELECT COUNT(DISTINCT im.file_id)
			FROM image_metadata im
			INNER JOIN home_file hf ON hf.id = im.file_id
			WHERE hf.type = 2
				AND hf.deleted_at IS NULL
				AND COALESCE(im.classification_category, '') <> ''
		),
		0
	) AS image_classified;
