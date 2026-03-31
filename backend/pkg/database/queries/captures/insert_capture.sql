INSERT INTO captures (name, file_name, file_path, media_type, mime_type, size, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;
