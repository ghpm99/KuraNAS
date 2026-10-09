hf.id IN (
        SELECT name_match.id
        FROM home_file name_match
        WHERE name_match.deleted_at IS NULL
            AND kuranas_fold(name_match.name) LIKE kuranas_fold(@1)
        UNION
        SELECT content_match.file_id
        FROM image_metadata content_match
        WHERE kuranas_fold(content_match.ai_search_text) LIKE kuranas_fold(@2)
    )
