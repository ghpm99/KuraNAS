package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilepathJoin(t *testing.T) {
	withSep := FilepathJoin(true, "a", "b")
	if withSep[len(withSep)-1] != os.PathSeparator {
		t.Fatalf("expected trailing path separator, got %q", withSep)
	}

	withoutSep := FilepathJoin(false, "a", "b")
	if withoutSep[len(withoutSep)-1] == os.PathSeparator {
		t.Fatalf("expected no trailing separator, got %q", withoutSep)
	}
}

func TestGetWithFallback(t *testing.T) {
	t.Setenv("CFG_TEST_KEY", "value")
	if got := Get("CFG_TEST_KEY", "fallback"); got != "value" {
		t.Fatalf("expected env value, got %q", got)
	}
	if got := Get("CFG_TEST_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestInitializeConfig(t *testing.T) {
	t.Setenv("ENTRY_POINT", "/data")
	t.Setenv("LANGUAGE", "pt-BR")
	t.Setenv("ENABLE_WORKERS", "true")
	t.Setenv("ENV", "test")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "db")
	t.Setenv("WORKER_RETRY_DEFAULT_MAX_ATTEMPTS", "5")
	t.Setenv("WORKER_STEP_CONCURRENCY_CHECKSUM", "4")

	InitializeConfig()

	if AppConfig.EntryPoint != "/data" || !AppConfig.EnableWorkers || AppConfig.Lang != "pt-BR" {
		t.Fatalf("unexpected app config values: %+v", AppConfig)
	}
	if AppConfig.RecentFilesKeep != 10 {
		t.Fatalf("expected default recent files keep 10")
	}
	if AppConfig.WorkerRetryDefaultMaxAttempts != 5 {
		t.Fatalf("expected worker retry max attempts from env, got %d", AppConfig.WorkerRetryDefaultMaxAttempts)
	}
	if AppConfig.WorkerStepConcurrencyChecksum != 4 {
		t.Fatalf("expected checksum step concurrency from env, got %d", AppConfig.WorkerStepConcurrencyChecksum)
	}
}

func TestBuildConfigAndLoadConfig(t *testing.T) {
	// LoadConfig should not fail even if .env does not exist.
	if err := LoadConfig(); err != nil {
		t.Fatalf("expected LoadConfig success, got %v", err)
	}

	keys := []string{
		"BuildVersion",
		"DbPath",
		"IconPath",
		"TranslationsPath",
		"EnvFilePath",
		"PythonScript",
		"ScriptPath",
		"ThumbnailPath",
	}
	for _, key := range keys {
		value := GetBuildConfig(key)
		if value == "" {
			t.Fatalf("expected non-empty build config for %s", key)
		}
	}
	if got := GetBuildConfig("unknown-key"); got != "" {
		t.Fatalf("expected empty value for unknown key, got %q", got)
	}

	// Ensure path values look path-like.
	if ext := filepath.Ext(GetBuildConfig("EnvFilePath")); ext != ".env" {
		t.Fatalf("expected EnvFilePath to end with .env, got %q", GetBuildConfig("EnvFilePath"))
	}
}
