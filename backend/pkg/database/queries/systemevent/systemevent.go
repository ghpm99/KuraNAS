package queries

import (
	_ "embed"
)

//go:embed insert_system_event.sql
var InsertSystemEventQuery string
