package ingest

import "errors"

var (
	ErrInvalidURL       = errors.New("a valid http(s) url is required")
	ErrInvalidPreset    = errors.New("unknown download preset")
	ErrInvalidTarget    = errors.New("target is not an enabled storage root")
	ErrInvalidSubfolder = errors.New("subfolder escapes the target root")
	ErrJobsUnavailable  = errors.New("jobs subsystem is not available")
)
