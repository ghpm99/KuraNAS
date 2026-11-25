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

	ErrCloudNotConfigured        = errors.New("vacuum: cloud credentials not configured")
	ErrCloudUnreachable          = errors.New("vacuum: cloud unreachable")
	ErrCloudCredentialsRejected  = errors.New("vacuum: cloud credentials rejected")
	ErrCloudServiceNotAuthorized = errors.New("vacuum: cloud service not authorized for the project")
	ErrCloudInvalidResponse      = errors.New("vacuum: cloud returned an unusable response")
	ErrInvalidCloudAccessID      = errors.New("vacuum: invalid cloud access id")
	ErrInvalidCloudAccessSecret  = errors.New("vacuum: invalid cloud access secret")
	ErrCloudAccessSecretRequired = errors.New("vacuum: cloud access secret required")
	ErrInvalidCloudRegion        = errors.New("vacuum: invalid cloud region")
	ErrMapChanged                = errors.New("vacuum: the robot map changed")
	ErrRoomNotFound              = errors.New("vacuum: room not found on the map")
	ErrNoRoomSelected            = errors.New("vacuum: no room selected")
	ErrInvalidRoomName           = errors.New("vacuum: invalid room name")
	ErrSpotNotOnFloor            = errors.New("vacuum: spot is not on a room floor")
)
