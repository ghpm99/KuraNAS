package app

import (
	"nas-go/api/internal/api/v1/configuration"
	"nas-go/api/internal/api/v1/files"
)

type configurationProcessingStages struct {
	preferences configuration.IndexingPreferencesReader
}

func newConfigurationProcessingStages(preferences configuration.IndexingPreferencesReader) files.ProcessingStagesReader {
	return configurationProcessingStages{preferences: preferences}
}

func (adapter configurationProcessingStages) ProcessingStages() files.ProcessingStages {
	indexingPreferences := adapter.preferences.IndexingPreferences()
	return files.ProcessingStages{
		ExtractMetadata:  indexingPreferences.ExtractMetadata,
		GeneratePreviews: indexingPreferences.GeneratePreviews,
	}
}
