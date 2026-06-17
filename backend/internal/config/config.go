package config

import (
	"os"
	"strconv"
	"time"
)

type AppConfigStruct struct {
	EntryPoint                 string
	Lang                       string
	EnableWorkers              bool
	EnableWebDAV               bool
	StartupTime                time.Time
	RecentFilesKeep            int
	Env                        string
	DbHost                     string
	DbPort                     string
	DbUser                     string
	DbPassword                 string
	DbName                     string
	WorkerConcurrencyChecksum  int
	WorkerConcurrencyMetadata  int
	WorkerConcurrencyThumbnail int
	WorkerRetryBackoffMS       int
	WorkerSchedulerPollMS      int
	WorkerMaxConcurrentJobs    int
	WorkerStepTimeoutSeconds   int
	WorkerHeartbeatSeconds     int
	WatcherReconcileHours      int
	LogLevel                   string
	LogMaxSizeMB               int
	LogMaxBackups              int
	LogMaxAgeDays              int
	EmailTokenKey              string
	EmailGoogleClientID        string
	EmailGoogleClientSecret    string
	EmailMSClientID            string
	EmailSyncIntervalMinutes   int
	EmailRetentionDays         int
	EmailMaxMessagesPerAccount int
	YtDlpPath                  string
	YtDlpCheckHours            int
}

var AppConfig AppConfigStruct

func InitializeConfig() {
	AppConfig = AppConfigStruct{
		EntryPoint:                 os.Getenv("ENTRY_POINT"),
		Lang:                       os.Getenv("LANGUAGE"),
		EnableWorkers:              os.Getenv("ENABLE_WORKERS") == "true",
		EnableWebDAV:               os.Getenv("WEBDAV_ENABLED") == "true",
		StartupTime:                time.Now(),
		RecentFilesKeep:            10,
		Env:                        os.Getenv("ENV"),
		DbHost:                     os.Getenv("DB_HOST"),
		DbPort:                     os.Getenv("DB_PORT"),
		DbUser:                     os.Getenv("DB_USER"),
		DbPassword:                 os.Getenv("DB_PASSWORD"),
		DbName:                     os.Getenv("DB_NAME"),
		WorkerConcurrencyChecksum:  parseEnvInt("WORKER_CONCURRENCY_CHECKSUM", 3),
		WorkerConcurrencyMetadata:  parseEnvInt("WORKER_CONCURRENCY_METADATA", 3),
		WorkerConcurrencyThumbnail: parseEnvInt("WORKER_CONCURRENCY_THUMBNAIL", 2),
		WorkerRetryBackoffMS:       parseEnvInt("WORKER_RETRY_BACKOFF_MS", 500),
		WorkerSchedulerPollMS:      parseEnvInt("WORKER_SCHEDULER_POLL_MS", 2000),
		WorkerMaxConcurrentJobs:    parseEnvInt("WORKER_MAX_CONCURRENT_JOBS", 4),
		WorkerStepTimeoutSeconds:   parseEnvInt("WORKER_STEP_TIMEOUT_SECONDS", 120),
		WorkerHeartbeatSeconds:     parseEnvInt("WORKER_HEARTBEAT_SECONDS", 60),
		WatcherReconcileHours:      parseEnvInt("WATCHER_RECONCILE_HOURS", 24),
		LogLevel:                   os.Getenv("LOG_LEVEL"),
		LogMaxSizeMB:               parseEnvInt("LOG_MAX_SIZE_MB", 50),
		LogMaxBackups:              parseEnvInt("LOG_MAX_BACKUPS", 10),
		LogMaxAgeDays:              parseEnvInt("LOG_MAX_AGE_DAYS", 30),
		EmailTokenKey:              os.Getenv("EMAIL_TOKEN_KEY"),
		EmailGoogleClientID:        os.Getenv("EMAIL_GOOGLE_CLIENT_ID"),
		EmailGoogleClientSecret:    os.Getenv("EMAIL_GOOGLE_CLIENT_SECRET"),
		EmailMSClientID:            os.Getenv("EMAIL_MS_CLIENT_ID"),
		EmailSyncIntervalMinutes:   parseEnvInt("EMAIL_SYNC_INTERVAL_MINUTES", 10),
		EmailRetentionDays:         parseEnvInt("EMAIL_RETENTION_DAYS", 30),
		EmailMaxMessagesPerAccount: parseEnvInt("EMAIL_MAX_MESSAGES_PER_ACCOUNT", 100),
		YtDlpPath:                  os.Getenv("YTDLP_PATH"),
		YtDlpCheckHours:            parseEnvInt("YTDLP_CHECK_HOURS", 24),
	}
}

// StepTimeout is the hard ceiling a single worker step may run before it is
// abandoned, derived from WORKER_STEP_TIMEOUT_SECONDS (default 120s). It is the
// backstop for external and AI calls whose own timeout may be misconfigured —
// e.g. an AI provider HTTP timeout set to 0 (infinite) at runtime — so a single
// stuck step can never freeze a worker slot forever.
func StepTimeout() time.Duration {
	timeout := time.Duration(AppConfig.WorkerStepTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return timeout
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
