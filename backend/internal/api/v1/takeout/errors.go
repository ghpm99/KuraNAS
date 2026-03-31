package takeout

import "errors"

var (
	ErrInvalidZipFile        = errors.New("invalid zip file")
	ErrUploadSessionNotFound = errors.New("upload session not found")
	ErrUploadOffsetMismatch  = errors.New("upload offset mismatch")
	ErrUploadIncomplete      = errors.New("upload incomplete")
)
