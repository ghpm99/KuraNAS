package queries

import (
	_ "embed"
)

//go:embed get_settings.sql
var GetSettingsQuery string

//go:embed upsert_settings.sql
var UpsertSettingsQuery string
