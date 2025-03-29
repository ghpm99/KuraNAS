package files

import (
	"context"
	"nas-go/api/pkg/utils"
)

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

func (s *Service) GetFileByNameAndPath(name string, path string) (FileDto, error) {
	fileModel, err := s.repository.GetFileByNameAndPath(name, path)

	if err != nil {
		return FileDto{}, err
	}

	return fileModel.ToDto(), nil
}

func (s *Service) CreateFile(fileDto FileDto) (FileDto, error) {
	ctx := context.Background()

	transaction, err := s.repository.dbContext.BeginTx(ctx, nil)

	defer transaction.Rollback()

	if err != nil {
		return fileDto, err
	}
	result, err := s.repository.CreateFile(transaction, fileDto.ToModel())

	if err == nil {
		err = transaction.Commit()
	}

	return result.ToDto(), err
}
