package files

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"nas-go/api/pkg/utils"
	"os"
)

type Service struct {
	repository *Repository
	tasks      chan utils.Task
}

type FileType int

const (
	Directory FileType = 1
	File      FileType = 2
)

type FileData struct {
	Name string
	Type FileType
	Size int64
}

func NewService(repository *Repository, tasksChannel chan utils.Task) *Service {
	return &Service{repository: repository, tasks: tasksChannel}
}

func (s *Service) GetFiles(filter FileFilter, fileDtoList *utils.PaginationResponse[FileDto]) error {

	entries, err := os.ReadDir(filter.Path)
	if err != nil {
		fmt.Printf("❌ Erro ao ler diretório %s: %v\n", filter.Path, err)
		return err
	}

	var fileArray []FileData

	for _, entry := range entries {
		parseDirEntryToFileData(entry, &fileArray)
		name := entry.Name()
		isDir := entry.IsDir()
		fmt.Printf("📄 Nome: %s, é diretório: %v\n", name, isDir)
	}

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

func parseDirEntryToFileData(entry os.DirEntry, fileArray *[]FileData) {
	if entry.IsDir() {
		parseDirToFileData(entry, fileArray)
		return
	}
	parseFileToFileData(entry, fileArray)
}

func parseDirToFileData(entry os.DirEntry, fileArray *[]FileData) {
	fileData := FileData{
		Name: entry.Name(),
		Type: Directory,
	}
	*fileArray = append(*fileArray, fileData)
}

func parseFileToFileData(entry os.DirEntry, fileArray *[]FileData) {
	fileInfo, err := os.Lstat(entry.Name())

	if err != nil {
		fmt.Printf("Erro ao obter informações do arquivo: %v\n", err)
		return
	}

	fileData := FileData{
		Name: entry.Name(),
		Type: File,
		Size: fileInfo.Size(),
	}
	*fileArray = append(*fileArray, fileData)
}

func (s *Service) GetFileByNameAndPath(name string, path string) (FileDto, error) {
	fileModel, err := s.repository.GetFileByNameAndPath(name, path)

	if err != nil {
		return FileDto{}, err
	}

	error := getCheckSumFromFile(fileModel)
	if error != nil {
		fmt.Println(error)
	}

	return fileModel.ToDto(), nil
}

func getCheckSumFromFile(fileModel FileModel) error {
	file, err := os.Open(fileModel.Path)

	if fileModel.Size > (1 * 1024 * 1024 * 1024) {
		return errors.New("arquivo muito grande")
	}
	if err != nil {
		return err
	}

	defer file.Close()

	h := sha256.New()

	if _, err := io.Copy(h, file); err != nil {
		return err
	}

	checkSumBytes := h.Sum(nil)
	checkSumString := fmt.Sprintf("%x", checkSumBytes)

	fmt.Printf("Check sum %s, tamanho %d\n", checkSumString, len(checkSumString))

	return nil
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
