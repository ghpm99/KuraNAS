package queries

import (
	_ "embed"
)

//go:embed get_diary.sql
var GetDiaryQuery string

//go:embed insert_diary.sql
var InsertDiaryQuery string

//go:embed update_diary.sql
var UpdateDiaryQuery string
