package vacuum

import "errors"

var (
	ErrNotConfigured     = errors.New("vacuum: device id or local key not configured")
	ErrDeviceUnreachable = errors.New("vacuum: device unreachable")
	ErrLocalKeyRejected  = errors.New("vacuum: local key rejected by device")
	ErrDiscoveryNotFound = errors.New("vacuum: no device accepted the device id and local key")
	ErrInvalidRequest    = errors.New("vacuum: invalid request")
	ErrInvalidCommand    = errors.New("vacuum: invalid command")
	ErrInvalidDirection  = errors.New("vacuum: invalid direction")
	ErrInvalidPart       = errors.New("vacuum: invalid consumable part")
	ErrInvalidPreference = errors.New("vacuum: invalid preference value")
	ErrInvalidAddress    = errors.New("vacuum: invalid address")
	ErrInvalidDeviceID   = errors.New("vacuum: invalid device id")
	ErrInvalidLocalKey   = errors.New("vacuum: invalid local key")
	ErrLocalKeyRequired  = errors.New("vacuum: local key required")
)
