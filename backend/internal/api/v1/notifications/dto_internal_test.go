package notifications

import (
	"database/sql"
	"testing"
	"time"
)

func TestToDtoBasic(t *testing.T) {
	now := time.Now()
	model := NotificationModel{
		ID:         1,
		Type:       "info",
		Title:      "Test",
		Message:    "Hello",
		IsRead:     false,
		CreatedAt:  now,
		GroupCount: 1,
		IsGrouped:  false,
	}

	dto := toDto(model)

	if dto.ID != 1 || dto.Type != "info" || dto.Title != "Test" || dto.Message != "Hello" {
		t.Fatalf("unexpected dto fields: %+v", dto)
	}
	if dto.IsRead != false || dto.GroupCount != 1 || dto.IsGrouped != false {
		t.Fatalf("unexpected dto flags: %+v", dto)
	}
	if dto.GroupKey != "" {
		t.Fatalf("expected empty group key, got %s", dto.GroupKey)
	}
	if dto.Metadata != nil {
		t.Fatalf("expected nil metadata")
	}
}

func TestToDtoWithGroupKey(t *testing.T) {
	model := NotificationModel{
		ID:        2,
		Type:      "success",
		Title:     "Grouped",
		Message:   "msg",
		GroupKey:  sql.NullString{String: "files", Valid: true},
		IsGrouped: true,
	}

	dto := toDto(model)
	if dto.GroupKey != "files" {
		t.Fatalf("expected group key 'files', got '%s'", dto.GroupKey)
	}
}

func TestToDtoWithValidMetadata(t *testing.T) {
	model := NotificationModel{
		ID:       3,
		Type:     "info",
		Title:    "Meta",
		Message:  "msg",
		Metadata: sql.NullString{String: `{"key":"value"}`, Valid: true},
	}

	dto := toDto(model)
	if dto.Metadata == nil {
		t.Fatalf("expected metadata to be parsed")
	}
	meta, ok := dto.Metadata.(map[string]any)
	if !ok {
		t.Fatalf("expected metadata map, got %T", dto.Metadata)
	}
	if meta["key"] != "value" {
		t.Fatalf("expected key=value, got %v", meta["key"])
	}
}

func TestToDtoWithInvalidMetadata(t *testing.T) {
	model := NotificationModel{
		ID:       4,
		Type:     "warning",
		Title:    "Bad Meta",
		Message:  "msg",
		Metadata: sql.NullString{String: "not-json", Valid: true},
	}

	dto := toDto(model)
	if dto.Metadata != nil {
		t.Fatalf("expected nil metadata for invalid JSON, got %v", dto.Metadata)
	}
}

func TestToDtoWithEmptyMetadata(t *testing.T) {
	model := NotificationModel{
		ID:       5,
		Type:     "info",
		Title:    "Empty Meta",
		Message:  "msg",
		Metadata: sql.NullString{String: "", Valid: true},
	}

	dto := toDto(model)
	if dto.Metadata != nil {
		t.Fatalf("expected nil metadata for empty string, got %v", dto.Metadata)
	}
}

func TestToDtoWithNullMetadata(t *testing.T) {
	model := NotificationModel{
		ID:       6,
		Type:     "info",
		Title:    "Null Meta",
		Message:  "msg",
		Metadata: sql.NullString{Valid: false},
	}

	dto := toDto(model)
	if dto.Metadata != nil {
		t.Fatalf("expected nil metadata for null, got %v", dto.Metadata)
	}
}

func TestToDtoWithInvalidGroupKey(t *testing.T) {
	model := NotificationModel{
		ID:       7,
		Type:     "info",
		Title:    "No Group",
		Message:  "msg",
		GroupKey: sql.NullString{Valid: false},
	}

	dto := toDto(model)
	if dto.GroupKey != "" {
		t.Fatalf("expected empty group key for invalid NullString, got '%s'", dto.GroupKey)
	}
}
