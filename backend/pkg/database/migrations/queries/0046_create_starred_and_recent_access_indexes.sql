CREATE INDEX IF NOT EXISTS "home_file_active_starred_by_name"
    ON "home_file" ("type", "name", "id" DESC)
    WHERE "starred" = TRUE AND "deleted_at" IS NULL;

CREATE INDEX IF NOT EXISTS "recent_file_by_file_and_accessed_at"
    ON "recent_file" ("file_id", "accessed_at" DESC);
