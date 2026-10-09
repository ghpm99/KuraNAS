SELECT
    hf.format,
    count(*)
FROM
    image_metadata im
    JOIN home_file hf ON hf.id = im.file_id
