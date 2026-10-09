UPDATE playlist_track
SET position = ranked.new_position
FROM (
    SELECT id, ROW_NUMBER() OVER (ORDER BY position, id) AS new_position
    FROM playlist_track
    WHERE playlist_id = $1
) AS ranked
WHERE playlist_track.id = ranked.id
  AND playlist_track.position <> ranked.new_position;
