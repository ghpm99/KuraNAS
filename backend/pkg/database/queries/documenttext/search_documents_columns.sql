SELECT
    hf.id,
    hf.name,
    hf.path,
    hf.parent_path,
    hf.format,
    hf.size,
    hf.updated_at,
    hf.starred,
    kuranas_fold(hf.name) LIKE ALL (kuranas_fold_terms(@1::text[])) AS is_name_match
