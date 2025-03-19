package images

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetAllImages() ([]Image, error) {
	return s.repository.GetAllImages()
}
