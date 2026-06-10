package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"nas-go/api/pkg/database"
	"nas-go/api/pkg/database/migrations"

	"github.com/lib/pq"
)

func pgEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// NewPostgresDB connects to a local PostgreSQL, ensures a dedicated test
// database named dbName exists, applies the real migrations (so the schema is
// byte-for-byte what production runs), and returns a DbContext ready for
// integration tests that exercise actual SQL — no sqlmock, no fakes.
//
// When no database is reachable it calls t.Skip, so the suite still passes on
// machines/CI without Postgres. Connection settings come from TEST_DB_*/DB_*
// env vars and default to the local cluster (postgres/postgres on 127.0.0.1).
func NewPostgresDB(t *testing.T, dbName string) *database.DbContext {
	t.Helper()

	host := pgEnv("TEST_DB_HOST", pgEnv("DB_HOST", "127.0.0.1"))
	port := pgEnv("TEST_DB_PORT", pgEnv("DB_PORT", "5432"))
	user := pgEnv("TEST_DB_USER", pgEnv("DB_USER", "postgres"))
	pass := pgEnv("TEST_DB_PASSWORD", pgEnv("DB_PASSWORD", "postgres"))

	adminDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable", host, port, user, pass)
	admin, err := sql.Open("postgres", adminDSN)
	if err != nil {
		t.Skipf("postgres unavailable (open admin): %v", err)
	}
	if pingErr := admin.Ping(); pingErr != nil {
		admin.Close()
		t.Skipf("postgres unavailable (ping admin): %v", pingErr)
	}

	var exists bool
	if scanErr := admin.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", dbName).Scan(&exists); scanErr != nil {
		admin.Close()
		t.Skipf("postgres unavailable (lookup db): %v", scanErr)
	}
	if !exists {
		if _, createErr := admin.Exec("CREATE DATABASE " + pq.QuoteIdentifier(dbName)); createErr != nil {
			admin.Close()
			t.Skipf("cannot create test db %q: %v", dbName, createErr)
		}
	}
	admin.Close()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("postgres unavailable (open %q): %v", dbName, err)
	}
	if pingErr := db.Ping(); pingErr != nil {
		db.Close()
		t.Skipf("postgres unavailable (ping %q): %v", dbName, pingErr)
	}

	migrations.Init(db)

	t.Cleanup(func() { db.Close() })
	return database.NewDbContext(db)
}
