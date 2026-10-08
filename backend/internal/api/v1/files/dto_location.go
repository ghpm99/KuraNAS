package files

type FileLocationDto struct {
	FileID          int    `json:"file_id"`
	Tier            string `json:"tier"`
	LogicalPath     string `json:"logical_path"`
	DiskPath        string `json:"disk_path"`
	LogicalDiskPath string `json:"logical_disk_path"`
	RootLabel       string `json:"root_label"`
	ExistsOnDisk    bool   `json:"exists_on_disk"`
}
