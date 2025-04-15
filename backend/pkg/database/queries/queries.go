package queries

import (
	_ "embed"
)

//go:embed file/create_table.sql
var CreateTableQuery string

//go:embed file/get_file_by_type.sql
var GetFileByTypeQuery string

//go:embed file/get_file_by_name_and_path.sql
var GetFileByNameAndPathQuery string

//go:embed file/insert_file.sql
var InsertFileQuery string

//go:embed file/get_files.sql
var GetFilesQuery string

//go:embed file/update_file.sql
var UpdateFileQuery string

//go:embed file/get_files_by_path.sql
var GetFilesByPathQuery string

//go:embed file/get_path_by_file_id.sql
var GetPathByFileIdQuery string
