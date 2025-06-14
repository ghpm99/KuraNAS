package files

import (
	"context"
	"database/sql"
	"fmt"
	"image"
	"nas-go/api/pkg/icons"
	"nas-go/api/pkg/img"
	"nas-go/api/pkg/utils"
	"os"
	"strconv"
)

type Service struct {
	Repository RepositoryInterface
	Tasks      chan utils.Task
}

func NewService(repository RepositoryInterface, tasksChannel chan utils.Task) ServiceInterface {
	return &Service{Repository: repository, Tasks: tasksChannel}
}

func (s *Service) CreateFile(fileDto FileDto) (fileDtoResult FileDto, err error) {

	err = s.withTransaction(context.Background(), func(tx *sql.Tx) (err error) {
		fileModel, err := fileDto.ToModel()
		if err != nil {
			return
		}

		result, err := s.Repository.CreateFile(tx, fileModel)
		if err != nil {
			return
		}

		fileDtoResult, err = result.ToDto()
		return
	})

	return
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

	for index := range paginationResponse.Items {
		if paginationResponse.Items[index].Type == Directory {
			paginationResponse.Items[index].DirectoryContentCount = s.getDirectoryContentCount(paginationResponse.Items[index])
		}
	}

	return paginationResponse, nil

}

func (s *Service) getDirectoryContentCount(file FileDto) int {
	contentCount, err := s.Repository.GetDirectoryContentCount(file.ID, file.Path)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	return contentCount
}

func (s *Service) GetFileByNameAndPath(name string, path string) (FileDto, error) {
	filter := FileFilter{
		Name: utils.Optional[string]{HasValue: true, Value: name},
		Path: utils.Optional[string]{HasValue: true, Value: path},
	}
	pagination, err := s.GetFiles(filter, 1, 5)

	if err != nil {
		return FileDto{}, fmt.Errorf("erro ao buscar arquivos: %w", err)
	}
	switch len(pagination.Items) {
	case 0:
		return FileDto{}, sql.ErrNoRows
	case 1:
		return pagination.Items[0], nil
	default:
		return FileDto{}, fmt.Errorf("multiple files found with the same name and path")
	}

}

func (s *Service) GetFileById(id int) (FileDto, error) {
	filter := FileFilter{
		ID: utils.Optional[int]{HasValue: true, Value: id},
	}
	pagination, err := s.GetFiles(filter, 1, 5)

	if err != nil {
		return FileDto{}, fmt.Errorf("erro ao buscar arquivo: %w", err)
	}
	switch len(pagination.Items) {
	case 0:
		return FileDto{}, sql.ErrNoRows
	case 1:
		return pagination.Items[0], nil
	default:
		return FileDto{}, fmt.Errorf("multiple files found with the same name and path")
	}

}

func (s *Service) withTransaction(ctx context.Context, fn func(tx *sql.Tx) error) (err error) {
	tx, err := s.Repository.GetDbContext().BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()

	if err = fn(tx); err != nil {
		return
	}

	return tx.Commit()
}

func (service *Service) UpdateFile(fileDto FileDto) (result bool, err error) {
	err = service.withTransaction(context.Background(), func(tx *sql.Tx) (err error) {
		fileModel, err := fileDto.ToModel()
		if err != nil {
			return
		}
		result, err = service.Repository.UpdateFile(tx, fileModel)

		return

	})

	return
}

func (s *Service) ScanFilesTask(data string) {
	task := utils.Task{
		Type: utils.ScanFiles,
		Data: "Escaneamento de arquivos",
	}
	s.Tasks <- task
	s.Tasks <- task
}

func (s *Service) ScanDirTask(data string) {
	task := utils.Task{
		Type: utils.ScanDir,
		Data: data,
	}
	s.Tasks <- task
}

func (s *Service) UpdateCheckSumTask(fileId int) {
	task := utils.Task{
		Type: utils.UpdateCheckSum,
		Data: strconv.Itoa(fileId),
	}
	s.Tasks <- task
}

func (s *Service) GetFileThumbnail(fileDto FileDto, width int) (image.Image, error) {

	if fileDto.Type == Directory {
		return icons.FolderIcon()
	}

	switch fileDto.Format {
	case ".jpg":
		image, err := img.OpenImageFromFile(fileDto.Path, fileDto.Format)
		if err != nil {
			return nil, err
		}
		return img.Thumbnail(image)
	case ".png":
		image, err := img.OpenImageFromFile(fileDto.Path, fileDto.Format)
		if err != nil {
			return nil, err
		}
		return img.Thumbnail(image)
	case ".pdf":
		return icons.PdfIcon()
	case ".mp3":
		return icons.Mp3Icon()
	case ".mp4":
		return icons.Mp4Icon()
	default:
		return icons.Icon()
	}

}

func (s *Service) GetFileBlobById(fileId int) (FileBlob, error) {

	file, err := s.GetFileById(fileId)

	if err != nil {
		return FileBlob{}, err
	}

	data, err := os.ReadFile(file.Path)

	if err != nil {
		return FileBlob{}, err
	}

	return FileBlob{
		ID:     file.ID,
		Blob:   data,
		Format: file.Format,
	}, nil
}
