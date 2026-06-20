package image

import (
	"fmt"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
)

// Service is the image-domain business-logic implementation.
type Service struct {
	Repository  RepositoryInterface
	JobEnqueuer JobEnqueuer
}

// NewService wires the image service. jobEnqueuer may be nil when the jobs
// subsystem is not available; the classification-backfill endpoints then report
// the feature as unavailable instead of panicking.
func NewService(repository RepositoryInterface, jobEnqueuer JobEnqueuer) ServiceInterface {
	return &Service{
		Repository:  repository,
		JobEnqueuer: jobEnqueuer,
	}
}

// GetImages returns a paginated, grouped list of image files with metadata.
func (s *Service) GetImages(page int, pageSize int, groupBy ImageGroupBy) (utils.PaginationResponse[files.FileDto], error) {
	filesModel, err := s.Repository.GetImages(page, pageSize, groupBy)
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, fmt.Errorf("GetImages: %w", err)
	}

	paginationResponse, err := files.ParsePaginationToDto(&filesModel)
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, fmt.Errorf("GetImages parse: %w", err)
	}

	return paginationResponse, nil
}
