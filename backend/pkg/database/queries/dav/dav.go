package queries

import (
	_ "embed"
)

//go:embed list_cold_files_by_parent_path.sql
var ListColdFilesByParentPathQuery string
