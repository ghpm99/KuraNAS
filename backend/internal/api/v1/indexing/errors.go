package indexing

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidScope       = errors.New("invalid reindex scope")
	ErrIdentifierRequired = errors.New("reindex identifier required")
	ErrRootNotFound       = errors.New("storage root not found")
	ErrFolderNotFound     = errors.New("folder not found")
)

type AlreadyRunningError struct {
	JobID int
}

func (alreadyRunning *AlreadyRunningError) Error() string {
	return fmt.Sprintf("reindex already running in job %d", alreadyRunning.JobID)
}
