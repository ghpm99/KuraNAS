package files

import "nas-go/api/pkg/utils"

func (s *Service) GetStarredFiles(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	models, err := s.Repository.GetStarredFiles(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}
	return s.toDtoPageWithCounts(models)
}

func (s *Service) GetRecentlyAccessedFiles(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	models, err := s.Repository.GetRecentlyAccessedFiles(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}
	return s.toDtoPageWithCounts(models)
}
