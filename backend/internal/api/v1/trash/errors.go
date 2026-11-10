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
	// ErrInvalidListing — unknown sort key or direction in the listing request.
	ErrInvalidListing = errors.New("invalid trash listing")
)

// RestoreConflictError carries the path that blocks a restore, so the client
// can tell the user what is in the way.
type RestoreConflictError struct {
	OccupiedPath string
}

func (conflict *RestoreConflictError) Error() string {
	return ErrRestoreConflict.Error() + ": " + conflict.OccupiedPath
}

func (conflict *RestoreConflictError) Unwrap() error {
	return ErrRestoreConflict
}
