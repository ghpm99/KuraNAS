kuranas_fold(hf.name) LIKE kuranas_fold(@1)
    AND kuranas_fold(hf.name) LIKE ALL (kuranas_fold_terms(@2::text[]))
