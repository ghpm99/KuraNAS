-- Cold files (bytes on the cold volume) accessed again since the cutoff: a real
-- user read (recent_file) earned them their way back to the hot tier. Most
-- recently accessed first. $1 = recent cutoff.
SELECT hf.id,
       hf.path,
       hf.physical_path,
       hf.size
FROM home_file hf
WHERE hf.type = 2
  AND hf.deleted_at IS NULL
  AND hf.physical_path IS NOT NULL
  AND EXISTS (
      SELECT 1 FROM recent_file rf
      WHERE rf.file_id = hf.id
        AND rf.accessed_at >= $1
  )
ORDER BY (
      SELECT MAX(rf.accessed_at) FROM recent_file rf
      WHERE rf.file_id = hf.id
  ) DESC;
