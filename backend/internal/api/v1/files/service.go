package files

import (
	"context"
	"fmt"
	"nas-go/api/pkg/utils"
)

type Service struct {
	repository *Repository
	tasks      chan utils.Task
}

func NewService(repository *Repository, tasksChannel chan utils.Task) *Service {
	return &Service{repository: repository, tasks: tasksChannel}
}

func (s *Service) GetFiles(filter FileFilter, fileDtoList *utils.PaginationResponse[FileDto]) error {

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

	error := fileModel.getCheckSumFromFile()
	if error != nil {
		fmt.Println(error)
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

func (service *Service) UpdateFile(file FileDto) (bool, error) {
	ctx := context.Background()
	transaction, err := service.repository.dbContext.BeginTx(ctx, nil)

	defer transaction.Rollback()

	if err != nil {
		return false, err
	}
	result, err := service.repository.UpdateFile(transaction, file.ToModel())

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
	s.tasks <- task
}

func (s *Service) ScanDirTask(data string) {
	task := utils.Task{
		Type: utils.ScanDir,
		Data: data,
	}
	s.tasks <- task
}
