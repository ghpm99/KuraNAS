package takeout

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidZipFile         = errors.New("invalid zip file")
	ErrUnsafeArchiveEntry     = errors.New("unsafe archive entry path")
	ErrMetadataEntryTooLarge  = errors.New("metadata entry exceeds size limit")
	ErrUploadSessionNotFound  = errors.New("upload session not found")
	ErrUploadOffsetMismatch   = errors.New("upload offset mismatch")
	ErrUploadChunkExceedsSize = errors.New("upload chunk exceeds declared size")
	ErrUploadAlreadyCompleted = errors.New("upload already completed")
	ErrUploadIncomplete       = errors.New("upload incomplete")
	ErrInvalidUploadSize      = errors.New("upload size must be greater than zero")
	ErrInvalidUploadFileName  = errors.New("upload file name must end in .zip")
	ErrInsufficientStorage    = errors.New("insufficient storage for upload")
)

type UploadIncompleteError struct {
	ExpectedSize int64
	ReceivedSize int64
}

func (incompleteError UploadIncompleteError) Error() string {
	return fmt.Sprintf("%s: expected %d bytes, received %d", ErrUploadIncomplete.Error(), incompleteError.ExpectedSize, incompleteError.ReceivedSize)
}

func (incompleteError UploadIncompleteError) Is(target error) bool {
	return target == ErrUploadIncomplete
}
