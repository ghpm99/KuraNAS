INSERT INTO video_manual_override (file_id, classification)
VALUES ($1, $2)
ON CONFLICT (file_id) DO UPDATE
SET classification = EXCLUDED.classification,
    updated_at = now();
