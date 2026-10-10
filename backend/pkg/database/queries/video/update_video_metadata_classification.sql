UPDATE video_metadata
SET
    classification = $2,
    classification_version = $3
WHERE
    id = $1;
