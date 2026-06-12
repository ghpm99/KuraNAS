-- Prefix swap for every descendant of a moved/renamed directory.
-- $1 = old directory path, $2 = new directory path, $3 = old path + OS separator.
-- Literal prefix operator ^@ (starts_with), never LIKE: PostgreSQL treats '\'
-- as the LIKE escape character, so Windows paths (e.g. D:\Folder) would
-- silently match nothing. The operator form is served by the SP-GiST index
-- home_file_path_prefix; a bare starts_with() call is not.
UPDATE home_file
SET
    path = $2 || substr(path, length($1) + 1),
    parent_path = $2 || substr(parent_path, length($1) + 1)
WHERE
    path ^@ $3
    OR parent_path = $1;
