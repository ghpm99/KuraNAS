package video

type VideoSummaryServiceInterface interface {
	GetVideoSummary(fileID int) (VideoSummaryDto, error)
}

type VideoSummaryService struct {
	repository VideoSummaryRepositoryInterface
}

func NewVideoSummaryService(repository VideoSummaryRepositoryInterface) *VideoSummaryService {
	return &VideoSummaryService{repository: repository}
}

func (s *VideoSummaryService) GetVideoSummary(fileID int) (VideoSummaryDto, error) {
	return s.repository.GetVideoSummaryByFileID(fileID)
}
