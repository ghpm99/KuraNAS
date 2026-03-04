package database

import (
	"database/sql"
	"errors"
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

func TestConfigDatabaseReturnsOpenError(t *testing.T) {
	prevConfig := config.AppConfig
	prevOpen := sqlOpenFn
	prevInit := migrationsInitFn
	t.Cleanup(func() {
		config.AppConfig = prevConfig
		sqlOpenFn = prevOpen
		migrationsInitFn = prevInit
	})

	config.AppConfig = config.AppConfigStruct{
		DbHost: "bad-host",
		DbPort: "5432",
		DbUser: "u",
		DbName: "d",
	}
	expectedErr := errors.New("open failed")
	sqlOpenFn = func(driverName, dataSourceName string) (*sql.DB, error) {
		return nil, expectedErr
	}
	migrationsInitFn = func(db *sql.DB) {}

	db, err := ConfigDatabase()
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected open error %v, got %v", expectedErr, err)
	}
	if db != nil {
		t.Fatalf("expected nil db on open error")
	}
}

func TestConfigDatabaseSuccessCallsMigrations(t *testing.T) {
	prevConfig := config.AppConfig
	prevOpen := sqlOpenFn
	prevInit := migrationsInitFn
	t.Cleanup(func() {
		config.AppConfig = prevConfig
		sqlOpenFn = prevOpen
		migrationsInitFn = prevInit
	})

	config.AppConfig = config.AppConfigStruct{
		DbHost: "ok-host",
		DbPort: "5432",
		DbUser: "u",
		DbName: "d",
	}

	expectedDB := &sql.DB{}
	sqlOpenFn = func(driverName, dataSourceName string) (*sql.DB, error) {
		if driverName != "postgres" {
			t.Fatalf("expected postgres driver, got %s", driverName)
		}
		if !strings.Contains(dataSourceName, "host=ok-host") {
			t.Fatalf("unexpected data source name: %s", dataSourceName)
		}
		return expectedDB, nil
	}

	called := false
	migrationsInitFn = func(db *sql.DB) {
		called = true
		if db != expectedDB {
			t.Fatalf("expected migrations to receive same db pointer")
		}
	}

	db, err := ConfigDatabase()
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if db != expectedDB {
		t.Fatalf("expected same db pointer from ConfigDatabase")
	}
	if !called {
		t.Fatalf("expected migrations init to be called")
	}
}
