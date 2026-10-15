package files

import "nas-go/api/pkg/utils"

func (s *Service) ToggleFileStarred(id int) (bool, error) {
	return s.Repository.ToggleFileStarred(id)
}

func (s *Service) SetFileStarred(id int, isStarred bool) (bool, error) {
	return s.Repository.SetFileStarred(id, isStarred)
}

func (s *Service) SetFilesStarred(ids []int, isStarred bool) ([]int, error) {
	return s.Repository.SetFilesStarred(ids, isStarred)
}

func (s *Service) GetStarredFiles(page int, pageSize int, listing StarredListing) (utils.PaginationResponse[FileDto], error) {
	models, err := s.Repository.GetStarredFiles(page, pageSize, listing)
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
