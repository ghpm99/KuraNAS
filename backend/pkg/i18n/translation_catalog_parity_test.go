package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func readCatalogKeys(t *testing.T, locale string) map[string]bool {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "..", "translations", locale+".json"))
	if err != nil {
		t.Fatalf("read %s catalog: %v", locale, err)
	}
	catalog := map[string]string{}
	if err := json.Unmarshal(content, &catalog); err != nil {
		t.Fatalf("decode %s catalog: %v", locale, err)
	}
	keys := make(map[string]bool, len(catalog))
	for key := range catalog {
		keys[key] = true
	}
	return keys
}

func keysMissingFrom(reference, candidate map[string]bool) []string {
	missing := []string{}
	for key := range reference {
		if !candidate[key] {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}

func TestCatalogsHaveTheSameKeys(t *testing.T) {
	portuguese := readCatalogKeys(t, "pt-BR")
	english := readCatalogKeys(t, "en-US")

	if missing := keysMissingFrom(portuguese, english); len(missing) > 0 {
		t.Errorf("keys in pt-BR missing from en-US: %v", missing)
	}
	if missing := keysMissingFrom(english, portuguese); len(missing) > 0 {
		t.Errorf("keys in en-US missing from pt-BR: %v", missing)
	}
}
