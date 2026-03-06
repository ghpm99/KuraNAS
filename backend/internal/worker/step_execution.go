package worker

import "errors"

type stepSkippedError struct {
	reason string
}

func (e *stepSkippedError) Error() string {
	if e.reason == "" {
		return "step skipped"
	}
	return e.reason
}

func newStepSkipped(reason string) error {
	return &stepSkippedError{reason: reason}
}

func isStepSkipped(err error) bool {
	if err == nil {
		return false
	}
	var skippedErr *stepSkippedError
	return errors.As(err, &skippedErr)
}
