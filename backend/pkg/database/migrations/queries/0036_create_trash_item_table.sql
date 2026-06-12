-- Recycle bin registry: one row per item moved into .kuranas-trash/.
-- original_path/trash_path are absolute; size is a snapshot taken at deletion
-- time so the trash listing does not have to stat the disk.
CREATE TABLE IF NOT EXISTS trash_item (
    id SERIAL PRIMARY KEY,
    original_path VARCHAR(1024) NOT NULL,
    trash_path VARCHAR(1024) NOT NULL UNIQUE,
    size BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS "trash_item_deleted_at" ON "trash_item" ("deleted_at");
