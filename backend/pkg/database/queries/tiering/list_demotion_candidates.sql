-- Hot files (bytes still at the logical path) idle since the cutoff and large
-- enough to be worth moving. "Idle" means neither modified nor accessed since
-- the cutoff: updated_at is older than it AND no recent_file access row is newer
-- than it. Least-recently-modified first so a partial pass migrates the coldest
-- files. $1 = min size bytes, $2 = idle cutoff.
SELECT hf.id,
       hf.path,
       hf.size
FROM home_file hf
WHERE hf.type = 2
  AND hf.deleted_at IS NULL
  AND hf.physical_path IS NULL
  AND hf.size >= $1
  AND hf.updated_at < $2
  AND NOT EXISTS (
      SELECT 1 FROM recent_file rf
      WHERE rf.file_id = hf.id
        AND rf.accessed_at >= $2
  )
ORDER BY hf.updated_at ASC;
