package logger

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
)

func TestSetExtraData(t *testing.T) {
	var model LoggerModel
	err := model.SetExtraData(LogExtraData{
		Data:  map[string]any{"ok": true},
		Error: "none",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !model.ExtraData.Valid {
		t.Fatalf("expected valid extra data")
	}

	var payload LogExtraData
	if err := json.Unmarshal([]byte(model.ExtraData.String), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload.Error != "none" {
		t.Fatalf("unexpected payload error: %q", payload.Error)
	}
}

func TestSetExtraDataKeepsExistingDataWhenDataNil(t *testing.T) {
	model := LoggerModel{
		ExtraData: sql.NullString{
			Valid:  true,
			String: `{"data":{"id":10},"error":""}`,
		},
	}

	err := model.SetExtraData(LogExtraData{
		Data:  nil,
		Error: "boom",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(model.ExtraData.String, `"id":10`) {
		t.Fatalf("expected previous data to be preserved, got %s", model.ExtraData.String)
	}
	if !strings.Contains(model.ExtraData.String, `"error":"boom"`) {
		t.Fatalf("expected updated error in payload, got %s", model.ExtraData.String)
	}
}

func TestSetExtraDataMarshalError(t *testing.T) {
	var model LoggerModel
	err := model.SetExtraData(LogExtraData{
		Data: make(chan int),
	})
	if err == nil {
		t.Fatalf("expected marshal error")
	}
	if model.ExtraData.Valid {
		t.Fatalf("expected invalid extra data on marshal error")
	}
}
