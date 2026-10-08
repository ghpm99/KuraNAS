UPDATE "home_file"
SET "format" = lower("format")
WHERE "format" <> lower("format");
