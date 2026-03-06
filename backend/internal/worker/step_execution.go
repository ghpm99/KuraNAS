package worker

import (
	"context"
	"errors"
)

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

type transientStepError struct {
	cause error
}

func (e *transientStepError) Error() string {
	if e == nil || e.cause == nil {
		return "transient step error"
	}
	return e.cause.Error()
}

func (e *transientStepError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func newTransientStepError(cause error) error {
	if cause == nil {
		return &transientStepError{}
	}
	return &transientStepError{cause: cause}
}

func isTransientStepError(err error) bool {
	if err == nil {
		return false
	}
	var transientErr *transientStepError
	return errors.As(err, &transientErr)
}

func isStepCanceled(err error) bool {
	return errors.Is(err, context.Canceled)
}
