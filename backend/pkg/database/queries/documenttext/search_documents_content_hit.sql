SELECT
    matched.id,
    matched.name,
    matched.path,
    matched.parent_path,
    matched.format,
    matched.size,
    matched.updated_at,
    COALESCE(content_hit.has_every_term, FALSE) AS has_content_match,
    CASE
        WHEN COALESCE(content_hit.has_every_term, FALSE)
            THEN substr(dt.extracted_text, GREATEST(content_hit.first_position - @2::int, 1), @3::int)
        ELSE ''
    END AS content_window,
    CASE
        WHEN COALESCE(content_hit.has_every_term, FALSE) AND position(E'\f' IN dt.extracted_text) > 0
            THEN 1
                + length(left(dt.extracted_text, content_hit.first_position - 1))
                - length(replace(left(dt.extracted_text, content_hit.first_position - 1), E'\f', ''))
    END AS page
FROM
    matched
    LEFT JOIN document_text dt ON dt.file_id = matched.id AND dt.error IS NULL
    LEFT JOIN LATERAL (
        SELECT
            MIN(position(kuranas_fold(term) IN folded_content.folded_text)) AS first_position,
            BOOL_AND(position(kuranas_fold(term) IN folded_content.folded_text) > 0) AS has_every_term
        FROM
            unnest(@1::text[]) AS term,
            (SELECT kuranas_fold(dt.extracted_text) AS folded_text) AS folded_content
    ) content_hit ON TRUE
ORDER BY
    matched.is_name_match DESC,
    matched.updated_at DESC,
    matched.id ASC
