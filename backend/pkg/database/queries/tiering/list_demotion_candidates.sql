-- Hot files (bytes still at the logical path) untouched since the cutoff and
-- large enough to be worth moving. Least-recently-used first so a partial pass
-- migrates the coldest files. $1 = min size bytes, $2 = idle cutoff, $3 = limit.
SELECT hf.id,
       hf.path,
       hf.size
FROM home_file hf
WHERE hf.type = 2
  AND hf.deleted_at IS NULL
  AND hf.physical_path IS NULL
  AND hf.size >= $1
  AND COALESCE(hf.last_interaction, hf.created_at) < $2
ORDER BY COALESCE(hf.last_interaction, hf.created_at) ASC
LIMIT $3;
