SELECT
    concat_ws(' ', NULLIF(im.make, ''), NULLIF(im.model, '')) AS camera,
    count(*)
FROM
    image_metadata im
    JOIN home_file hf ON hf.id = im.file_id
