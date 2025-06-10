package queries

import (
	_ "embed"
)

//go:embed create_table.sql
var CreateTableQuery string

//go:embed insert_log.sql
var InsertLogQuery string

//go:embed get_log_by_id.sql
var GetLogByIDQuery string

//go:embed get_log_paginated.sql
var GetLogsQuery string

//go:embed update_log.sql
var UpdateLogQuery string
