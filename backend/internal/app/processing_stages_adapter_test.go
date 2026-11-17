package app

import (
	"testing"

	"nas-go/api/internal/api/v1/configuration"
	"nas-go/api/internal/api/v1/files"
)

type fixedIndexingPreferences struct {
	preferences configuration.IndexingPreferences
}

func (reader fixedIndexingPreferences) IndexingPreferences() configuration.IndexingPreferences {
	return reader.preferences
}

func TestConfigurationProcessingStagesMapsPreferences(t *testing.T) {
	cases := []struct {
		name        string
		preferences configuration.IndexingPreferences
		expected    files.ProcessingStages
	}{
		{"all on", configuration.IndexingPreferences{ExtractMetadata: true, GeneratePreviews: true}, files.ProcessingStages{ExtractMetadata: true, GeneratePreviews: true}},
		{"metadata off", configuration.IndexingPreferences{GeneratePreviews: true}, files.ProcessingStages{GeneratePreviews: true}},
		{"previews off", configuration.IndexingPreferences{ExtractMetadata: true}, files.ProcessingStages{ExtractMetadata: true}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			stages := newConfigurationProcessingStages(fixedIndexingPreferences{preferences: testCase.preferences}).ProcessingStages()
			if stages != testCase.expected {
				t.Fatalf("expected %+v, got %+v", testCase.expected, stages)
			}
		})
	}
}
