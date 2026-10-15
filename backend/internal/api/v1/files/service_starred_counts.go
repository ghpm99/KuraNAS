package files

func (s *Service) CountStarredFilesByKind() (StarredCountsDto, error) {
	return s.Repository.CountStarredFilesByKind()
}
