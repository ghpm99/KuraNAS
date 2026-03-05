INSERT INTO playlist_track (playlist_id, file_id, position)
VALUES ($1, $2, COALESCE((SELECT MAX(position) + 1 FROM playlist_track WHERE playlist_id = $1), 1))
ON CONFLICT (playlist_id, file_id) DO NOTHING
RETURNING id, position, added_at;
