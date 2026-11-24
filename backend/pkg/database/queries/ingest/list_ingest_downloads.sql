SELECT
    download.id,
    download.job_id,
    download.url,
    download.title,
    download.preset,
    download.target_root_id,
    download.target_root_name,
    download.subfolder,
    download.state,
    download.progress,
    download.speed_bytes_per_second,
    download.eta_seconds,
    downloaded_file.id,
    download.error,
    download.created_at,
    download.ended_at
FROM ingest_download AS download
LEFT JOIN LATERAL (
    SELECT file.id
    FROM home_file AS file
    WHERE download.final_path <> ''
      AND file.path = download.final_path
      AND file.deleted_at IS NULL
    ORDER BY file.id
    LIMIT 1
) AS downloaded_file ON TRUE
WHERE $1::text = ''
   OR ($1::text = 'active') = (download.state IN ('queued', 'running'))
ORDER BY download.created_at DESC, download.id DESC
LIMIT $2 OFFSET $3;
