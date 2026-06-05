package worker

import (
	"context"

	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
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

	return workerContext.MusicService.RebuildAIClusters(ctx)
}
