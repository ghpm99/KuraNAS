hf.id IN (
        SELECT name_match.id
        FROM home_file name_match
        WHERE name_match.deleted_at IS NULL
            AND lower(name_match.name) LIKE lower(@1)
        UNION
        SELECT content_match.file_id
        FROM image_metadata content_match
        WHERE content_match.ai_search_text LIKE @2
    )
