SELECT
    id,
    job_id,
    file_name,
    size,
    state,
    total_entries,
    processed_entries,
    imported,
    already_existing,
    ignored,
    failed_entries,
    error,
    created_at,
    ended_at
FROM takeout_import
WHERE id = $1;
