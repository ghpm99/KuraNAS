package files

import (
	"errors"
	"testing"

	"nas-go/api/pkg/utils"
)

func TestServiceGetTopFilesBySizePageConvertsModelsAndKeepsPagination(t *testing.T) {
	repository := &filesRepoMock{
		getTopFilesBySizePageFn: func(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{
				Items:      []FileModel{{ID: 4, Name: "big.bin", Size: 10}},
				Pagination: utils.Pagination{Page: page, PageSize: pageSize, HasNext: true, HasPrev: true},
			}, nil
		},
	}
	service := &Service{Repository: repository}

	page, err := service.GetTopFilesBySizePage(2, 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != 4 || page.Items[0].Tier != TierHot {
		t.Fatalf("unexpected items: %+v", page.Items)
	}
	if page.Pagination != (utils.Pagination{Page: 2, PageSize: 5, HasNext: true, HasPrev: true}) {
		t.Fatalf("unexpected pagination: %+v", page.Pagination)
	}
}

func TestServiceGetTopFilesBySizePagePropagatesRepositoryError(t *testing.T) {
	repository := &filesRepoMock{
		getTopFilesBySizePageFn: func(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{}, errors.New("boom")
		},
	}
	service := &Service{Repository: repository}

	if _, err := service.GetTopFilesBySizePage(1, 20); err == nil {
		t.Fatal("expected error")
	}
}
