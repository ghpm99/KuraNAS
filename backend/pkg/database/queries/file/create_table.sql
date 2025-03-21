CREATE TABLE
    IF NOT EXISTS "home_file" (
        "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
        "name" varchar(256) NOT NULL,
        "path" varchar(512) NOT NULL,
        "format" varchar(256) NOT NULL,
        "size" integer NOT NULL,
        "updated_at" datetime NOT NULL,
        "created_at" datetime NOT NULL,
        "last_interaction" datetime NOT NULL,
        "last_backup" datetime NULL
    );