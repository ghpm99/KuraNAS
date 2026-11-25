package captures

import (
	"errors"
	"fmt"
)

var ErrCaptureNotFound = errors.New("capture not found")

var ErrCaptureNotRetryable = errors.New("capture promotion can only be retried when failed")

var ErrPromotionDispatcherUnavailable = errors.New("capture promotion dispatcher is not configured")

// RetryCapturePromotion enqueues a new promotion job for a capture whose
// promotion failed and returns the job id. It returns ErrCaptureNotFound when
// the capture does not exist and ErrCaptureNotRetryable when its status is not
// failed.
func (s *Service) RetryCapturePromotion(captureID int) (int, error) {
	capture, err := s.findCaptureModel(captureID)
	if errors.Is(err, ErrCaptureNotFound) {
		return 0, err
	}
	if err != nil {
		return 0, fmt.Errorf("RetryCapturePromotion: %w", err)
	}
	if capture.Status != CaptureStatusFailed {
		return 0, ErrCaptureNotRetryable
	}
	if s.UploadJobDispatcher == nil {
		return 0, ErrPromotionDispatcherUnavailable
	}
	return s.UploadJobDispatcher.CreateCaptureProcessJob(captureID)
}
