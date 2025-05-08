package queries

import (
	_ "embed"
)

//go:embed create_table.sql
var CreateTableQuery string

//go:embed insert_file.sql
var InsertFileQuery string

//go:embed get_files.sql
var GetFilesQuery string

//go:embed update_file.sql
var UpdateFileQuery string

//go:embed get_children_count.sql
var GetChildrenCountQuery string
