package worker

import (
	"context"

	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/systemevent"
)

// executeAIPlaylistClusterStep rebuilds the AI-curated music playlists by asking
// the music service to (re)cluster artists and materialize one playlist per
// category. It runs on the worker so the (possibly slow, local) model never
// blocks an HTTP request. The step is skipped when no music service is wired in.
func executeAIPlaylistClusterStep(workerContext *WorkerContext, _ jobs.StepModel) error {
	if workerContext == nil || workerContext.MusicService == nil {
		return ErrStepSkipped
	}

	// Hard backstop so a stalled local model can never freeze the worker slot;
	// the AI provider's own HTTP timeout is runtime-editable and may be 0.
	ctx, cancel := context.WithTimeout(context.Background(), config.StepTimeout())
	defer cancel()

	err := workerContext.MusicService.RebuildAIClusters(ctx)
	if err != nil && workerContext.SystemEvents != nil {
		// Record an audit marker (no error text — that lives in the forensic
		// file log) so a silently unreachable AI provider surfaces on the
		// dashboard. This is the exact operation that failed silently in prod.
		_ = workerContext.SystemEvents.RecordEvent(
			systemevent.EventTypeAIProviderUnavailable,
			i18n.GetMessage("SYSTEM_EVENT_AI_PROVIDER_UNAVAILABLE"),
		)
	}
	return err
}
