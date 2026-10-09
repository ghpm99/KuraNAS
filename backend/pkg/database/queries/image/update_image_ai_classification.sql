UPDATE
    image_metadata
SET
    classification_category = $2,
    classification_confidence = $3,
    classification_suggested_name = $4,
    ai_caption = $5,
    ai_tags = $6,
    ai_ocr_text = $7,
    ai_search_text = $8,
    ai_classified_at = now()
WHERE
    file_id = $1;
