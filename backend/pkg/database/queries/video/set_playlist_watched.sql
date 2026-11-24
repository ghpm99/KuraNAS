INSERT INTO video_watch_progress (
    client_id,
    video_id,
    position_seconds,
    duration_seconds,
    completed,
    updated_at
)
SELECT
    $1,
    target.video_id,
    CASE WHEN $4::boolean THEN COALESCE(wp.duration_seconds, 0) ELSE 0 END,
    COALESCE(wp.duration_seconds, 0),
    $4::boolean,
    NOW()
FROM (
    SELECT DISTINCT vpi.video_id
    FROM video_playlist_item vpi
    JOIN home_file hf ON hf.id = vpi.video_id
    LEFT JOIN video_manual_override mo ON mo.file_id = vpi.video_id
    WHERE vpi.playlist_id = $2
      AND hf.deleted_at IS NULL
      AND ($3::int IS NULL OR COALESCE(mo.season, vpi.season) = $3::int)
) target
LEFT JOIN video_watch_progress wp ON wp.client_id = $1 AND wp.video_id = target.video_id
ON CONFLICT (client_id, video_id)
DO UPDATE SET
    position_seconds = EXCLUDED.position_seconds,
    duration_seconds = EXCLUDED.duration_seconds,
    completed = EXCLUDED.completed,
    updated_at = NOW();
