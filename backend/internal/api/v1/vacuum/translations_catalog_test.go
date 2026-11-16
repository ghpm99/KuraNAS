package vacuum

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestErrorMessageKeysExistInBothCatalogs(t *testing.T) {
	for _, locale := range []string{"pt-BR", "en-US"} {
		t.Run(locale, func(t *testing.T) {
			catalogBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "translations", locale+".json"))
			if err != nil {
				t.Fatal(err)
			}
			var catalog map[string]string
			if err := json.Unmarshal(catalogBytes, &catalog); err != nil {
				t.Fatal(err)
			}

			requiredKeys := []string{"ERROR_VACUUM_FAILED"}
			for _, rule := range errorResponseRules {
				requiredKeys = append(requiredKeys, rule.messageKey)
			}
			for _, key := range requiredKeys {
				if catalog[key] == "" {
					t.Errorf("%s is missing from %s", key, locale)
				}
			}
		})
	}
}
