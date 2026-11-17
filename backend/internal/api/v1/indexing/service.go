package indexing

import (
	"database/sql"
	"errors"
	"fmt"

	"nas-go/api/internal/api/v1/files"
)

type Service struct {
	launcher JobLauncher
	roots    RootsReader
	folders  FolderReader
}

func NewService(launcher JobLauncher, roots RootsReader, folders FolderReader) *Service {
	return &Service{launcher: launcher, roots: roots, folders: folders}
}

// Reindex starts one job for the requested scope and returns its id. When a
// job of the same scope is still queued or running it returns an
// *AlreadyRunningError carrying that job's id.
func (service *Service) Reindex(request ReindexRequestDto) (int, error) {
	switch request.Scope {
	case ScopeAll:
		return service.reindexAllRoots()
	case ScopeRoot:
		return service.reindexRoot(request.RootID)
	case ScopeFolder:
		return service.reindexFolder(request.FolderID)
	default:
		return 0, ErrInvalidScope
	}
}

func (service *Service) reindexAllRoots() (int, error) {
	enabledRoots, err := service.enabledRoots()
	if err != nil {
		return 0, err
	}
	if len(enabledRoots) == 0 {
		return 0, ErrRootNotFound
	}
	rootPaths := make([]string, 0, len(enabledRoots))
	for _, enabledRoot := range enabledRoots {
		rootPaths = append(rootPaths, enabledRoot.Path)
	}
	return service.launch(ScopeAll, rootPaths)
}

func (service *Service) reindexRoot(rootID int) (int, error) {
	if rootID <= 0 {
		return 0, ErrIdentifierRequired
	}
	enabledRoots, err := service.enabledRoots()
	if err != nil {
		return 0, err
	}
	for _, enabledRoot := range enabledRoots {
		if enabledRoot.ID == rootID {
			return service.launch(fmt.Sprintf("%s:%d", ScopeRoot, rootID), []string{enabledRoot.Path})
		}
	}
	return 0, ErrRootNotFound
}

func (service *Service) reindexFolder(folderID int) (int, error) {
	if folderID <= 0 {
		return 0, ErrIdentifierRequired
	}
	folder, err := service.folders.GetFileById(folderID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrFolderNotFound
	}
	if err != nil {
		return 0, err
	}
	if folder.Type != files.Directory || folder.DeletedAt.HasValue {
		return 0, ErrFolderNotFound
	}
	return service.launch(fmt.Sprintf("%s:%d", ScopeFolder, folderID), []string{folder.Path})
}

func (service *Service) enabledRoots() ([]rootEntry, error) {
	registeredRoots, err := service.roots.GetRoots()
	if err != nil {
		return nil, err
	}
	enabledRoots := make([]rootEntry, 0, len(registeredRoots))
	for _, registeredRoot := range registeredRoots {
		if registeredRoot.Enabled {
			enabledRoots = append(enabledRoots, rootEntry{ID: registeredRoot.ID, Path: registeredRoot.Path})
		}
	}
	return enabledRoots, nil
}

func (service *Service) launch(reindexKey string, paths []string) (int, error) {
	jobID, isAlreadyRunning, err := service.launcher.LaunchReindex(reindexKey, paths)
	if err != nil {
		return 0, err
	}
	if isAlreadyRunning {
		return jobID, &AlreadyRunningError{JobID: jobID}
	}
	return jobID, nil
}

type rootEntry struct {
	ID   int
	Path string
}
