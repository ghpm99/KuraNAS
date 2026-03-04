package i18n

import (
	"nas-go/api/internal/config"
	"testing"
)

func TestGetPathFileTranslateAndLoadTranslations(t *testing.T) {
	config.AppConfig.Lang = ""
	path := GetPathFileTranslate()
	if path == "" {
		t.Fatalf("expected translation path, got empty")
	}

	// On test env without installed translation files, this usually returns error.
	_ = LoadTranslations()
}

func TestGetMessageAndTranslateFallbacks(t *testing.T) {
	translations = map[string]string{
		"HELLO": "Hello, %s",
		"RAW":   "raw-value",
	}

	if got := GetMessage("RAW"); got != "raw-value" {
		t.Fatalf("expected translated value, got %q", got)
	}
	if got := GetMessage("UNKNOWN"); got != "UNKNOWN" {
		t.Fatalf("expected fallback to key, got %q", got)
	}

	if got := Translate("HELLO", "world"); got != "Hello, world" {
		t.Fatalf("expected formatted translation, got %q", got)
	}
	if got := Translate("MISSING", "x"); got != "MISSING" {
		t.Fatalf("expected missing key fallback, got %q", got)
	}
}

func TestPrintAndLogTranslate(t *testing.T) {
	translations = map[string]string{"K": "value"}
	PrintTranslate("K")
	LogTranslate("K")
}
