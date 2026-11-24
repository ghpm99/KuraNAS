WITH episodes AS (
    SELECT
        vpi.id AS item_id,
        vpi.video_id,
        vpi.order_index,
        COALESCE(mo.season, vpi.season) AS season,
        COALESCE(mo.episode, vpi.episode) AS episode,
        COALESCE(wp.position_seconds, 0) AS position_seconds,
        COALESCE(wp.completed, FALSE) AS is_completed,
        wp.updated_at AS progress_updated_at
    FROM video_playlist_item vpi
    JOIN home_file hf ON hf.id = vpi.video_id
    LEFT JOIN video_manual_override mo ON mo.file_id = vpi.video_id
    LEFT JOIN video_watch_progress wp ON wp.video_id = vpi.video_id AND wp.client_id = $2
    WHERE vpi.playlist_id = $1
      AND hf.deleted_at IS NULL
),
candidates AS (
    SELECT *, 'continue' AS mode, 1 AS priority, position_seconds AS resume_position
    FROM episodes
    WHERE NOT is_completed AND position_seconds > 0
    UNION ALL
    SELECT *, 'start' AS mode, 2 AS priority, 0 AS resume_position
    FROM episodes
    WHERE NOT is_completed
    UNION ALL
    SELECT *, 'rewatch' AS mode, 3 AS priority, 0 AS resume_position
    FROM episodes
)
SELECT video_id, season, episode, resume_position, mode
FROM candidates
ORDER BY
    priority,
    CASE WHEN priority = 1 THEN progress_updated_at END DESC,
    order_index,
    item_id
LIMIT 1;
