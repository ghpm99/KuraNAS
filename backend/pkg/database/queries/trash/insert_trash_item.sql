-- Registers an item just moved into the trash directory.
INSERT INTO trash_item (original_path, trash_path, size, deleted_at)
VALUES ($1, $2, $3, $4)
RETURNING id, original_path, trash_path, size, deleted_at;
