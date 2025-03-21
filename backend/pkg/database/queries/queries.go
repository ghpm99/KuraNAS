package queries

import (
	_ "embed"
)

//go:embed file/create_table.sql
var CreateTableQuery string

//go:embed file/get_file_by_type.sql
var GetFileByTypeQuery string

//go:embed file/insert_file.sql
var InsertFileQuery string

//go:embed file/get_files.sql
var GetFilesQuery string
