SELECT id, name, file_name, file_path, media_type, mime_type, size, created_at
FROM captures
WHERE
    ($1::boolean OR name ILIKE '%' || $2::text || '%')
    AND ($3::boolean OR media_type = $4::text)
ORDER BY created_at DESC
LIMIT $5
OFFSET $6;
