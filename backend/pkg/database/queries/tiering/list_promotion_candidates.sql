-- Cold files (bytes on the cold volume) that were used again since the cutoff:
-- they earned their way back to the hot tier. $1 = recent cutoff, $2 = limit.
SELECT hf.id,
       hf.path,
       hf.physical_path,
       hf.size
FROM home_file hf
WHERE hf.type = 2
  AND hf.deleted_at IS NULL
  AND hf.physical_path IS NOT NULL
  AND hf.last_interaction IS NOT NULL
  AND hf.last_interaction >= $1
ORDER BY hf.last_interaction DESC
LIMIT $2;
