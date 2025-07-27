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

}

func diaryMigrationList() {
	addMigration("0001_create_diary_table",
		defaultMigrationFunc(CreateDiaryTableQuery))
}

func logMigrationList() {
	addMigration("0001_create_log_table",
		defaultMigrationFunc(CreateLogTableQuery))
}
