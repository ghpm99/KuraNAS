-- Unregisters a storage root (indexed home_file rows are kept).
DELETE FROM storage_root
WHERE id = $1;
