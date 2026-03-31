package captures

import (
	"nas-go/api/pkg/utils"
	"testing"
	"time"
)

func TestCaptureModelToDto(t *testing.T) {
	now := time.Now()
	model := CaptureModel{
		ID:        1,
		Name:      "test",
		FileName:  "video.ts",
		FilePath:  "/data/capturas/test/video.ts",
		MediaType: "hls",
		MimeType:  "video/mp2t",
		Size:      1024,
		CreatedAt: now,
	}

	dto := model.ToDto()

	if dto.ID != 1 {
		t.Fatalf("expected ID 1, got %d", dto.ID)
	}
	if dto.Name != "test" {
		t.Fatalf("expected Name test, got %s", dto.Name)
	}
	if dto.FileName != "video.ts" {
		t.Fatalf("expected FileName video.ts, got %s", dto.FileName)
	}
	if dto.FilePath != "/data/capturas/test/video.ts" {
		t.Fatalf("expected correct FilePath, got %s", dto.FilePath)
	}
	if dto.MediaType != "hls" {
		t.Fatalf("expected MediaType hls, got %s", dto.MediaType)
	}
	if dto.MimeType != "video/mp2t" {
		t.Fatalf("expected MimeType video/mp2t, got %s", dto.MimeType)
	}
	if dto.Size != 1024 {
		t.Fatalf("expected Size 1024, got %d", dto.Size)
	}
	if !dto.CreatedAt.Equal(now) {
		t.Fatalf("expected CreatedAt to match")
	}
}

func TestParsePaginationToDto(t *testing.T) {
	now := time.Now()
	pagination := &utils.PaginationResponse[CaptureModel]{
		Items: []CaptureModel{
			{ID: 1, Name: "a", CreatedAt: now},
			{ID: 2, Name: "b", CreatedAt: now},
		},
		Pagination: utils.Pagination{
			Page:     1,
			PageSize: 10,
			HasNext:  true,
			HasPrev:  false,
		},
	}

	result := ParsePaginationToDto(pagination)

	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].Name != "a" {
		t.Fatalf("expected first item name a, got %s", result.Items[0].Name)
	}
	if result.Items[1].Name != "b" {
		t.Fatalf("expected second item name b, got %s", result.Items[1].Name)
	}
}

func TestParsePaginationToDtoEmpty(t *testing.T) {
	pagination := &utils.PaginationResponse[CaptureModel]{
		Items: []CaptureModel{},
		Pagination: utils.Pagination{
			Page:     1,
			PageSize: 10,
		},
	}

	result := ParsePaginationToDto(pagination)

	if len(result.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(result.Items))
	}
}
