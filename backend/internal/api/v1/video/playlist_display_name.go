package video

import (
	"database/sql"

	"nas-go/api/internal/api/v1/video/playlist"
	"nas-go/api/pkg/i18n"
)

func persistedAutoPlaylistName(group playlist.SmartGroup) string {
	if group.NameKey == "" {
		return group.Name
	}
	return i18n.GetMessage(group.NameKey)
}

func toNullableNameKey(nameKey string) sql.NullString {
	return sql.NullString{String: nameKey, Valid: nameKey != ""}
}
