package tuyacloud

import "errors"

var (
	// ErrUnreachable means the Tuya cloud or the map file host could not be reached.
	ErrUnreachable = errors.New("tuyacloud: cloud unreachable")

	// ErrCredentialsRejected means the cloud refused the access id, the access secret or the token.
	ErrCredentialsRejected = errors.New("tuyacloud: credentials rejected")

	// ErrServiceNotAuthorized means the cloud project is not subscribed to the sweeping robot service.
	ErrServiceNotAuthorized = errors.New("tuyacloud: sweeping robot service not authorized for the project")

	// ErrInvalidResponse means the cloud or the map file answered with content that cannot be used.
	ErrInvalidResponse = errors.New("tuyacloud: invalid response")

	// ErrUnknownRegion means the credentials name a region that has no cloud endpoint.
	ErrUnknownRegion = errors.New("tuyacloud: unknown region")
)
