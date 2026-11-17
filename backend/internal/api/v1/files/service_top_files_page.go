package files

import (
	"fmt"

	"nas-go/api/pkg/utils"
)

func (s *Service) GetTopFilesBySizePage(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	topFilesPage, err := s.Repository.GetTopFilesBySizePage(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, fmt.Errorf("error getting top files by size page: %w", err)
	}
	return ParsePaginationToDto(&topFilesPage)
}
