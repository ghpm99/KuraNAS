SELECT DISTINCT vp.id
FROM video_playlist vp
JOIN video_playlist_item vpi ON vpi.playlist_id = vp.id
WHERE vpi.video_id = $1
  AND vpi.source_kind = 'auto'
  AND vp.is_auto = TRUE
  AND (vp.type = 'series' OR vp.classification IN ('series', 'anime'))
ORDER BY vp.id;
