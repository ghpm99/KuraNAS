package engine

import (
	"errors"

	ingestapi "nas-go/api/internal/api/v1/ingest"
	"nas-go/api/pkg/i18n"
)

const unknownDownloadFailureKey = "INGEST_DOWNLOAD_ERROR_UNKNOWN"

// remoteFetchFailure carries the translation key of the reason a download
// failed, so the stored record shows a translated message free of server paths
// while the error itself keeps the technical detail for the logs.
type remoteFetchFailure struct {
	messageKey string
	cause      error
}

func (failure *remoteFetchFailure) Error() string { return failure.cause.Error() }

func (failure *remoteFetchFailure) Unwrap() error { return failure.cause }

func failedBecause(messageKey string, cause error) error {
	return &remoteFetchFailure{messageKey: messageKey, cause: cause}
}

func translatedDownloadFailure(fetchErr error) string {
	var failure *remoteFetchFailure
	if errors.As(fetchErr, &failure) {
		return i18n.GetMessage(failure.messageKey)
	}
	return i18n.GetMessage(unknownDownloadFailureKey)
}

type remoteFetchOutcome struct {
	title     string
	finalPath string
}

// downloadRecord keeps one ingest_download row in step with the running step.
// Without a recorder or a download id (jobs enqueued before records existed)
// every method is a no-op.
type downloadRecord struct {
	recorder   ingestapi.DownloadRecorderInterface
	downloadID int
}

func newDownloadRecord(workerContext *WorkerContext, downloadID int) downloadRecord {
	if workerContext == nil || workerContext.DownloadRecorder == nil || downloadID <= 0 {
		return downloadRecord{}
	}
	return downloadRecord{recorder: workerContext.DownloadRecorder, downloadID: downloadID}
}

func (record downloadRecord) isTracked() bool {
	return record.recorder != nil
}

func (record downloadRecord) markRunning() {
	if record.isTracked() {
		_ = record.recorder.MarkRunning(record.downloadID)
	}
}

func (record downloadRecord) reportProgress(progress ingestapi.DownloadProgress) {
	if record.isTracked() {
		_ = record.recorder.UpdateProgress(record.downloadID, progress)
	}
}

func (record downloadRecord) reportTitle(title string) {
	if record.isTracked() {
		_ = record.recorder.UpdateTitle(record.downloadID, title)
	}
}

func (record downloadRecord) finish(fetchErr error, outcome remoteFetchOutcome) {
	if !record.isTracked() {
		return
	}
	switch {
	case fetchErr == nil:
		_ = record.recorder.MarkCompleted(record.downloadID, outcome.title, outcome.finalPath)
	case errors.Is(fetchErr, ErrStepCanceled):
		_ = record.recorder.MarkCanceled(record.downloadID)
	default:
		_ = record.recorder.MarkFailed(record.downloadID, translatedDownloadFailure(fetchErr))
	}
}
