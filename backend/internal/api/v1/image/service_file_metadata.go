package image

type ImageSummaryServiceInterface interface {
	GetImageSummary(fileID int) (ImageSummaryDto, error)
}

type ImageSummaryService struct {
	repository ImageSummaryRepositoryInterface
}

func NewImageSummaryService(repository ImageSummaryRepositoryInterface) *ImageSummaryService {
	return &ImageSummaryService{repository: repository}
}

func (s *ImageSummaryService) GetImageSummary(fileID int) (ImageSummaryDto, error) {
	return s.repository.GetImageSummaryByFileID(fileID)
}
