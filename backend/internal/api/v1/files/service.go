package files

import (
	"context"
	"nas-go/api/pkg/utils"
)

type Service struct {
	Repository RepositoryInterface
	Tasks      chan utils.Task
}

func NewService(repository RepositoryInterface, tasksChannel chan utils.Task) *Service {
	return &Service{Repository: repository, Tasks: tasksChannel}
}

func (s *Service) GetFiles(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {

	filesModel, err := s.Repository.GetFiles(filter, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}

	paginationResponse, err := ParsePaginationToDto(&filesModel)

	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}

	return paginationResponse, nil

}

func (s *Service) CreateFile(fileDto FileDto) (FileDto, error) {
	ctx := context.Background()

	transaction, err := s.Repository.GetDbContext().BeginTx(ctx, nil)

	defer transaction.Rollback()

	if err != nil {
		return fileDto, err
	}

	fileModel, err := fileDto.ToModel()

	if err != nil {
		return fileDto, err
	}

	result, err := s.Repository.CreateFile(transaction, fileModel)

	if err == nil {
		err = transaction.Commit()
	}

	fileDtoResult, err := result.ToDto()

	if err != nil {
		return fileDtoResult, err
	}

	return fileDtoResult, nil
}

func (service *Service) UpdateFile(fileDto FileDto) (bool, error) {
	ctx := context.Background()
	transaction, err := service.Repository.GetDbContext().BeginTx(ctx, nil)

	defer transaction.Rollback()

	if err != nil {
		return false, err
	}

	fileModel, err := fileDto.ToModel()

	if err != nil {
		return false, err
	}

	result, err := service.Repository.UpdateFile(transaction, fileModel)

	if result {
		err = transaction.Commit()
	}

	return result, err
}

func (s *Service) ScanFilesTask(data string) {
	task := utils.Task{
		Type: utils.ScanFiles,
		Data: "Escaneamento de arquivos",
	}
	s.Tasks <- task
}

func (s *Service) ScanDirTask(data string) {
	task := utils.Task{
		Type: utils.ScanDir,
		Data: data,
	}
	s.Tasks <- task
}
