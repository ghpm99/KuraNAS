package files

import "nas-go/api/pkg/utils"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetFiles(fileDtoList *utils.PaginationResponse[FileDto]) error {

	filesModel, err := s.repository.GetFiles(fileDtoList.Pagination)
	if err != nil {
		return err
	}

	for _, imageModel := range filesModel.Items {
		fileDtoList.Items = append(fileDtoList.Items, imageModel.ToDto())
	}
	fileDtoList.Pagination = filesModel.Pagination

	return nil

}

func (i *FileModel) ToDto() FileDto {
	return FileDto{
		ID:              i.ID,
		Name:            i.Name,
		Path:            i.Path,
		Format:          i.Format,
		Size:            i.Size,
		UpdatedAt:       i.UpdatedAt,
		CreatedAt:       i.CreatedAt,
		LastInteraction: i.LastInteraction,
		LastBackup:      i.LastBackup,
	}
}
