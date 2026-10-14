package video

type SubtitleServiceInterface interface {
	ListVideoSubtitles(parentPath string, videoName string) ([]SubtitleDto, error)
}

type SubtitleService struct {
	repository SubtitleRepositoryInterface
}

func NewSubtitleService(repository SubtitleRepositoryInterface) *SubtitleService {
	return &SubtitleService{repository: repository}
}

func (s *SubtitleService) ListVideoSubtitles(parentPath string, videoName string) ([]SubtitleDto, error) {
	subtitleFiles, err := s.repository.ListSubtitleFilesByNamePrefix(parentPath, videoName)
	if err != nil {
		return nil, err
	}
	subtitles := []SubtitleDto{}
	for _, subtitleFile := range subtitleFiles {
		nameParts, isMatch := ParseSubtitleName(videoName, subtitleFile.Name)
		if isMatch {
			subtitles = append(subtitles, BuildSubtitleDto(subtitleFile.ID, nameParts))
		}
	}
	return subtitles, nil
}
