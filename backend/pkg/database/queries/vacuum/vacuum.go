package queries

import (
	_ "embed"
)

//go:embed get_settings.sql
var GetSettingsQuery string

//go:embed upsert_settings.sql
var UpsertSettingsQuery string

//go:embed get_cloud_settings.sql
var GetCloudSettingsQuery string

//go:embed upsert_cloud_settings.sql
var UpsertCloudSettingsQuery string

//go:embed get_room_names.sql
var GetRoomNamesQuery string

//go:embed upsert_room_names.sql
var UpsertRoomNamesQuery string
