SELECT COUNT(*) FILTER (WHERE physical_path IS NULL)                          AS hot_files,
       COALESCE(SUM(size) FILTER (WHERE physical_path IS NULL), 0)           AS hot_bytes,
       COUNT(*) FILTER (WHERE physical_path IS NOT NULL)                     AS cold_files,
       COALESCE(SUM(size) FILTER (WHERE physical_path IS NOT NULL), 0)       AS cold_bytes
FROM home_file
WHERE type = 2
  AND deleted_at IS NULL;
