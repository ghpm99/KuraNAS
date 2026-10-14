package video

import (
	"database/sql"

	"nas-go/api/internal/api/v1/video/playlist"
	"nas-go/api/pkg/i18n"
)

func resolvePlaylistName(storedName string, nameKey sql.NullString) string {
	if !nameKey.Valid || nameKey.String == "" {
		return storedName
	}
	return i18n.GetMessage(nameKey.String)
}

func persistedAutoPlaylistName(group playlist.SmartGroup) string {
	if group.NameKey == "" {
		return group.Name
	}
	return i18n.GetMessage(group.NameKey)
}

func toNullableNameKey(nameKey string) sql.NullString {
	return sql.NullString{String: nameKey, Valid: nameKey != ""}
}
