package files

import "errors"

var (
    ErrFileNotFound    = errors.New("file not found")
    ErrInvalidFormat   = errors.New("unsupported file format")
    ErrFileMissingDisk = errors.New("file missing on disk")
    ErrDatabase        = errors.New("database error")
)