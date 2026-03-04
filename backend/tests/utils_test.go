package tests

import "testing"

func TestConfigInMemoryDatabase(t *testing.T) {
	db := ConfigInMemoryDatabase()
	if db == nil {
		t.Fatalf("expected non-nil database")
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("expected ping success, got %v", err)
	}
	_ = db.Close()
}
