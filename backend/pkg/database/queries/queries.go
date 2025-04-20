package queries

import (
	_ "embed"
)

//go:embed file/create_table.sql
var CreateTableQuery string

//go:embed file/insert_file.sql
var InsertFileQuery string

//go:embed file/get_files.sql
var GetFilesQuery string

//go:embed file/update_file.sql
var UpdateFileQuery string
