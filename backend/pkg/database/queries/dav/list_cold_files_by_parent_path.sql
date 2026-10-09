-- Active files of a directory whose bytes live on the cold tier. WebDAV merges
-- them into the listing because they are absent from the hot disk. $1 = parent path.
SELECT hf.name,
       hf.path,
       hf.physical_path,
       hf.size,
       hf.updated_at
FROM home_file hf
WHERE hf.parent_path = $1
  AND hf.type = 2
  AND hf.deleted_at IS NULL
  AND hf.physical_path IS NOT NULL
ORDER BY hf.name;
