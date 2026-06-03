package ollama

import "errors"

var (
	ErrInvalidModelName = errors.New("model name is required")
	ErrModelNotFound    = errors.New("model not found")
	ErrJobsUnavailable  = errors.New("jobs subsystem is not available")
)
