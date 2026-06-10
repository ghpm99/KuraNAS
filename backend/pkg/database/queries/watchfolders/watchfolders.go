package queries

import (
	_ "embed"
)

//go:embed get_watch_folders.sql
var GetWatchFoldersQuery string

//go:embed get_watch_folder_by_id.sql
var GetWatchFolderByIDQuery string

//go:embed create_watch_folder.sql
var CreateWatchFolderQuery string

//go:embed update_watch_folder.sql
var UpdateWatchFolderQuery string

//go:embed delete_watch_folder.sql
var DeleteWatchFolderQuery string

//go:embed update_watch_folder_last_scan.sql
var UpdateWatchFolderLastScanQuery string
