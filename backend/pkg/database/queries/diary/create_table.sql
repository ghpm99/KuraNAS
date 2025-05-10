CREATE TABLE
    IF NOT EXISTS "activity_diary" (
        "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
        "name" varchar(256) NOT NULL,
        "description" varchar(256) NULL,
        "start_time" datetime NOT NULL,
        "end_time" datetime NULL
    );