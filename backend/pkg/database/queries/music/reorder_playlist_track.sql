WITH other_tracks AS (
    SELECT file_id, ROW_NUMBER() OVER (ORDER BY position, id) AS rank
    FROM playlist_track
    WHERE playlist_id = $2
      AND file_id <> $3
),
target AS (
    SELECT LEAST(GREATEST($1::int, 1), (SELECT COUNT(*) FROM other_tracks) + 1) AS position
),
renumbered AS (
    SELECT other_tracks.file_id,
           CASE WHEN other_tracks.rank >= target.position THEN other_tracks.rank + 1 ELSE other_tracks.rank END AS new_position
    FROM other_tracks, target
    UNION ALL
    SELECT $3::int, target.position
    FROM target
)
UPDATE playlist_track
SET position = renumbered.new_position
FROM renumbered
WHERE playlist_track.playlist_id = $2
  AND playlist_track.file_id = renumbered.file_id
  AND EXISTS (SELECT 1 FROM playlist_track WHERE playlist_id = $2 AND file_id = $3);
