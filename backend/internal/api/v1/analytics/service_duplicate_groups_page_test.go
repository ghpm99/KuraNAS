package analytics

import (
	"errors"
	"testing"
	"time"

	"nas-go/api/pkg/utils"
)

func TestServiceGetDuplicateGroupsPageMapsModelsAndKeepsPagination(t *testing.T) {
	updatedAt := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	service := &Service{Repository: &repositoryStub{groupsPage: utils.PaginationResponse[DuplicateGroupDetailModel]{
		Items: []DuplicateGroupDetailModel{{
			DuplicateGroupModel: DuplicateGroupModel{Signature: "abc", Copies: 2, ItemSize: 50, ReclaimableSize: 50, Paths: []string{"/a", "/b"}},
			Files: []DuplicateCopyModel{
				{ID: 1, Path: "/a", UpdatedAt: updatedAt, Tier: "hot"},
				{ID: 2, Path: "/b", UpdatedAt: updatedAt, Tier: "cold"},
			},
		}},
		Pagination: utils.Pagination{Page: 3, PageSize: 5, HasNext: true, HasPrev: true},
	}}}

	page, err := service.GetDuplicateGroupsPage(DuplicateCategoryAll, 3, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Pagination != (utils.Pagination{Page: 3, PageSize: 5, HasNext: true, HasPrev: true}) {
		t.Fatalf("unexpected pagination: %+v", page.Pagination)
	}
	group := page.Items[0]
	if group.Signature != "abc" || group.SizeBytes != 50 || group.ReclaimableSize != 50 || group.Copies != 2 {
		t.Fatalf("unexpected group: %+v", group)
	}
	if len(group.Files) != 2 || group.Files[1].Tier != "cold" || group.Files[0].UpdatedAt != "2026-03-04T05:06:07Z" {
		t.Fatalf("unexpected copies: %+v", group.Files)
	}
}

func TestServiceGetDuplicateGroupsPagePropagatesRepositoryError(t *testing.T) {
	service := &Service{Repository: &repositoryStub{err: errors.New("boom")}}
	if _, err := service.GetDuplicateGroupsPage(DuplicateCategoryAll, 1, 20); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseDuplicateCategory(t *testing.T) {
	for _, raw := range []string{"", "image", "audio", "video", "document", "other"} {
		if _, isValid := ParseDuplicateCategory(raw); !isValid {
			t.Fatalf("%q should be valid", raw)
		}
	}
	for _, raw := range []string{"archive", "Image", "pdf"} {
		if _, isValid := ParseDuplicateCategory(raw); isValid {
			t.Fatalf("%q should be invalid", raw)
		}
	}
}
