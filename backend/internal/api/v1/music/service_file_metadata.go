package music

type AudioSummaryServiceInterface interface {
	GetAudioSummary(fileID int) (AudioSummaryDto, error)
}

type AudioSummaryService struct {
	repository AudioSummaryRepositoryInterface
}

func NewAudioSummaryService(repository AudioSummaryRepositoryInterface) *AudioSummaryService {
	return &AudioSummaryService{repository: repository}
}

func (s *AudioSummaryService) GetAudioSummary(fileID int) (AudioSummaryDto, error) {
	return s.repository.GetAudioSummaryByFileID(fileID)
}
