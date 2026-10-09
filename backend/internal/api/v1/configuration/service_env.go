package configuration

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"nas-go/api/internal/config"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	// ErrInvalidEnvKey is returned when a change targets a key outside the
	// editable catalog (we never write arbitrary keys to the .env file).
	ErrInvalidEnvKey = errors.New("invalid env key")
	// ErrInvalidEnvValue is returned when a value fails its kind/shape check
	// (bad int, non true/false bool, malformed origins, bad token key).
	ErrInvalidEnvValue = errors.New("invalid env value")
	// ErrEnvConfirmationRequired is returned when a dangerous key is changed
	// without the explicit confirmation flag.
	ErrEnvConfirmationRequired = errors.New("confirmation required for dangerous env change")
)

type envKind string

const (
	envKindString envKind = "string"
	envKindInt    envKind = "int"
	envKindBool   envKind = "bool"
	envKindSecret envKind = "secret"
)

const (
	envGroupGeneral  = "general"
	envGroupDatabase = "database"
	envGroupAccess   = "access"
	envGroupEmail    = "email"
	envGroupAI       = "ai"
	envGroupWorkers  = "workers"
)

// envField describes one editable .env variable. Default is the effective value
// the running binary falls back to when the key is absent (mirrors the fallbacks
// in config.InitializeConfig), so the wizard shows what is actually in force.
// Dangerous marks keys that can lock the operator out (DB_*, CORS) or invalidate
// data (EMAIL_TOKEN_KEY) and therefore require explicit confirmation.
type envField struct {
	Key       string
	Group     string
	Kind      envKind
	Default   string
	Dangerous bool
}

var envCatalog = []envField{
	{Key: "LANGUAGE", Group: envGroupGeneral, Kind: envKindString},
	{Key: "ENABLE_WORKERS", Group: envGroupGeneral, Kind: envKindBool, Default: "false"},
	{Key: "WEBDAV_ENABLED", Group: envGroupGeneral, Kind: envKindBool, Default: "false"},
	{Key: "ENTRY_POINT", Group: envGroupGeneral, Kind: envKindString},
	{Key: "ENV", Group: envGroupGeneral, Kind: envKindString},
	{Key: "YTDLP_PATH", Group: envGroupGeneral, Kind: envKindString},
	{Key: "YTDLP_CHECK_HOURS", Group: envGroupGeneral, Kind: envKindInt, Default: "24"},
	{Key: "LOG_LEVEL", Group: envGroupGeneral, Kind: envKindString},
	{Key: "LOG_MAX_SIZE_MB", Group: envGroupGeneral, Kind: envKindInt, Default: "50"},
	{Key: "WATCHER_RECONCILE_HOURS", Group: envGroupGeneral, Kind: envKindInt, Default: "24"},

	{Key: "DB_HOST", Group: envGroupDatabase, Kind: envKindString, Dangerous: true},
	{Key: "DB_PORT", Group: envGroupDatabase, Kind: envKindString, Dangerous: true},
	{Key: "DB_USER", Group: envGroupDatabase, Kind: envKindString, Dangerous: true},
	{Key: "DB_NAME", Group: envGroupDatabase, Kind: envKindString, Dangerous: true},
	{Key: "DB_PASSWORD", Group: envGroupDatabase, Kind: envKindSecret, Dangerous: true},

	{Key: "ALLOWED_ORIGINS", Group: envGroupAccess, Kind: envKindString, Dangerous: true},

	{Key: "EMAIL_GOOGLE_CLIENT_ID", Group: envGroupEmail, Kind: envKindString},
	{Key: "EMAIL_MS_CLIENT_ID", Group: envGroupEmail, Kind: envKindString},
	{Key: "EMAIL_SYNC_INTERVAL_MINUTES", Group: envGroupEmail, Kind: envKindInt, Default: "10"},
	{Key: "EMAIL_RETENTION_DAYS", Group: envGroupEmail, Kind: envKindInt, Default: "30"},
	{Key: "EMAIL_MAX_MESSAGES_PER_ACCOUNT", Group: envGroupEmail, Kind: envKindInt, Default: "100"},
	{Key: "EMAIL_TOKEN_KEY", Group: envGroupEmail, Kind: envKindSecret, Dangerous: true},
	{Key: "EMAIL_GOOGLE_CLIENT_SECRET", Group: envGroupEmail, Kind: envKindSecret},

	{Key: "AI_OPENAI_API_KEY", Group: envGroupAI, Kind: envKindSecret},
	{Key: "AI_ANTHROPIC_API_KEY", Group: envGroupAI, Kind: envKindSecret},

	{Key: "WORKER_CONCURRENCY_CHECKSUM", Group: envGroupWorkers, Kind: envKindInt, Default: "3"},
	{Key: "WORKER_CONCURRENCY_METADATA", Group: envGroupWorkers, Kind: envKindInt, Default: "3"},
	{Key: "WORKER_CONCURRENCY_THUMBNAIL", Group: envGroupWorkers, Kind: envKindInt, Default: "2"},
	{Key: "WORKER_RETRY_BACKOFF_MS", Group: envGroupWorkers, Kind: envKindInt, Default: "500"},
	{Key: "WORKER_SCHEDULER_POLL_MS", Group: envGroupWorkers, Kind: envKindInt, Default: "2000"},
	{Key: "WORKER_MAX_CONCURRENT_JOBS", Group: envGroupWorkers, Kind: envKindInt, Default: "4"},
	{Key: "WORKER_STEP_TIMEOUT_SECONDS", Group: envGroupWorkers, Kind: envKindInt, Default: "120"},
	{Key: "WORKER_HEARTBEAT_SECONDS", Group: envGroupWorkers, Kind: envKindInt, Default: "60"},
}

var envCatalogByKey = func() map[string]envField {
	index := make(map[string]envField, len(envCatalog))
	for _, field := range envCatalog {
		index[field.Key] = field
	}
	return index
}()

// pingDatabaseFn opens a short-lived connection with the candidate DSN and pings
// it. Indirected through a var so tests can validate behavior without a server.
var pingDatabaseFn = pingDatabase

// envFilePathFn resolves the .env location. Indirected so tests can point it at a
// temp file.
var envFilePathFn = func() string {
	return config.GetBuildConfig("EnvFilePath")
}

func (s *Service) GetEnvConfig() (EnvConfigDto, error) {
	values, err := readEnvFile()
	if err != nil {
		return EnvConfigDto{}, err
	}

	return s.buildEnvConfig(values), nil
}

func (s *Service) buildEnvConfig(values map[string]string) EnvConfigDto {
	fields := make([]EnvFieldDto, 0, len(envCatalog))
	for _, field := range envCatalog {
		raw, present := values[field.Key]
		dto := EnvFieldDto{
			Key:       field.Key,
			Group:     field.Group,
			Kind:      string(field.Kind),
			Dangerous: field.Dangerous,
		}
		if field.Kind == envKindSecret {
			dto.Configured = present && strings.TrimSpace(raw) != ""
		} else {
			value := raw
			if !present || value == "" {
				value = field.Default
			}
			dto.Value = value
			dto.Configured = value != ""
		}
		fields = append(fields, dto)
	}

	return EnvConfigDto{
		Fields:          fields,
		RestartRequired: s.envRestartPending(),
	}
}

func (s *Service) UpdateEnvConfig(request UpdateEnvConfigRequest) (EnvConfigDto, error) {
	sanitized, err := sanitizeEnvChanges(request.Changes)
	if err != nil {
		return EnvConfigDto{}, err
	}

	if len(sanitized) == 0 {
		return s.GetEnvConfig()
	}

	if changesIncludeDangerous(sanitized) && !request.Confirmed {
		return EnvConfigDto{}, ErrEnvConfirmationRequired
	}

	current, err := readEnvFile()
	if err != nil {
		return EnvConfigDto{}, err
	}

	for key, value := range sanitized {
		current[key] = value
	}

	if err := writeEnvFile(current); err != nil {
		return EnvConfigDto{}, err
	}

	s.markEnvRestartPending()

	return s.buildEnvConfig(current), nil
}

func (s *Service) TestDatabaseConnection(request TestDatabaseRequest) error {
	host := strings.TrimSpace(request.Host)
	port := strings.TrimSpace(request.Port)
	user := strings.TrimSpace(request.User)
	name := strings.TrimSpace(request.Name)
	if host == "" || port == "" || user == "" || name == "" {
		return fmt.Errorf("%w: database fields are required", ErrInvalidEnvValue)
	}

	password := request.Password
	if password == "" {
		values, err := readEnvFile()
		if err != nil {
			return err
		}
		password = values["DB_PASSWORD"]
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		host, port, user, name, password)

	return pingDatabaseFn(dsn)
}

func (s *Service) TestPath(request TestPathRequest) error {
	trimmed := strings.TrimSpace(request.Path)
	if trimmed == "" {
		return fmt.Errorf("%w: path is required", ErrInvalidEnvValue)
	}

	clean := filepath.Clean(trimmed)
	if !filepath.IsAbs(clean) {
		return fmt.Errorf("%w: path must be absolute", ErrInvalidEnvValue)
	}

	info, err := os.Stat(clean)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEnvValue, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: path is not a directory", ErrInvalidEnvValue)
	}

	return nil
}

// sanitizeEnvChanges drops no-op secret entries, rejects unknown keys and
// validates each value against its catalog kind. Returns the keys to persist.
func sanitizeEnvChanges(changes map[string]string) (map[string]string, error) {
	sanitized := make(map[string]string, len(changes))
	for key, value := range changes {
		field, ok := envCatalogByKey[key]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrInvalidEnvKey, key)
		}

		if field.Kind == envKindSecret {
			// An empty secret means "keep current" — never overwrite with blank.
			if strings.TrimSpace(value) == "" {
				continue
			}
		}

		if err := validateEnvValue(field, value); err != nil {
			return nil, err
		}

		sanitized[key] = value
	}

	return sanitized, nil
}

func validateEnvValue(field envField, value string) error {
	switch field.Kind {
	case envKindInt:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || parsed <= 0 {
			return fmt.Errorf("%w: %s must be a positive integer", ErrInvalidEnvValue, field.Key)
		}
	case envKindBool:
		if value != "true" && value != "false" {
			return fmt.Errorf("%w: %s must be true or false", ErrInvalidEnvValue, field.Key)
		}
	case envKindSecret:
		if field.Key == "EMAIL_TOKEN_KEY" {
			if err := validateTokenKey(value); err != nil {
				return err
			}
		}
	case envKindString:
		if field.Key == "ALLOWED_ORIGINS" {
			if err := validateAllowedOrigins(value); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateTokenKey enforces the EMAIL_TOKEN_KEY contract: base64 decoding to
// exactly 32 bytes (AES-256). A wrong length silently breaks token decryption.
func validateTokenKey(value string) error {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil || len(decoded) != 32 {
		return fmt.Errorf("%w: EMAIL_TOKEN_KEY must be base64 of 32 bytes", ErrInvalidEnvValue)
	}
	return nil
}

func validateAllowedOrigins(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%w: ALLOWED_ORIGINS must not be empty", ErrInvalidEnvValue)
	}
	for _, origin := range strings.Split(trimmed, ",") {
		origin = strings.TrimSpace(origin)
		if origin == "*" {
			continue
		}
		if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			return fmt.Errorf("%w: each origin must start with http:// or https://", ErrInvalidEnvValue)
		}
	}
	return nil
}

func changesIncludeDangerous(changes map[string]string) bool {
	for key := range changes {
		if field, ok := envCatalogByKey[key]; ok && field.Dangerous {
			return true
		}
	}
	return false
}

// readEnvFile parses the current .env into a map. A missing file is not an error
// (first run): it yields an empty map the wizard can populate.
func readEnvFile() (map[string]string, error) {
	path := envFilePathFn()
	values, err := godotenv.Read(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("falha ao ler arquivo .env: %w", err)
	}
	return values, nil
}

// writeEnvFile backs up the existing .env to a timestamped .bak before
// overwriting, so a value that prevents the server from booting can be restored.
func writeEnvFile(values map[string]string) error {
	path := envFilePathFn()
	if err := backupEnvFile(path); err != nil {
		return err
	}
	if err := godotenv.Write(values, path); err != nil {
		return fmt.Errorf("falha ao gravar arquivo .env: %w", err)
	}
	return nil
}

func backupEnvFile(path string) error {
	original, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("falha ao ler .env para backup: %w", err)
	}
	backupPath := fmt.Sprintf("%s.%s.bak", path, time.Now().Format("20060102-150405"))
	if err := os.WriteFile(backupPath, original, 0o600); err != nil {
		return fmt.Errorf("falha ao criar backup do .env: %w", err)
	}
	return nil
}

func (s *Service) markEnvRestartPending() {
	s.envMu.Lock()
	defer s.envMu.Unlock()
	s.envRestartRequired = true
}

func (s *Service) envRestartPending() bool {
	s.envMu.Lock()
	defer s.envMu.Unlock()
	return s.envRestartRequired
}

func pingDatabase(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEnvValue, err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(5 * time.Second)
	if err := db.Ping(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEnvValue, err)
	}
	return nil
}
