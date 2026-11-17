package engine

import (
	"testing"
	"time"

	"nas-go/api/internal/api/v1/configuration"
)

type stubIndexingPreferences struct {
	preferences configuration.IndexingPreferences
}

func (stub stubIndexingPreferences) IndexingPreferences() configuration.IndexingPreferences {
	return stub.preferences
}

func TestCurrentIndexingPreferencesFallsBackToDefaultsWithoutReader(t *testing.T) {
	if got := currentIndexingPreferences(nil); got != configuration.DefaultIndexingPreferences() {
		t.Fatalf("nil context: got %+v", got)
	}
	if got := currentIndexingPreferences(&WorkerContext{}); got != configuration.DefaultIndexingPreferences() {
		t.Fatalf("context without reader: got %+v", got)
	}
}

func TestCurrentIndexingPreferencesReadsTheInjectedReader(t *testing.T) {
	chosen := configuration.IndexingPreferences{GeneratePreviews: false, ExtractMetadata: true, ScanInterval: 6 * time.Hour}
	context := &WorkerContext{IndexingPreferences: stubIndexingPreferences{preferences: chosen}}

	if got := currentIndexingPreferences(context); got != chosen {
		t.Fatalf("got %+v, want %+v", got, chosen)
	}
}
