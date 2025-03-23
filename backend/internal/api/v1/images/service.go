package images

import "nas-go/api/pkg/utils"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetAllImages(pagination utils.Pagination) (utils.PaginationResponse[ImageDto], error) {

	paginationResponse := utils.PaginationResponse[ImageDto]{
		Items:      nil,
		Pagination: pagination,
	}

	imagesModel, err := s.repository.GetAllImages(pagination)
	if err != nil {
		return paginationResponse, err
	}
	for _, imageModel := range imagesModel.Items {
		paginationResponse.Items = append(paginationResponse.Items, imageModel.ToDto())
	}
	return paginationResponse, nil

}

func (i *ImageModel) ToDto() ImageDto {
	return ImageDto{
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
