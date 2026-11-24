(
    kuranas_fold(hf.name) LIKE ALL (kuranas_fold_terms(@2::text[]))
    OR (
        dt.error IS NULL
        AND kuranas_fold(dt.extracted_text) LIKE kuranas_fold(@1)
        AND kuranas_fold(dt.extracted_text) LIKE ALL (kuranas_fold_terms(@2::text[]))
    )
)
