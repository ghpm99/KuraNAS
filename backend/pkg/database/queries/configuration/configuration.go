package queries

import (
	_ "embed"
)

//go:embed get_setting.sql
var GetSettingQuery string

//go:embed upsert_setting.sql
var UpsertSettingQuery string
