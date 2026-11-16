package utils

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var updateGoldenFiles = flag.Bool("update", false, "rewrite golden files")

type envelopeItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func buildEnvelopeItems(count int) []envelopeItem {
	names := []string{"alpha", "beta", "gamma"}
	items := make([]envelopeItem, 0, count)
	for index := 0; index < count; index++ {
		items = append(items, envelopeItem{ID: index + 1, Name: names[index]})
	}
	return items
}

func TestPaginationEnvelopeMatchesGoldenFiles(t *testing.T) {
	scenarios := []struct {
		goldenFileName string
		page           int
		pageSize       int
		loadedItems    int
	}{
		{"pagination_page_with_next.json", 1, 2, 3},
		{"pagination_last_page.json", 2, 2, 1},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.goldenFileName, func(t *testing.T) {
			response := PaginationResponse[envelopeItem]{
				Items:      buildEnvelopeItems(scenario.loadedItems),
				Pagination: Pagination{Page: scenario.page, PageSize: scenario.pageSize},
			}
			response.UpdatePagination()

			serialized, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				t.Fatalf("marshal envelope: %v", err)
			}
			serialized = append(serialized, '\n')

			goldenPath := filepath.Join("testdata", scenario.goldenFileName)
			if *updateGoldenFiles {
				if err := os.MkdirAll("testdata", 0o755); err != nil {
					t.Fatalf("create testdata dir: %v", err)
				}
				if err := os.WriteFile(goldenPath, serialized, 0o644); err != nil {
					t.Fatalf("write golden file: %v", err)
				}
			}

			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden file: %v", err)
			}
			if !bytes.Equal(expected, serialized) {
				t.Fatalf("envelope differs from %s\nexpected:\n%s\ngot:\n%s", goldenPath, expected, serialized)
			}
		})
	}
}
