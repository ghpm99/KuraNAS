CREATE INDEX IF NOT EXISTS "home_file_active_physical_path"
    ON "home_file" ("physical_path")
    WHERE "physical_path" IS NOT NULL AND "deleted_at" IS NULL;
