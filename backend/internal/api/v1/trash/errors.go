package trash

import "errors"

var (
	// ErrItemNotFound — the trash item id does not exist in the registry.
	ErrItemNotFound = errors.New("trash item not found")
	// ErrRestoreConflict — the original path is occupied again; restoring would
	// overwrite live data.
	ErrRestoreConflict = errors.New("original path already exists")
	// ErrItemContentMissing — the registry row exists but its bytes are no
	// longer inside the trash directory, so there is nothing left to restore.
	ErrItemContentMissing = errors.New("trash item content is missing")
	// ErrInvalidRetention — retention must be a positive number of days.
	ErrInvalidRetention = errors.New("invalid retention days")
)
