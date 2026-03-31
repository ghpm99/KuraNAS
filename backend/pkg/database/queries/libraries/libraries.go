package queries

import (
	_ "embed"
)

//go:embed get_libraries.sql
var GetLibrariesQuery string

//go:embed get_library_by_category.sql
var GetLibraryByCategoryQuery string

//go:embed upsert_library.sql
var UpsertLibraryQuery string
