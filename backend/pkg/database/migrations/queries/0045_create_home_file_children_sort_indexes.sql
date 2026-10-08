CREATE INDEX IF NOT EXISTS "home_file_active_children_by_name"
    ON "home_file" ("parent_path", "type", "name", "id" DESC)
    WHERE "deleted_at" IS NULL;

CREATE INDEX IF NOT EXISTS "home_file_active_children_by_size"
    ON "home_file" ("parent_path", "type", "size", "id" DESC)
    WHERE "deleted_at" IS NULL;

CREATE INDEX IF NOT EXISTS "home_file_active_children_by_updated_at"
    ON "home_file" ("parent_path", "type", "updated_at", "id" DESC)
    WHERE "deleted_at" IS NULL;

CREATE INDEX IF NOT EXISTS "home_file_active_children_by_created_at"
    ON "home_file" ("parent_path", "type", "created_at", "id" DESC)
    WHERE "deleted_at" IS NULL;
