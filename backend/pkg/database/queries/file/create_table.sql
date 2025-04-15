CREATE TABLE
    IF NOT EXISTS "home_file" (
        "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
        "name" varchar(256) NOT NULL,
        "path" varchar(512) NOT NULL,
        "format" varchar(256) NOT NULL,
        "size" integer NOT NULL,
        "updated_at" datetime NOT NULL,
        "created_at" datetime NOT NULL,
        "last_interaction" datetime NULL,
        "last_backup" datetime NULL,
        "type" INTEGER,
        "checksum" VARCHAR(64),
        "deleted_at" DATETIME NULL
    );

CREATE INDEX IF NOT EXISTS "home_file_path" ON "home_file" ("path");

CREATE INDEX IF NOT EXISTS "home_file_name" ON "home_file" ("name");

CREATE INDEX IF NOT EXISTS "home_file_path_name" ON "home_file" ("path", "name");