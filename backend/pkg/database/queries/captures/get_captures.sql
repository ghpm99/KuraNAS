SELECT id, name, file_name, file_path, media_type, mime_type, size, episode_key, created_at,
    file_id, status, title, episode_title, season, episode, description, release_year,
    genres, cast_members, directors, studio, content_rating, platform, source_url, thumbnail_url, content_type, raw_metadata
FROM captures
WHERE
    ($1::boolean OR name ILIKE '%' || $2::text || '%')
    AND ($3::boolean OR media_type = $4::text)
ORDER BY created_at DESC
LIMIT $5
OFFSET $6;
