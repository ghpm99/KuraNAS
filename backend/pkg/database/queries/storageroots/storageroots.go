package storageroots

import _ "embed"

//go:embed get_storage_roots.sql
var GetStorageRootsQuery string

//go:embed get_storage_root_by_id.sql
var GetStorageRootByIDQuery string

//go:embed insert_storage_root.sql
var InsertStorageRootQuery string

//go:embed update_storage_root.sql
var UpdateStorageRootQuery string

//go:embed delete_storage_root.sql
var DeleteStorageRootQuery string
