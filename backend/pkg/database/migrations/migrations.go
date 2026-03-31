package migrations

import (
	"database/sql"
	_ "embed"
)

//go:embed queries/0001_create_home_file_table.sql
var CreateHomeFileTableQuery string

//go:embed queries/0002_add_file_starred_column.sql
var CreateFileStarredColumnQuery string

//go:embed queries/0003_create_recent_file_table.sql
var CreateRecentFileTableQuery string

//go:embed queries/0004_create_home_file_index.sql
var CreateHomeFileIndex4Query string

//go:embed queries/0005_create_home_file_index.sql
var CreateHomeFileIndex5Query string

//go:embed queries/0006_create_home_file_index.sql
var CreateHomeFileIndex6Query string

//go:embed queries/0007_create_home_file_index.sql
var CreateHomeFileIndex7Query string

//go:embed queries/0008_create_image_metadata_table.sql
var CreateImageMetadataTableQuery string

//go:embed queries/0009_create_audio_metadata_table.sql
var CreateAudioMetadataTableQuery string

//go:embed queries/0010_create_video_metadata_table.sql
var CreateVideoMetadataTableQuery string

//go:embed queries/0001_create_diary_table.sql
var CreateDiaryTableQuery string

//go:embed queries/0001_create_log_table.sql
var CreateLogTableQuery string

//go:embed queries/0011_create_playlist_table.sql
var CreatePlaylistTableQuery string

//go:embed queries/0012_create_playlist_track_table.sql
var CreatePlaylistTrackTableQuery string

//go:embed queries/0013_create_player_state_table.sql
var CreatePlayerStateTableQuery string

//go:embed queries/0014_create_video_playback_tables.sql
var CreateVideoPlaybackTablesQuery string

//go:embed queries/0015_extend_video_playlist_for_smart_grouping.sql
var ExtendVideoPlaylistForSmartGroupingQuery string

//go:embed queries/0016_create_worker_job_tables.sql
var CreateWorkerJobTablesQuery string

//go:embed queries/0017_create_video_behavior_event_table.sql
var CreateVideoBehaviorEventTableQuery string

//go:embed queries/0018_update_video_playlist_type_check.sql
var UpdateVideoPlaylistTypeCheckQuery string

//go:embed queries/0019_add_image_classification_columns.sql
var AddImageClassificationColumnsQuery string

//go:embed queries/0020_create_app_settings_table.sql
var CreateAppSettingsTableQuery string

//go:embed queries/0021_create_notifications_table.sql
var CreateNotificationsTableQuery string

//go:embed queries/0022_create_system_event_log_table.sql
var CreateSystemEventLogTableQuery string

//go:embed queries/0023_create_captures_table.sql
var CreateCapturesTableQuery string

func defaultMigrationFunc(query string) func(tx *sql.Tx) error {
	return func(tx *sql.Tx) error {
		_, err := tx.Exec(query)
		return err
	}
}

func fileMigrationList() {
	addMigration("0001_create_home_file_table",
		defaultMigrationFunc(CreateHomeFileTableQuery))

	addMigration("0002_add_file_starred_column",
		defaultMigrationFunc(CreateFileStarredColumnQuery))

	addMigration("0003_create_recent_file_table",
		defaultMigrationFunc(CreateRecentFileTableQuery))

	addMigration("0004_create_home_file_index_4",
		defaultMigrationFunc(CreateHomeFileIndex4Query))

	addMigration("0005_create_home_file_index_5",
		defaultMigrationFunc(CreateHomeFileIndex5Query))

	addMigration("0006_create_home_file_index_6",
		defaultMigrationFunc(CreateHomeFileIndex6Query))

	addMigration("0007_create_home_file_index_7",
		defaultMigrationFunc(CreateHomeFileIndex7Query))

	addMigration("0008_create_image_metadata_table",
		defaultMigrationFunc(CreateImageMetadataTableQuery))

	addMigration("0009_create_audio_metadata_table",
		defaultMigrationFunc(CreateAudioMetadataTableQuery))

	addMigration("0010_create_video_metadata_table",
		defaultMigrationFunc(CreateVideoMetadataTableQuery))

	addMigration("0019_add_image_classification_columns",
		defaultMigrationFunc(AddImageClassificationColumnsQuery))

}

func diaryMigrationList() {
	addMigration("0001_create_diary_table",
		defaultMigrationFunc(CreateDiaryTableQuery))
}

func logMigrationList() {
	addMigration("0001_create_log_table",
		defaultMigrationFunc(CreateLogTableQuery))
}

func musicMigrationList() {
	addMigration("0011_create_playlist_table",
		defaultMigrationFunc(CreatePlaylistTableQuery))

	addMigration("0012_create_playlist_track_table",
		defaultMigrationFunc(CreatePlaylistTrackTableQuery))

	addMigration("0013_create_player_state_table",
		defaultMigrationFunc(CreatePlayerStateTableQuery))
}

func videoMigrationList() {
	addMigration("0014_create_video_playback_tables",
		defaultMigrationFunc(CreateVideoPlaybackTablesQuery))

	addMigration("0015_extend_video_playlist_for_smart_grouping",
		defaultMigrationFunc(ExtendVideoPlaylistForSmartGroupingQuery))

	addMigration("0017_create_video_behavior_event_table",
		defaultMigrationFunc(CreateVideoBehaviorEventTableQuery))

	addMigration("0018_update_video_playlist_type_check",
		defaultMigrationFunc(UpdateVideoPlaylistTypeCheckQuery))
}

func workerMigrationList() {
	addMigration("0016_create_worker_job_tables",
		defaultMigrationFunc(CreateWorkerJobTablesQuery))
}

func configurationMigrationList() {
	addMigration("0020_create_app_settings_table",
		defaultMigrationFunc(CreateAppSettingsTableQuery))
}

func notificationsMigrationList() {
	addMigration("0021_create_notifications_table",
		defaultMigrationFunc(CreateNotificationsTableQuery))
}

func systemEventMigrationList() {
	addMigration("0022_create_system_event_log_table",
		defaultMigrationFunc(CreateSystemEventLogTableQuery))
}

func capturesMigrationList() {
	addMigration("0023_create_captures_table",
		defaultMigrationFunc(CreateCapturesTableQuery))
}
