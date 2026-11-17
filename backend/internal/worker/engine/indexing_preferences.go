package engine

import "nas-go/api/internal/api/v1/configuration"

func currentIndexingPreferences(context *WorkerContext) configuration.IndexingPreferences {
	if context == nil || context.IndexingPreferences == nil {
		return configuration.DefaultIndexingPreferences()
	}
	return context.IndexingPreferences.IndexingPreferences()
}
