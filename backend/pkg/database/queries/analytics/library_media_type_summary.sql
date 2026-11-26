WITH metadata_steps AS (
	SELECT
		step.id AS step_id,
		step.status,
		COALESCE(step.ended_at, step.created_at) AS finished_at,
		COALESCE(
			(step.payload::jsonb ->> 'file_id')::int,
			(parent_job.scope::jsonb ->> 'file_id')::int,
			(
				SELECT file.id
				FROM home_file file
				WHERE file.path = COALESCE(step.payload::jsonb ->> 'path', parent_job.scope::jsonb ->> 'path')
				LIMIT 1
			)
		) AS file_id
	FROM worker_step step
	JOIN worker_job parent_job ON parent_job.id = step.job_id
	WHERE step.type = 'metadata'
),
latest_terminal_metadata_step AS (
	SELECT DISTINCT ON (file_id)
		file_id,
		status
	FROM metadata_steps
	WHERE file_id IS NOT NULL
		AND status IN ('failed', 'completed', 'skipped')
	ORDER BY file_id, finished_at DESC, step_id DESC
),
files_with_active_metadata_step AS (
	SELECT DISTINCT file_id
	FROM metadata_steps
	WHERE file_id IS NOT NULL
		AND status IN ('queued', 'running')
),
media_files AS (
	SELECT
		file.id,
		CASE
			WHEN lower(COALESCE(file.format, '')) = ANY($1::text[]) THEN 'video'
			WHEN lower(COALESCE(file.format, '')) = ANY($2::text[]) THEN 'image'
			WHEN lower(COALESCE(file.format, '')) = ANY($3::text[]) THEN 'audio'
		END AS media_type
	FROM home_file file
	WHERE file.type = 2
		AND file.deleted_at IS NULL
),
classified_media_files AS (
	SELECT
		media.media_type,
		(
			(media.media_type = 'audio' AND EXISTS (SELECT 1 FROM audio_metadata am WHERE am.file_id = media.id))
			OR (media.media_type = 'video' AND EXISTS (SELECT 1 FROM video_metadata vm WHERE vm.file_id = media.id))
			OR (media.media_type = 'image' AND EXISTS (SELECT 1 FROM image_metadata im WHERE im.file_id = media.id))
		) AS has_metadata,
		(
			latest.status = 'failed'
			AND active.file_id IS NULL
		) AS has_failed
	FROM media_files media
	LEFT JOIN latest_terminal_metadata_step latest ON latest.file_id = media.id
	LEFT JOIN files_with_active_metadata_step active ON active.file_id = media.id
	WHERE media.media_type IS NOT NULL
)
SELECT
	media_type_catalog.media_type,
	COUNT(classified.media_type) AS total,
	COUNT(classified.media_type) FILTER (WHERE classified.has_metadata) AS with_metadata,
	COUNT(classified.media_type) FILTER (WHERE NOT classified.has_metadata AND NOT COALESCE(classified.has_failed, FALSE)) AS pending,
	COUNT(classified.media_type) FILTER (WHERE NOT classified.has_metadata AND COALESCE(classified.has_failed, FALSE)) AS failed
FROM (VALUES ('audio'), ('video'), ('image')) AS media_type_catalog(media_type)
LEFT JOIN classified_media_files classified ON classified.media_type = media_type_catalog.media_type
GROUP BY media_type_catalog.media_type
ORDER BY CASE media_type_catalog.media_type WHEN 'audio' THEN 1 WHEN 'video' THEN 2 ELSE 3 END;
