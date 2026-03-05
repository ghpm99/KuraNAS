package diary

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/pkg/utils"
)

func TestDiaryParsePaginationToDtoAndDurationBranch(t *testing.T) {
	now := time.Now()
	pagination := utils.PaginationResponse[DiaryModel]{
		Items: []DiaryModel{
			{
				ID:        1,
				Name:      "closed",
				StartTime: now.Add(-2 * time.Hour),
				EndTime:   sql.NullTime{Valid: true, Time: now.Add(-1 * time.Hour)},
			},
			{
				ID:        2,
				Name:      "open",
				StartTime: now.Add(-10 * time.Minute),
				EndTime:   sql.NullTime{Valid: false},
			},
		},
		Pagination: utils.Pagination{Page: 1, PageSize: 10},
	}

	dtoPage, err := ParsePaginationToDto(&pagination)
	if err != nil {
		t.Fatalf("expected ParsePaginationToDto success, got %v", err)
	}
	if len(dtoPage.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(dtoPage.Items))
	}
	if dtoPage.Items[0].Duration <= 0 || dtoPage.Items[1].Duration <= 0 {
		t.Fatalf("expected positive duration for both items")
	}
}
