SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.size,
    COALESCE(im.width, 0),
    COALESCE(im.height, 0),
    im.taken_at,
    COALESCE(im.classification_category, 'other'),
    hf.starred,
    hf.physical_path IS NOT NULL,
    hf.updated_at
FROM
    image_metadata im
    JOIN home_file hf ON hf.id = im.file_id
