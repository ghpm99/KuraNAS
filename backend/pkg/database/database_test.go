package database

import (
	"strings"
	"testing"

	"nas-go/api/internal/config"
)

func TestApplyDatabaseConfig(t *testing.T) {
	prev := config.AppConfig
	t.Cleanup(func() { config.AppConfig = prev })

	config.AppConfig = config.AppConfigStruct{
		DbHost:     "db.local",
		DbPort:     "5432",
		DbUser:     "user",
		DbPassword: "pass",
		DbName:     "kuranas",
	}

	got := applyDatabaseConfig()
	expectedParts := []string{
		"host=db.local",
		"port=5432",
		"user=user",
		"dbname=kuranas",
		"password=pass",
		"sslmode=disable",
	}
	for _, part := range expectedParts {
		if !strings.Contains(got, part) {
			t.Fatalf("expected connection string to contain %q, got %q", part, got)
		}
	}
}

func TestConfigDatabasePanicsWhenMigrationsCannotStart(t *testing.T) {
	prev := config.AppConfig
	t.Cleanup(func() { config.AppConfig = prev })

	config.AppConfig = config.AppConfigStruct{
		DbHost:     "127.0.0.1",
		DbPort:     "1",
		DbUser:     "x",
		DbPassword: "x",
		DbName:     "x",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic due failed migration transaction")
		}
	}()

	_, _ = ConfigDatabase()
}
