package watchfolders

import "errors"

var (
	ErrPathNotExists               = errors.New("watch folder path does not exist")
	ErrPathIsSubfolderOfEntryPoint = errors.New("watch folder path cannot be inside entry point")
	ErrPathAlreadyWatched          = errors.New("watch folder path already watched")
	ErrWatchFolderNotFound         = errors.New("watch folder not found")
	ErrInvalidWatchFolderID        = errors.New("invalid watch folder id")
)
