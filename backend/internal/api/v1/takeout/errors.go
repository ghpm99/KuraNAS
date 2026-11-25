package takeout

import "errors"

var (
	ErrInvalidZipFile        = errors.New("invalid zip file")
	ErrUnsafeArchiveEntry    = errors.New("unsafe archive entry path")
	ErrMetadataEntryTooLarge = errors.New("metadata entry exceeds size limit")
	ErrUploadSessionNotFound = errors.New("upload session not found")
	ErrUploadOffsetMismatch  = errors.New("upload offset mismatch")
	ErrUploadIncomplete      = errors.New("upload incomplete")
	ErrInvalidUploadSize     = errors.New("upload size must be greater than zero")
	ErrInvalidUploadFileName = errors.New("upload file name must end in .zip")
	ErrInsufficientStorage   = errors.New("insufficient storage for upload")
)
