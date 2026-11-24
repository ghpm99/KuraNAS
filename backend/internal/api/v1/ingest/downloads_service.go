package ingest

import (
	"errors"

	"nas-go/api/pkg/utils"
)

var ErrInvalidDownloadStateFilter = errors.New("ingest: invalid download state filter")

type downloadLister interface {
	ListDownloads(stateFilter string, page int, pageSize int) ([]DownloadDto, error)
}

type DownloadsService struct {
	repository downloadLister
}

func NewDownloadsService(repository downloadLister) *DownloadsService {
	return &DownloadsService{repository: repository}
}

// ListDownloads returns one page of registered downloads, newest first.
// stateFilter is "", "active" or "finished"; anything else is
// ErrInvalidDownloadStateFilter.
func (s *DownloadsService) ListDownloads(stateFilter string, page int, pageSize int) (utils.PaginationResponse[DownloadDto], error) {
	if !isValidDownloadStateFilter(stateFilter) {
		return utils.PaginationResponse[DownloadDto]{}, ErrInvalidDownloadStateFilter
	}

	downloads, err := s.repository.ListDownloads(stateFilter, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[DownloadDto]{}, err
	}
	if downloads == nil {
		downloads = []DownloadDto{}
	}

	response := utils.PaginationResponse[DownloadDto]{
		Items:      downloads,
		Pagination: utils.Pagination{Page: page, PageSize: pageSize},
	}
	response.UpdatePagination()
	return response, nil
}

func isValidDownloadStateFilter(stateFilter string) bool {
	switch stateFilter {
	case "", DownloadStateFilterActive, DownloadStateFilterFinished:
		return true
	}
	return false
}
