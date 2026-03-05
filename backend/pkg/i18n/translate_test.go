package i18n

import (
	"nas-go/api/internal/config"
	"os"
	"path/filepath"
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

func TestLoadTranslationsSuccessAndInvalidJSON(t *testing.T) {
	prevLang := config.AppConfig.Lang
	t.Cleanup(func() { config.AppConfig.Lang = prevLang })

	translationDir := filepath.Join("etc", "kuranas", "translations")
	if err := os.MkdirAll(translationDir, 0755); err != nil {
		t.Fatalf("failed to create translation dir: %v", err)
	}

	validLang := "test-valid"
	validPath := filepath.Join(translationDir, validLang+".json")
	if err := os.WriteFile(validPath, []byte(`{"HELLO":"ola"}`), 0644); err != nil {
		t.Fatalf("failed to write valid translation file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(validPath) })

	config.AppConfig.Lang = validLang
	if err := LoadTranslations(); err != nil {
		t.Fatalf("expected LoadTranslations success, got %v", err)
	}
	if got := GetMessage("HELLO"); got != "ola" {
		t.Fatalf("expected loaded translation value, got %q", got)
	}

	invalidLang := "test-invalid"
	invalidPath := filepath.Join(translationDir, invalidLang+".json")
	if err := os.WriteFile(invalidPath, []byte(`{invalid-json`), 0644); err != nil {
		t.Fatalf("failed to write invalid translation file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(invalidPath) })

	config.AppConfig.Lang = invalidLang
	if err := LoadTranslations(); err == nil {
		t.Fatalf("expected json decode error for invalid translation content")
	}
}
