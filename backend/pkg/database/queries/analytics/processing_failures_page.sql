WITH staged_steps AS (
	SELECT
		step.id AS step_id,
		step.status,
		step.last_error,
		COALESCE(step.ended_at, step.created_at) AS finished_at,
		CASE step.type
			WHEN 'checksum' THEN 'checksum'
			WHEN 'metadata' THEN 'metadata'
			WHEN 'thumbnail' THEN 'thumbnail'
			WHEN 'document_text_index' THEN 'document_text'
			WHEN 'image_classify_batch' THEN 'ai'
			WHEN 'ai_playlist_cluster' THEN 'ai'
		END AS stage,
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
	WHERE step.status IN ('failed', 'completed', 'skipped')
),
latest_per_file_stage AS (
	SELECT DISTINCT ON (file_id, stage)
		step_id,
		status,
		last_error,
		finished_at,
		stage,
		file_id
	FROM staged_steps
	WHERE stage IS NOT NULL
		AND file_id IS NOT NULL
	ORDER BY file_id, stage, finished_at DESC, step_id DESC
)
SELECT
	latest.file_id,
	file.name,
	latest.stage,
	COALESCE(latest.last_error, '') AS last_error,
	latest.finished_at AS failed_at
FROM latest_per_file_stage latest
JOIN home_file file ON file.id = latest.file_id
WHERE latest.status = 'failed'
	AND file.deleted_at IS NULL
	AND ($1::text IS NULL OR latest.stage = $1::text)
ORDER BY latest.finished_at DESC, latest.step_id DESC
LIMIT $2 OFFSET $3;
