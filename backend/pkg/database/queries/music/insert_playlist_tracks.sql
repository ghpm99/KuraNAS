INSERT INTO playlist_track (playlist_id, file_id, position)
SELECT $1, entry.file_id, entry.position
FROM unnest($2::int[]) WITH ORDINALITY AS entry(file_id, position)
ON CONFLICT (playlist_id, file_id) DO NOTHING;
