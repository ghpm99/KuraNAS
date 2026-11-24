FROM
    home_file hf
    LEFT JOIN document_text dt ON dt.file_id = hf.id
WHERE
    hf.type = 2
    AND hf.deleted_at IS NULL
    AND hf.format = ANY(@1)
