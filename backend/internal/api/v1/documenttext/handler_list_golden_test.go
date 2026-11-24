package documenttext

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/pkg/utils"
)

var updateGoldenFiles = flag.Bool("update", false, "rewrite golden files")

func TestListDocumentsHandlerResponseMatchesGoldenFiles(t *testing.T) {
	moment := time.Date(2024, 3, 5, 10, 0, 0, 0, time.UTC)
	buildItems := func(firstID int, count int) []DocumentItemDto {
		items := make([]DocumentItemDto, 0, count)
		for offset := 0; offset < count; offset++ {
			items = append(items, DocumentItemDto{
				ID: firstID + offset, Name: "documento.pdf", Format: ".pdf", ParentPath: "/Casa",
				Size: 1024, ModifiedAt: moment, CreatedAt: moment, Starred: offset == 0, IndexState: StateIndexed,
			})
		}
		return items
	}
	scenarios := []struct {
		goldenFileName string
		target         string
		page           utils.PaginationResponse[DocumentItemDto]
	}{
		{
			"documents_page_with_next.json", "/documents?page=1&page_size=2",
			utils.PaginationResponse[DocumentItemDto]{Items: buildItems(1, 2), Pagination: utils.Pagination{Page: 1, PageSize: 2, HasNext: true}},
		},
		{
			"documents_last_page.json", "/documents?page=2&page_size=2",
			utils.PaginationResponse[DocumentItemDto]{Items: buildItems(3, 1), Pagination: utils.Pagination{Page: 2, PageSize: 2}},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.goldenFileName, func(t *testing.T) {
			recorder := performListRequest(NewHandler(&stubService{documentPage: scenario.page}), scenario.target)
			if recorder.Code != http.StatusOK {
				t.Fatalf("status %d body %s", recorder.Code, recorder.Body.String())
			}
			var indented bytes.Buffer
			if err := json.Indent(&indented, recorder.Body.Bytes(), "", "  "); err != nil {
				t.Fatal(err)
			}
			indented.WriteByte('\n')

			goldenPath := filepath.Join("testdata", scenario.goldenFileName)
			if *updateGoldenFiles {
				if err := os.MkdirAll("testdata", 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(goldenPath, indented.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(expected, indented.Bytes()) {
				t.Fatalf("response differs from %s\nexpected:\n%s\ngot:\n%s", goldenPath, expected, indented.Bytes())
			}
		})
	}
}
