package configuration

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withTempEnvFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if content != "" {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to seed env file: %v", err)
		}
	}

	previous := envFilePathFn
	envFilePathFn = func() string { return path }
	t.Cleanup(func() { envFilePathFn = previous })
	return path
}

func findField(fields []EnvFieldDto, key string) (EnvFieldDto, bool) {
	for _, field := range fields {
		if field.Key == key {
			return field, true
		}
	}
	return EnvFieldDto{}, false
}

func TestGetEnvConfigMissingFileUsesDefaults(t *testing.T) {
	withTempEnvFile(t, "")
	service := &Service{}

	config, err := service.GetEnvConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(config.Fields) != len(envCatalog) {
		t.Fatalf("expected %d fields, got %d", len(envCatalog), len(config.Fields))
	}

	workers, _ := findField(config.Fields, "WORKER_CONCURRENCY_CHECKSUM")
	if workers.Value != "3" {
		t.Fatalf("expected default 3 for checksum concurrency, got %q", workers.Value)
	}
	secret, _ := findField(config.Fields, "DB_PASSWORD")
	if secret.Configured {
		t.Fatalf("expected secret not configured on empty file")
	}
	if secret.Value != "" {
		t.Fatalf("secret value must never be exposed, got %q", secret.Value)
	}
}

func TestGetEnvConfigReadsFileAndMasksSecrets(t *testing.T) {
	withTempEnvFile(t, "LANGUAGE=\"en-US\"\nDB_PASSWORD=\"hunter2\"\n")
	service := &Service{}

	config, err := service.GetEnvConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lang, _ := findField(config.Fields, "LANGUAGE")
	if lang.Value != "en-US" {
		t.Fatalf("expected language en-US, got %q", lang.Value)
	}
	secret, _ := findField(config.Fields, "DB_PASSWORD")
	if !secret.Configured {
		t.Fatalf("expected DB_PASSWORD to read as configured")
	}
	if secret.Value != "" {
		t.Fatalf("secret value leaked: %q", secret.Value)
	}
}

func TestUpdateEnvConfigWritesAndBacksUp(t *testing.T) {
	path := withTempEnvFile(t, "LANGUAGE=\"pt-BR\"\n")
	service := &Service{}

	config, err := service.UpdateEnvConfig(UpdateEnvConfigRequest{
		Changes: map[string]string{"LANGUAGE": "en-US", "ENABLE_WORKERS": "true"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !config.RestartRequired {
		t.Fatalf("expected restart_required to be true after write")
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written env: %v", err)
	}
	if !strings.Contains(string(written), "en-US") {
		t.Fatalf("expected new language persisted, got %s", written)
	}

	entries, _ := os.ReadDir(filepath.Dir(path))
	backupFound := false
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".env.") && strings.HasSuffix(entry.Name(), ".bak") {
			backupFound = true
		}
	}
	if !backupFound {
		t.Fatalf("expected a .env backup to be created")
	}
}

func TestUpdateEnvConfigRejectsUnknownKey(t *testing.T) {
	withTempEnvFile(t, "")
	service := &Service{}

	_, err := service.UpdateEnvConfig(UpdateEnvConfigRequest{
		Changes: map[string]string{"TOTALLY_UNKNOWN": "x"},
	})
	if !errors.Is(err, ErrInvalidEnvKey) {
		t.Fatalf("expected ErrInvalidEnvKey, got %v", err)
	}
}

func TestUpdateEnvConfigValidatesValues(t *testing.T) {
	withTempEnvFile(t, "")
	service := &Service{}

	cases := map[string]map[string]string{
		"bad int":     {"LOG_MAX_SIZE_MB": "0"},
		"bad bool":    {"ENABLE_WORKERS": "yes"},
		"bad origins": {"ALLOWED_ORIGINS": "ftp://nope"},
		"bad token":   {"EMAIL_TOKEN_KEY": "not-base64-32"},
	}
	for name, changes := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := service.UpdateEnvConfig(UpdateEnvConfigRequest{Changes: changes, Confirmed: true})
			if !errors.Is(err, ErrInvalidEnvValue) {
				t.Fatalf("expected ErrInvalidEnvValue, got %v", err)
			}
		})
	}
}

func TestUpdateEnvConfigDangerousRequiresConfirmation(t *testing.T) {
	withTempEnvFile(t, "")
	service := &Service{}

	_, err := service.UpdateEnvConfig(UpdateEnvConfigRequest{
		Changes: map[string]string{"DB_HOST": "db.local"},
	})
	if !errors.Is(err, ErrEnvConfirmationRequired) {
		t.Fatalf("expected ErrEnvConfirmationRequired, got %v", err)
	}
}

func TestUpdateEnvConfigEmptySecretIsKept(t *testing.T) {
	withTempEnvFile(t, "DB_PASSWORD=\"existing\"\n")
	service := &Service{}

	config, err := service.UpdateEnvConfig(UpdateEnvConfigRequest{
		Changes: map[string]string{"DB_PASSWORD": ""},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.RestartRequired {
		t.Fatalf("blank secret must be a no-op, not a pending restart")
	}
	secret, _ := findField(config.Fields, "DB_PASSWORD")
	if !secret.Configured {
		t.Fatalf("existing secret must remain configured")
	}
}

func TestUpdateEnvConfigAcceptsValidTokenKey(t *testing.T) {
	withTempEnvFile(t, "")
	service := &Service{}
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))

	config, err := service.UpdateEnvConfig(UpdateEnvConfigRequest{
		Changes:   map[string]string{"EMAIL_TOKEN_KEY": key},
		Confirmed: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	secret, _ := findField(config.Fields, "EMAIL_TOKEN_KEY")
	if !secret.Configured {
		t.Fatalf("expected token key to be stored")
	}
}

func TestTestDatabaseConnection(t *testing.T) {
	withTempEnvFile(t, "DB_PASSWORD=\"stored\"\n")
	service := &Service{}

	if err := service.TestDatabaseConnection(TestDatabaseRequest{}); !errors.Is(err, ErrInvalidEnvValue) {
		t.Fatalf("expected validation error on empty fields, got %v", err)
	}

	var captured string
	previous := pingDatabaseFn
	pingDatabaseFn = func(dsn string) error {
		captured = dsn
		return nil
	}
	t.Cleanup(func() { pingDatabaseFn = previous })

	err := service.TestDatabaseConnection(TestDatabaseRequest{
		Host: "localhost", Port: "5432", User: "postgres", Name: "kuranas",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(captured, "password=stored") {
		t.Fatalf("expected stored password reused in DSN, got %q", captured)
	}

	pingDatabaseFn = func(dsn string) error { return errors.New("refused") }
	if err := service.TestDatabaseConnection(TestDatabaseRequest{
		Host: "h", Port: "1", User: "u", Name: "n", Password: "p",
	}); err == nil {
		t.Fatalf("expected ping error to propagate")
	}
}

func TestTestPath(t *testing.T) {
	service := &Service{}
	dir := t.TempDir()

	if err := service.TestPath(TestPathRequest{Path: dir}); err != nil {
		t.Fatalf("expected existing dir to pass, got %v", err)
	}
	if err := service.TestPath(TestPathRequest{Path: ""}); !errors.Is(err, ErrInvalidEnvValue) {
		t.Fatalf("expected empty path error, got %v", err)
	}
	if err := service.TestPath(TestPathRequest{Path: "relative/path"}); !errors.Is(err, ErrInvalidEnvValue) {
		t.Fatalf("expected absolute path error, got %v", err)
	}
	if err := service.TestPath(TestPathRequest{Path: filepath.Join(dir, "missing")}); !errors.Is(err, ErrInvalidEnvValue) {
		t.Fatalf("expected missing path error, got %v", err)
	}

	file := filepath.Join(dir, "file.txt")
	_ = os.WriteFile(file, []byte("x"), 0o600)
	if err := service.TestPath(TestPathRequest{Path: file}); !errors.Is(err, ErrInvalidEnvValue) {
		t.Fatalf("expected non-directory error, got %v", err)
	}
}
