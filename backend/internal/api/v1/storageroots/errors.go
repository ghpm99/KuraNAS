package storageroots

import "errors"

var (
	// ErrRootNotFound — id not registered.
	ErrRootNotFound = errors.New("storage root not found")
	// ErrInvalidRootPath — empty path, not absolute, missing on disk or not a directory.
	ErrInvalidRootPath = errors.New("invalid storage root path")
	// ErrOverlappingRoot — path is ancestor or descendant of a registered root.
	ErrOverlappingRoot = errors.New("storage root overlaps a registered root")
	// ErrDuplicateRoot — path or label already registered.
	ErrDuplicateRoot = errors.New("storage root already registered")
	// ErrInvalidRootLabel — empty or path-like label.
	ErrInvalidRootLabel = errors.New("invalid storage root label")
	// ErrPrimaryRootImmutable — the seeded primary root cannot be disabled/removed.
	ErrPrimaryRootImmutable = errors.New("the primary storage root cannot be disabled or removed")
)
