package playlist

import (
	"database/sql"

	"nas-go/api/pkg/i18n"
)

func ResolveDisplayName(storedName string, nameKey sql.NullString) string {
	if !nameKey.Valid || nameKey.String == "" {
		return storedName
	}
	return i18n.GetMessage(nameKey.String)
}
