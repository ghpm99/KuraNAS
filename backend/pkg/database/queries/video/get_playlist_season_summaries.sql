SELECT
    COALESCE(mo.season, vpi.season) AS season,
    COUNT(*) AS episode_count,
    COUNT(*) FILTER (WHERE COALESCE(wp.completed, FALSE)) AS watched_count
FROM video_playlist_item vpi
JOIN home_file hf ON hf.id = vpi.video_id
LEFT JOIN video_manual_override mo ON mo.file_id = vpi.video_id
LEFT JOIN video_watch_progress wp ON wp.video_id = vpi.video_id AND wp.client_id = $2
WHERE vpi.playlist_id = $1
  AND hf.deleted_at IS NULL
  AND COALESCE(mo.season, vpi.season) IS NOT NULL
GROUP BY COALESCE(mo.season, vpi.season)
ORDER BY season;
