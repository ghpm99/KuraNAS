package indexing

import (
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/storageroots"
)

type ServiceInterface interface {
	Reindex(request ReindexRequestDto) (int, error)
}

type JobLauncher interface {
	LaunchReindex(reindexKey string, paths []string) (jobID int, isAlreadyRunning bool, err error)
}

type RootsReader interface {
	GetRoots() ([]storageroots.StorageRootDto, error)
}

type FolderReader interface {
	GetFileById(id int) (files.FileDto, error)
}
