-- Updates label/enabled of a root. The path is immutable: changing it would
-- orphan every indexed row under the old path — delete and re-add instead.
UPDATE storage_root
SET
    label = $2,
    enabled = $3
WHERE
    id = $1
RETURNING id, path, label, enabled, created_at;
