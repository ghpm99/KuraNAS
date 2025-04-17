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

func (s *Service) GetFiles(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {

	paginationResponse := utils.PaginationResponse[FileDto]{
		Items: []FileDto{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	if filter.FileParent == 0 {
		filter.Path = config.AppConfig.EntryPoint
	} else {
		path, error := s.Repository.GetPathByFileId(filter.FileParent)
		if error != nil {
			return paginationResponse, error
		}
		filter.Path = path
	}

	filesModel, err := s.Repository.GetFiles(filter, page, pageSize)
	if err != nil {
		return paginationResponse, err
	}

	for _, imageModel := range filesModel.Items {
		fileDtoResult, err := imageModel.ToDto()

		if err != nil {
			continue
		}
		paginationResponse.Items = append(paginationResponse.Items, fileDtoResult)
	}
	paginationResponse.Pagination = filesModel.Pagination

	return paginationResponse, nil

}

func (s *Service) GetFilesByPath(path string) ([]FileDto, error) {

	filesModel, err := s.Repository.GetFilesByPath(path)
	if err != nil {
		return nil, err
	}

	var fileDtoList []FileDto
	for _, fileModel := range filesModel {
		fileDtoResult, err := fileModel.ToDto()

		if err != nil {
			continue
		}

		fileDtoList = append(fileDtoList, fileDtoResult)
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

	fileDtoResult, err := fileModel.ToDto()

	if err != nil {
		return fileDtoResult, err
	}

	return fileDtoResult, nil
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
