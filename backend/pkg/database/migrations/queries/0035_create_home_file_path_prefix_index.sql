-- SP-GiST radix index on path: serves the prefix predicates (path ^@ $n) of
-- the subtree walk (get_files_by_path_prefix), mark_deleted_subtree and
-- update_descendant_paths. The plain b-tree on path (0004) only serves
-- equality lookups, not prefix matching.
CREATE INDEX IF NOT EXISTS "home_file_path_prefix" ON "home_file" USING spgist ("path");
