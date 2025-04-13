package files

import (
	"context"
	"fmt"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/utils"
)

type Service struct {
	Repository RepositoryInterface
	Tasks      chan utils.Task
}

func NewService(repository RepositoryInterface, tasksChannel chan utils.Task) *Service {
	return &Service{Repository: repository, Tasks: tasksChannel}
}

func (s *Service) GetFiles(filter FileFilter, fileDtoList *utils.PaginationResponse[FileDto]) error {

	if filter.FileParent == 0 {
		filter.Path = config.AppConfig.EntryPoint
	} else {
		path, error := s.Repository.GetPathByFileId(filter.FileParent)
		if error != nil {
			return error
		}
		filter.Path = path
	}

	filesModel, err := s.Repository.GetFiles(filter, fileDtoList.Pagination)
	if err != nil {
		return err
	}

	for _, imageModel := range filesModel.Items {
		fileDtoList.Items = append(fileDtoList.Items, imageModel.ToDto())
	}
	fileDtoList.Pagination = filesModel.Pagination

	return nil

}

func (s *Service) GetFilesByPath(path string) ([]FileDto, error) {

	filesModel, err := s.Repository.GetFilesByPath(path)
	if err != nil {
		return nil, err
	}

	var fileDtoList []FileDto
	for _, fileModel := range filesModel {
		fileDtoList = append(fileDtoList, fileModel.ToDto())
	}

	return fileDtoList, nil
}

func (s *Service) GetFileByNameAndPath(name string, path string) (FileDto, error) {
	fileModel, err := s.Repository.GetFileByNameAndPath(name, path)

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

	transaction, err := s.Repository.GetDbContext().BeginTx(ctx, nil)

	defer transaction.Rollback()

	if err != nil {
		return fileDto, err
	}
	result, err := s.Repository.CreateFile(transaction, fileDto.ToModel())

	if err == nil {
		err = transaction.Commit()
	}

	return result.ToDto(), err
}

func (service *Service) UpdateFile(file FileDto) (bool, error) {
	ctx := context.Background()
	transaction, err := service.Repository.GetDbContext().BeginTx(ctx, nil)

	defer transaction.Rollback()

	if err != nil {
		return false, err
	}
	result, err := service.Repository.UpdateFile(transaction, file.ToModel())

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
