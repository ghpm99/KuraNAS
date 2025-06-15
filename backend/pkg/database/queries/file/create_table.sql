CREATE TABLE
    IF NOT EXISTS "home_file" (
        "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
        "name" varchar(256) NOT NULL,
        "path" varchar(1024) NOT NULL,
        "parent_path" varchar(1024) NOT NULL,
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

CREATE INDEX IF NOT EXISTS "home_file_parent_path" ON "home_file" ("parent_path");

CREATE INDEX IF NOT EXISTS "home_file_name" ON "home_file" ("name");

CREATE INDEX IF NOT EXISTS "home_file_path_name" ON "home_file" ("path", "name");

CREATE TABLE
    IF NOT EXISTS recent_file (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        ip_address VARCHAR(45) NOT NULL,
        file_id INTEGER NOT NULL,
        accessed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        UNIQUE (ip_address, file_id)
    );