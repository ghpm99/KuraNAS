package files

type RecentFileService struct {
	Repo RecentFileRepositoryInterface
}

func NewRecentFileService(repo RecentFileRepositoryInterface) *RecentFileService {
	return &RecentFileService{Repo: repo}
}

var DEFAULT_LIMIT = 10

// Registra acesso e mantém só os N mais recentes
func (s *RecentFileService) RegisterAccess(ip string, fileID int, keep int) error {
	if err := s.Repo.Upsert(ip, fileID); err != nil {
		return err
	}
	return s.Repo.DeleteOld(ip, keep)
}

func (s *RecentFileService) GetRecentFiles(page int, limit int) ([]RecentFileDto, error) {
	if limit <= 0 {
		limit = DEFAULT_LIMIT
	}

	if page < 1 {
		page = 1
	}
	recentAccess, err := s.Repo.GetRecentFiles(page, limit)
	if err != nil {
		return nil, err
	}
	var result []RecentFileDto
	for _, access := range recentAccess {
		dto := access.ToDto()
		result = append(result, dto)
	}
	return result, nil
}

func (s *RecentFileService) DeleteRecentFile(ip string, fileID int) error {
	return s.Repo.Delete(ip, fileID)
}

func (s *RecentFileService) GetRecentAccessByFileID(fileID int) ([]RecentFileDto, error) {
	recentAccess, err := s.Repo.GetByFileID(fileID)
	if err != nil {
		return nil, err
	}
	var result []RecentFileDto
	for _, access := range recentAccess {
		dto := access.ToDto()
		result = append(result, dto)
	}
	return result, nil
}
