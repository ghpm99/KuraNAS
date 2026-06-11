-- Prefix swap for every descendant of a moved/renamed directory.
-- $1 = old directory path, $2 = new directory path, $3 = old path + OS separator.
-- Literal starts_with(), never LIKE: PostgreSQL treats '\' as the LIKE escape
-- character, so Windows paths (e.g. D:\Folder) would silently match nothing.
UPDATE home_file
SET
    path = $2 || substr(path, length($1) + 1),
    parent_path = $2 || substr(parent_path, length($1) + 1)
WHERE
    starts_with(path, $3)
    OR parent_path = $1;
