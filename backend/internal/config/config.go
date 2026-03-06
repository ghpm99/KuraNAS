package config

import (
	"os"
	"strconv"
	"time"
)

type AppConfigStruct struct {
	EntryPoint                          string
	Lang                                string
	EnableWorkers                       bool
	StartupTime                         time.Time
	RecentFilesKeep                     int
	Env                                 string
	DbHost                              string
	DbPort                              string
	DbUser                              string
	DbPassword                          string
	DbName                              string
	AllowedOrigins                      string
	WorkerSchedulerPollIntervalSecond   int
	WorkerMaxJobsPerTick                int
	WorkerRetryBaseBackoffMillis        int
	WorkerRetryMaxBackoffMillis         int
	WorkerRetryDefaultMaxAttempts       int
	WorkerStepConcurrencyDefault        int
	WorkerStepConcurrencyScanFilesystem int
	WorkerStepConcurrencyDiffAgainstDB  int
	WorkerStepConcurrencyMetadata       int
	WorkerStepConcurrencyChecksum       int
	WorkerStepConcurrencyPersist        int
	WorkerStepConcurrencyThumbnail      int
	WorkerStepConcurrencyPlaylistIndex  int
	WorkerStepConcurrencyMarkDeleted    int
}

var AppConfig AppConfigStruct

func InitializeConfig() {
	AppConfig = AppConfigStruct{
		EntryPoint:                          os.Getenv("ENTRY_POINT"),
		Lang:                                os.Getenv("LANGUAGE"),
		EnableWorkers:                       os.Getenv("ENABLE_WORKERS") == "true",
		StartupTime:                         time.Now(),
		RecentFilesKeep:                     10,
		Env:                                 os.Getenv("ENV"),
		DbHost:                              os.Getenv("DB_HOST"),
		DbPort:                              os.Getenv("DB_PORT"),
		DbUser:                              os.Getenv("DB_USER"),
		DbPassword:                          os.Getenv("DB_PASSWORD"),
		DbName:                              os.Getenv("DB_NAME"),
		AllowedOrigins:                      os.Getenv("ALLOWED_ORIGINS"),
		WorkerSchedulerPollIntervalSecond:   getEnvInt("WORKER_SCHEDULER_POLL_INTERVAL_SECONDS", 2),
		WorkerMaxJobsPerTick:                getEnvInt("WORKER_MAX_JOBS_PER_TICK", 50),
		WorkerRetryBaseBackoffMillis:        getEnvInt("WORKER_RETRY_BASE_BACKOFF_MILLIS", 500),
		WorkerRetryMaxBackoffMillis:         getEnvInt("WORKER_RETRY_MAX_BACKOFF_MILLIS", 30000),
		WorkerRetryDefaultMaxAttempts:       getEnvInt("WORKER_RETRY_DEFAULT_MAX_ATTEMPTS", 3),
		WorkerStepConcurrencyDefault:        getEnvInt("WORKER_STEP_CONCURRENCY_DEFAULT", 1),
		WorkerStepConcurrencyScanFilesystem: getEnvInt("WORKER_STEP_CONCURRENCY_SCAN_FILESYSTEM", 1),
		WorkerStepConcurrencyDiffAgainstDB:  getEnvInt("WORKER_STEP_CONCURRENCY_DIFF_AGAINST_DB", 1),
		WorkerStepConcurrencyMetadata:       getEnvInt("WORKER_STEP_CONCURRENCY_METADATA", 2),
		WorkerStepConcurrencyChecksum:       getEnvInt("WORKER_STEP_CONCURRENCY_CHECKSUM", 2),
		WorkerStepConcurrencyPersist:        getEnvInt("WORKER_STEP_CONCURRENCY_PERSIST", 2),
		WorkerStepConcurrencyThumbnail:      getEnvInt("WORKER_STEP_CONCURRENCY_THUMBNAIL", 1),
		WorkerStepConcurrencyPlaylistIndex:  getEnvInt("WORKER_STEP_CONCURRENCY_PLAYLIST_INDEX", 1),
		WorkerStepConcurrencyMarkDeleted:    getEnvInt("WORKER_STEP_CONCURRENCY_MARK_DELETED", 1),
	}
}

func getEnvInt(key string, fallback int) int {
	rawValue := os.Getenv(key)
	if rawValue == "" {
		return fallback
	}

	parsedValue, err := strconv.Atoi(rawValue)
	if err != nil {
		return fallback
	}

	return parsedValue
}
