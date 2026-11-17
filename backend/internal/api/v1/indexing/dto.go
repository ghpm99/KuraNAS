package indexing

const (
	ScopeAll    = "all"
	ScopeRoot   = "root"
	ScopeFolder = "folder"
)

type ReindexRequestDto struct {
	Scope    string `json:"scope"`
	RootID   int    `json:"root_id"`
	FolderID int    `json:"folder_id"`
}

type ReindexJobDto struct {
	JobID int `json:"job_id"`
}

type ReindexConflictDto struct {
	Error string `json:"error"`
	JobID int    `json:"job_id"`
}
