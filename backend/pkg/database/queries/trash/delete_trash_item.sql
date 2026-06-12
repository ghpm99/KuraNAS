-- Removes one item from the registry (after restore or definitive purge).
DELETE FROM trash_item
WHERE id = $1;
