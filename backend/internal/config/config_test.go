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
	t.Setenv("WORKER_CONCURRENCY_CHECKSUM", "7")
	t.Setenv("WORKER_CONCURRENCY_METADATA", "6")
	t.Setenv("WORKER_CONCURRENCY_THUMBNAIL", "5")
	t.Setenv("WORKER_RETRY_BACKOFF_MS", "1500")
	t.Setenv("WORKER_SCHEDULER_POLL_MS", "2500")

	InitializeConfig()

	if AppConfig.EntryPoint != "/data" || !AppConfig.EnableWorkers || AppConfig.Lang != "pt-BR" {
		t.Fatalf("unexpected app config values: %+v", AppConfig)
	}
	if AppConfig.RecentFilesKeep != 10 {
		t.Fatalf("expected default recent files keep 10")
	}
	if AppConfig.WorkerConcurrencyChecksum != 7 ||
		AppConfig.WorkerConcurrencyMetadata != 6 ||
		AppConfig.WorkerConcurrencyThumbnail != 5 ||
		AppConfig.WorkerRetryBackoffMS != 1500 ||
		AppConfig.WorkerSchedulerPollMS != 2500 {
		t.Fatalf("unexpected worker config values: %+v", AppConfig)
	}
}

func TestInitializeConfigFallbackWorkerValues(t *testing.T) {
	t.Setenv("ENTRY_POINT", "/base")
	t.Setenv("LANGUAGE", "en-US")
	t.Setenv("ENABLE_WORKERS", "false")
	t.Setenv("WORKER_CONCURRENCY_CHECKSUM", "")
	t.Setenv("WORKER_CONCURRENCY_METADATA", "0")
	t.Setenv("WORKER_CONCURRENCY_THUMBNAIL", "invalid")
	t.Setenv("WORKER_RETRY_BACKOFF_MS", "-1")
	t.Setenv("WORKER_SCHEDULER_POLL_MS", "abc")
	t.Setenv("WORKER_MAX_CONCURRENT_JOBS", "")

	InitializeConfig()

	if AppConfig.WorkerConcurrencyChecksum != 3 {
		t.Fatalf("expected default checksum concurrency 3, got %d", AppConfig.WorkerConcurrencyChecksum)
	}
	if AppConfig.WorkerConcurrencyMetadata != 3 {
		t.Fatalf("expected default metadata concurrency 3, got %d", AppConfig.WorkerConcurrencyMetadata)
	}
	if AppConfig.WorkerConcurrencyThumbnail != 2 {
		t.Fatalf("expected default thumbnail concurrency 2, got %d", AppConfig.WorkerConcurrencyThumbnail)
	}
	if AppConfig.WorkerRetryBackoffMS != 500 {
		t.Fatalf("expected default retry backoff 500, got %d", AppConfig.WorkerRetryBackoffMS)
	}
	if AppConfig.WorkerSchedulerPollMS != 2000 {
		t.Fatalf("expected default scheduler poll 2000, got %d", AppConfig.WorkerSchedulerPollMS)
	}
	if AppConfig.WorkerMaxConcurrentJobs != 4 {
		t.Fatalf("expected default max jobs 4, got %d", AppConfig.WorkerMaxConcurrentJobs)
	}
}

func TestRelativeAndAbsolutePathHelpers(t *testing.T) {
	AppConfig.EntryPoint = "/data/root"

	if got := ToRelativePath("/data/root"); got != "/" {
		t.Fatalf("expected root relative path '/', got %q", got)
	}

	if got := ToRelativePath("/data/root/library/file.mp4"); got != "/library/file.mp4" {
		t.Fatalf("expected relative file path, got %q", got)
	}

	if got := ToAbsolutePath(""); got != "/data/root" {
		t.Fatalf("expected entrypoint for empty relative path, got %q", got)
	}
	if got := ToAbsolutePath("/"); got != "/data/root" {
		t.Fatalf("expected entrypoint for '/' relative path, got %q", got)
	}
	if got := ToAbsolutePath("library/file.mp4"); got != filepath.Clean("/data/root/library/file.mp4") {
		t.Fatalf("expected joined absolute path, got %q", got)
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
