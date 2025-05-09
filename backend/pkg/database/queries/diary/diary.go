package queries

import (
	_ "embed"
)

//go:embed create_table.sql
var CreateTableQuery string
