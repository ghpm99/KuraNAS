package queries

import (
	_ "embed"
)

//go:embed get_backup_settings.sql
var GetBackupSettingsQuery string

//go:embed upsert_backup_settings.sql
var UpsertBackupSettingsQuery string

//go:embed count_pending_backup.sql
var CountPendingBackupQuery string

//go:embed get_last_backup_job.sql
var GetLastBackupJobQuery string

//go:embed update_last_backup_by_path.sql
var UpdateLastBackupByPathQuery string
