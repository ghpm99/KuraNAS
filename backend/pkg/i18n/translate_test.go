package i18n

import (
	"nas-go/api/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func setProgramFilesRootForTest(t *testing.T) {
	t.Helper()
	t.Setenv("ProgramFiles", filepath.Join(t.TempDir(), "ProgramFiles"))
}

func TestGetPathFileTranslateAndLoadTranslations(t *testing.T) {
	setProgramFilesRootForTest(t)
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
	setProgramFilesRootForTest(t)
	prevLang := config.AppConfig.Lang
	t.Cleanup(func() { config.AppConfig.Lang = prevLang })

	translationPath := GetPathFileTranslate()
	translationDir := filepath.Dir(translationPath)
	if err := os.MkdirAll(translationDir, 0755); err != nil {
		t.Fatalf("failed to create translation dir: %v", err)
	}

	validLang := "test-valid"
	config.AppConfig.Lang = validLang
	validPath := GetPathFileTranslate()
	if err := os.WriteFile(validPath, []byte(`{"HELLO":"ola"}`), 0644); err != nil {
		t.Fatalf("failed to write valid translation file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(validPath) })

	if err := LoadTranslations(); err != nil {
		t.Fatalf("expected LoadTranslations success, got %v", err)
	}
	if got := GetMessage("HELLO"); got != "ola" {
		t.Fatalf("expected loaded translation value, got %q", got)
	}

	invalidLang := "test-invalid"
	config.AppConfig.Lang = invalidLang
	invalidPath := GetPathFileTranslate()
	if err := os.WriteFile(invalidPath, []byte(`{invalid-json`), 0644); err != nil {
		t.Fatalf("failed to write invalid translation file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(invalidPath) })

	if err := LoadTranslations(); err == nil {
		t.Fatalf("expected json decode error for invalid translation content")
	}
}

func TestResolveTranslationsPathFallsBackToWorkspaceDirectory(t *testing.T) {
	setProgramFilesRootForTest(t)
	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousWorkingDir)
	}()

	root := t.TempDir()
	nested := filepath.Join(root, "workspace", "backend", "internal")
	fallbackDir := filepath.Join(root, "workspace", "backend", "translations")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}
	if err := os.MkdirAll(fallbackDir, 0755); err != nil {
		t.Fatalf("failed to create fallback dir: %v", err)
	}

	if err := os.Chdir(nested); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	resolved := ResolveTranslationsPath()
	if resolved != fallbackDir+string(os.PathSeparator) {
		t.Fatalf("expected fallback translations path %q, got %q", fallbackDir+string(os.PathSeparator), resolved)
	}

	detectedPath, ok := findFallbackTranslationsPath()
	if !ok || detectedPath != fallbackDir+string(os.PathSeparator) {
		t.Fatalf("expected findFallbackTranslationsPath to resolve %q, got %q ok=%v", fallbackDir+string(os.PathSeparator), detectedPath, ok)
	}
}

func TestGetPathFileTranslateByLangDefaultsAndFallbackMiss(t *testing.T) {
	setProgramFilesRootForTest(t)
	if path := GetPathFileTranslateByLang(""); path == "" || filepath.Base(path) != "en-US.json" {
		t.Fatalf("expected default en-US translation file, got %q", path)
	}

	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousWorkingDir)
	}()

	emptyRoot := t.TempDir()
	if err := os.Chdir(emptyRoot); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	if path, ok := findFallbackTranslationsPath(); ok || path != "" {
		t.Fatalf("expected no fallback translations path, got %q ok=%v", path, ok)
	}
}

func TestGetPathFileTranslateByLangSanitizesLocale(t *testing.T) {
	setProgramFilesRootForTest(t)
	testCases := map[string]string{
		" pt-BR ":            "pt-BR.json",
		"../secrets":         "en-US.json",
		`..\\secrets`:        "en-US.json",
		"pt-BR/../../secret": "en-US.json",
		"pt.BR":              "en-US.json",
		"pt BR":              "en-US.json",
	}

	for locale, expectedFile := range testCases {
		path := GetPathFileTranslateByLang(locale)
		if got := filepath.Base(path); got != expectedFile {
			t.Fatalf("expected %q for locale %q, got %q", expectedFile, locale, got)
		}
	}
}

func TestResolveTranslationFilePathKeepsFileWithinTranslationsDirectory(t *testing.T) {
	setProgramFilesRootForTest(t)
	testCases := map[string]string{
		"pt-BR":              "pt-BR.json",
		"../secrets":         "en-US.json",
		"pt-BR/../../secret": "en-US.json",
	}

	baseDir, err := filepath.Abs(ResolveTranslationsPath())
	if err != nil {
		t.Fatalf("failed to resolve translations path: %v", err)
	}

	for locale, expectedFile := range testCases {
		resolvedPath, err := resolveTranslationFilePath(locale)
		if err != nil {
			t.Fatalf("expected resolved path for locale %q, got error %v", locale, err)
		}
		if got := filepath.Base(resolvedPath); got != expectedFile {
			t.Fatalf("expected %q for locale %q, got %q", expectedFile, locale, got)
		}
		if !isPathWithinDirectory(resolvedPath, baseDir) {
			t.Fatalf("expected path %q to remain inside %q", resolvedPath, baseDir)
		}
	}
}

func TestResolveTranslationsPathPrefersConfiguredDirectoryAndRootTranslationsFallback(t *testing.T) {
	setProgramFilesRootForTest(t)
	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousWorkingDir)
	}()

	root := t.TempDir()
	nested := filepath.Join(root, "workspace", "internal")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir root failed: %v", err)
	}

	configuredPath := config.GetBuildConfig("TranslationsPath")
	if err := os.MkdirAll(configuredPath, 0755); err != nil {
		t.Fatalf("failed to create configured translations dir: %v", err)
	}
	if resolved := ResolveTranslationsPath(); resolved != configuredPath {
		t.Fatalf("expected configured translations path %q, got %q", configuredPath, resolved)
	}

	rootFallback := filepath.Join(root, "translations")
	if err := os.MkdirAll(rootFallback, 0755); err != nil {
		t.Fatalf("failed to create root fallback dir: %v", err)
	}
	if err := os.Chdir(nested); err != nil {
		t.Fatalf("Chdir nested failed: %v", err)
	}

	if detectedPath, ok := findFallbackTranslationsPath(); !ok || detectedPath != rootFallback+string(os.PathSeparator) {
		t.Fatalf("expected root fallback path %q, got %q ok=%v", rootFallback+string(os.PathSeparator), detectedPath, ok)
	}
}
