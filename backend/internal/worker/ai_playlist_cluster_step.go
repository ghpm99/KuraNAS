package worker

import (
	"context"

	jobs "nas-go/api/internal/api/v1/jobs"
)

// executeAIPlaylistClusterStep rebuilds the AI-curated music playlists by asking
// the music service to (re)cluster artists and materialize one playlist per
// category. It runs on the worker so the (possibly slow, local) model never
// blocks an HTTP request. The step is skipped when no music service is wired in.
func executeAIPlaylistClusterStep(workerContext *WorkerContext, _ jobs.StepModel) error {
	if workerContext == nil || workerContext.MusicService == nil {
		return ErrStepSkipped
	}

	return workerContext.MusicService.RebuildAIClusters(context.Background())
}
