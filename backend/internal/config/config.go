package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type AppConfigStruct struct {
	EntryPoint                 string
	Lang                       string
	EnableWorkers              bool
	StartupTime                time.Time
	RecentFilesKeep            int
	Env                        string
	DbHost                     string
	DbPort                     string
	DbUser                     string
	DbPassword                 string
	DbName                     string
	AllowedOrigins             string
	WorkerConcurrencyChecksum  int
	WorkerConcurrencyMetadata  int
	WorkerConcurrencyThumbnail int
	WorkerRetryBackoffMS       int
	WorkerSchedulerPollMS      int
	WorkerMaxConcurrentJobs    int
}

var AppConfig AppConfigStruct

func InitializeConfig() {
	AppConfig = AppConfigStruct{
		EntryPoint:                 os.Getenv("ENTRY_POINT"),
		Lang:                       os.Getenv("LANGUAGE"),
		EnableWorkers:              os.Getenv("ENABLE_WORKERS") == "true",
		StartupTime:                time.Now(),
		RecentFilesKeep:            10,
		Env:                        os.Getenv("ENV"),
		DbHost:                     os.Getenv("DB_HOST"),
		DbPort:                     os.Getenv("DB_PORT"),
		DbUser:                     os.Getenv("DB_USER"),
		DbPassword:                 os.Getenv("DB_PASSWORD"),
		DbName:                     os.Getenv("DB_NAME"),
		AllowedOrigins:             os.Getenv("ALLOWED_ORIGINS"),
		WorkerConcurrencyChecksum:  parseEnvInt("WORKER_CONCURRENCY_CHECKSUM", 3),
		WorkerConcurrencyMetadata:  parseEnvInt("WORKER_CONCURRENCY_METADATA", 3),
		WorkerConcurrencyThumbnail: parseEnvInt("WORKER_CONCURRENCY_THUMBNAIL", 2),
		WorkerRetryBackoffMS:       parseEnvInt("WORKER_RETRY_BACKOFF_MS", 500),
		WorkerSchedulerPollMS:      parseEnvInt("WORKER_SCHEDULER_POLL_MS", 2000),
		WorkerMaxConcurrentJobs:    parseEnvInt("WORKER_MAX_CONCURRENT_JOBS", 4),
	}
}

// ToRelativePath strips the EntryPoint prefix from an absolute path,
// returning a path relative to the entry point (e.g. "/imagens/fotos").
func ToRelativePath(absolutePath string) string {
	entryPoint := filepath.Clean(AppConfig.EntryPoint)
	cleaned := filepath.Clean(absolutePath)
	rel := strings.TrimPrefix(cleaned, entryPoint)
	if rel == "" || rel == "." {
		return "/"
	}
	if !strings.HasPrefix(rel, "/") {
		rel = "/" + rel
	}
	return rel
}

// ToAbsolutePath prepends the EntryPoint to a relative path.
func ToAbsolutePath(relativePath string) string {
	entryPoint := filepath.Clean(AppConfig.EntryPoint)
	if relativePath == "" || relativePath == "/" {
		return entryPoint
	}
	return filepath.Join(entryPoint, relativePath)
}

func parseEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
