-- One active file with its logical and (when cold) physical location. $1 = file id.
SELECT hf.id,
       hf.path,
       hf.physical_path,
       hf.size
FROM home_file hf
WHERE hf.id = $1
  AND hf.type = 2
  AND hf.deleted_at IS NULL;
