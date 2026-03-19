package ai

import "errors"

var (
	ErrNoProviderForTask = errors.New("no provider configured for task type")
	ErrEmptyPrompt       = errors.New("prompt cannot be empty")
	ErrProviderTimeout   = errors.New("provider request timed out")
	ErrProviderRateLimit = errors.New("provider rate limit exceeded")
	ErrProviderAuth      = errors.New("provider authentication failed")
	ErrAllAttemptsFailed = errors.New("all retry attempts failed")
)
