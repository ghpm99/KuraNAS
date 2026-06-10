package image

import (
	"fmt"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
)

// Service is the image-domain business-logic implementation.
type Service struct {
	Repository RepositoryInterface
}

func NewService(repository RepositoryInterface) ServiceInterface {
	return &Service{
		Repository: repository,
	}
}

// GetImages returns a paginated, grouped list of image files with metadata.
func (s *Service) GetImages(page int, pageSize int, groupBy files.ImageGroupBy) (utils.PaginationResponse[files.FileDto], error) {
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
