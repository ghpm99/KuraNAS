package notifications

import (
	"database/sql"
	"testing"
	"time"
)

func TestToDto(t *testing.T) {
	now := time.Now()

	cases := []struct {
		name            string
		model           NotificationModel
		expectGroupKey  string
		expectMetaNil   bool
		expectMetaKey   string
		expectMetaValue string
	}{
		{
			name: "basic fields without nullable values",
			model: NotificationModel{
				ID: 1, Type: "info", Title: "Test", Message: "Hello",
				IsRead: false, CreatedAt: now, GroupCount: 1, IsGrouped: false,
			},
			expectGroupKey: "", expectMetaNil: true,
		},
		{
			name: "valid group key is extracted",
			model: NotificationModel{
				ID: 2, Type: "success", Title: "Grouped", Message: "msg",
				GroupKey: sql.NullString{String: "files", Valid: true}, IsGrouped: true,
			},
			expectGroupKey: "files", expectMetaNil: true,
		},
		{
			name: "null group key returns empty string",
			model: NotificationModel{
				ID: 7, Type: "info", Title: "No Group", Message: "msg",
				GroupKey: sql.NullString{Valid: false},
			},
			expectGroupKey: "", expectMetaNil: true,
		},
		{
			name: "valid JSON metadata is parsed",
			model: NotificationModel{
				ID: 3, Type: "info", Title: "Meta", Message: "msg",
				Metadata: sql.NullString{String: `{"key":"value"}`, Valid: true},
			},
			expectMetaNil: false, expectMetaKey: "key", expectMetaValue: "value",
		},
		{
			name: "invalid JSON metadata returns nil",
			model: NotificationModel{
				ID: 4, Type: "warning", Title: "Bad Meta", Message: "msg",
				Metadata: sql.NullString{String: "not-json", Valid: true},
			},
			expectMetaNil: true,
		},
		{
			name: "empty metadata string returns nil",
			model: NotificationModel{
				ID: 5, Type: "info", Title: "Empty Meta", Message: "msg",
				Metadata: sql.NullString{String: "", Valid: true},
			},
			expectMetaNil: true,
		},
		{
			name: "null metadata returns nil",
			model: NotificationModel{
				ID: 6, Type: "info", Title: "Null Meta", Message: "msg",
				Metadata: sql.NullString{Valid: false},
			},
			expectMetaNil: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dto := toDto(tc.model)

			if dto.ID != tc.model.ID || dto.Type != tc.model.Type || dto.Title != tc.model.Title || dto.Message != tc.model.Message {
				t.Fatalf("basic fields mismatch: %+v", dto)
			}
			if dto.GroupKey != tc.expectGroupKey {
				t.Fatalf("expected group key %q, got %q", tc.expectGroupKey, dto.GroupKey)
			}
			if tc.expectMetaNil && dto.Metadata != nil {
				t.Fatalf("expected nil metadata, got %v", dto.Metadata)
			}
			if !tc.expectMetaNil {
				meta, ok := dto.Metadata.(map[string]any)
				if !ok {
					t.Fatalf("expected metadata map, got %T", dto.Metadata)
				}
				if meta[tc.expectMetaKey] != tc.expectMetaValue {
					t.Fatalf("expected metadata %s=%s, got %v", tc.expectMetaKey, tc.expectMetaValue, meta[tc.expectMetaKey])
				}
			}
		})
	}
}
