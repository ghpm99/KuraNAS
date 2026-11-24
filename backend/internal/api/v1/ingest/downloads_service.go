package ingest

import (
	"errors"

	"nas-go/api/internal/roots"
	"nas-go/api/pkg/utils"
)

var ErrInvalidDownloadStateFilter = errors.New("ingest: invalid download state filter")

type downloadReader interface {
	ListDownloads(stateFilter string, page int, pageSize int) ([]DownloadDto, error)
	GetDownload(downloadID int) (DownloadDto, error)
}

type downloadFetcher interface {
	Fetch(request FetchRequestDto) (FetchResponseDto, error)
}

type DownloadsService struct {
	repository downloadReader
	fetcher    downloadFetcher
}

func NewDownloadsService(repository downloadReader, fetcher downloadFetcher) *DownloadsService {
	return &DownloadsService{repository: repository, fetcher: fetcher}
}

// RetryDownload enqueues a new download with the url, destination and preset
// of a failed or canceled one and returns the new item. The original stays
// untouched. Errors: ErrDownloadNotFound, ErrDownloadNotRetryable, and the
// same validation errors Fetch returns (e.g. ErrInvalidTarget when the
// destination root is no longer enabled).
func (s *DownloadsService) RetryDownload(downloadID int) (DownloadDto, error) {
	original, err := s.repository.GetDownload(downloadID)
	if err != nil {
		return DownloadDto{}, err
	}
	if !isRetryableDownloadState(original.State) {
		return DownloadDto{}, ErrDownloadNotRetryable
	}
	targetRoot, isTargetEnabled := findEnabledRootByID(original.TargetRootID)
	if !isTargetEnabled {
		return DownloadDto{}, ErrInvalidTarget
	}

	fetched, err := s.fetcher.Fetch(FetchRequestDto{
		URL:        original.URL,
		Preset:     original.Preset,
		TargetRoot: targetRoot.Path,
		Subfolder:  original.Subfolder,
	})
	if err != nil {
		return DownloadDto{}, err
	}
	return s.repository.GetDownload(fetched.DownloadID)
}

func isRetryableDownloadState(state string) bool {
	return state == "failed" || state == "canceled"
}

func findEnabledRootByID(rootID int) (roots.Root, bool) {
	for _, root := range roots.Enabled() {
		if root.ID == rootID {
			return root, true
		}
	}
	return roots.Root{}, false
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
