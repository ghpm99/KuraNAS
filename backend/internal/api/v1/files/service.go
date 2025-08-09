package files

import (
	"database/sql"
	"fmt"
	"image"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/icons"
	"nas-go/api/pkg/img"
	"nas-go/api/pkg/utils"
	"os"
	"sort"
	"strings"
)

type Service struct {
	Repository         RepositoryInterface
	MetadataRepository MetadataRepositoryInterface
	Tasks              chan utils.Task
}

func NewService(repository RepositoryInterface, metadataRepository MetadataRepositoryInterface, tasksChannel chan utils.Task) ServiceInterface {
	return &Service{
		Repository:         repository,
		MetadataRepository: metadataRepository,
		Tasks:              tasksChannel,
	}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) (err error) {
	return s.Repository.GetDbContext().ExecTx(fn)
}

func (s *Service) CreateFile(fileDto FileDto) (fileDtoResult FileDto, err error) {

	err = s.withTransaction(func(tx *sql.Tx) (err error) {
		fileModel, err := fileDto.ToModel()
		if err != nil {
			return
		}

		result, err := s.Repository.CreateFile(tx, fileModel)
		if err != nil {
			return
		}

		metadata, err := s.UpsertMetadata(tx, fileDtoResult)
		if err != nil {
			return
		}
		fileDtoResult.Metadata = metadata

		fileDtoResult, err = result.ToDto()
		return
	})
	if err != nil {
		return FileDto{}, fmt.Errorf("error creating file: %w", err)
	}
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

func (service *Service) UpdateFile(fileDto FileDto) (result bool, err error) {
	err = service.withTransaction(func(tx *sql.Tx) (err error) {
		fileModel, err := fileDto.ToModel()
		if err != nil {
			return
		}
		result, err = service.Repository.UpdateFile(tx, fileModel)

		if err != nil {
			return
		}

		if fileDto.Metadata != nil {
			_, err = service.UpsertMetadata(tx, fileDto)
		}

		if err != nil {
			return
		}
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

func (s *Service) UpdateCheckSum(fileId int) error {

	fileDto, err := s.GetFileById(fileId)

	if err != nil {
		return err
	}

	switch fileDto.Type {
	case File:
		return s.updateFileCheckSum(fileDto)
	case Directory:
		return s.updateDirectoryCheckSum(fileDto)
	default:
		return fmt.Errorf("file type not found")
	}

}

func (s *Service) updateFileCheckSum(
	fileDto FileDto,
) error {
	checkSumHash, err := fileDto.GetCheckSumFromFile()

	if err != nil {
		return err
	}

	fileDto.CheckSum = checkSumHash
	result, err := s.UpdateFile(fileDto)

	if err != nil {
		return err
	}

	if !result {
		return fmt.Errorf("erro ao atualizar arquivo: %v\n", err)
	}

	return nil

}

func (s *Service) updateDirectoryCheckSum(fileDto FileDto) error {

	var page = 1
	var hasNext = true
	var checkSumFiles []string

	for hasNext {

		filesInDirectory, err := s.GetFiles(FileFilter{
			ParentPath: utils.Optional[string]{
				Value:    fileDto.Path,
				HasValue: true,
			},
		}, page, 1000)

		if err != nil {
			return err
		}

		for _, file := range filesInDirectory.Items {
			checkSumFiles = append(checkSumFiles, file.CheckSum)
		}
		hasNext = filesInDirectory.Pagination.HasNext

		if hasNext {
			page = filesInDirectory.Pagination.Page + 1
		}
	}

	resultCheckSum := fileDto.GetCheckSumFromPath(checkSumFiles)

	fileDto.CheckSum = resultCheckSum
	result, err := s.UpdateFile(fileDto)

	if err != nil {
		return err
	}

	if !result {
		return fmt.Errorf("nenhum diretorio atualizado")
	}

	return nil
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

func (s *Service) GetTotalSpaceUsed() (int, error) {
	return s.Repository.GetTotalSpaceUsed()
}

func (s *Service) GetTotalFiles() (int, error) {
	return s.Repository.GetCountByType(File)
}
func (s *Service) GetTotalDirectory() (int, error) {
	return s.Repository.GetCountByType(Directory)
}

func (s *Service) GetReportSizeByFormat() ([]SizeReportDto, error) {
	report, err := s.Repository.GetReportSizeByFormat()
	if err != nil {
		return nil, fmt.Errorf("error getting report size by format: %w", err)
	}
	sizeReportMap := make(map[string]SizeReportDto, len(report))

	var totalUsed int64
	for _, item := range report {
		totalUsed += item.Size
		formatType := utils.GetFormatTypeByExtension(item.Format)
		if dto, exists := sizeReportMap[formatType.Type]; exists {
			dto.Total += item.Total
			dto.Size += item.Size
			sizeReportMap[formatType.Type] = dto
		} else {
			sizeReportMap[formatType.Type] = SizeReportDto{
				Format: formatType.Type,
				Total:  item.Total,
				Size:   item.Size,
			}
		}
	}

	sizeReportDtos := make([]SizeReportDto, 0, len(sizeReportMap))

	for typeName, dto := range sizeReportMap {
		dto.Percentage = (float64(dto.Size) / float64(totalUsed)) * 100
		dto.Format = i18n.Translate(typeName)
		sizeReportDtos = append(sizeReportDtos, dto)
	}

	sort.Slice(sizeReportDtos, func(i, j int) bool {
		return sizeReportDtos[i].Size > sizeReportDtos[j].Size
	})

	return sizeReportDtos, nil
}

func (s *Service) GetTopFilesBySize(limit int) ([]FileDto, error) {
	files, err := s.Repository.GetTopFilesBySize(limit)
	if err != nil {
		return nil, fmt.Errorf("error getting top files by size: %w", err)
	}

	fileDtos := make([]FileDto, len(files))
	for i, file := range files {
		fileDto, err := file.ToDto()
		if err != nil {
			return nil, fmt.Errorf("error converting file model to dto: %w", err)
		}
		fileDtos[i] = fileDto
	}

	return fileDtos, nil
}

func (s *Service) GetDuplicateFiles(page int, pageSize int) (DuplicateFileReportDto, error) {
	duplicateFiles, err := s.Repository.GetDuplicateFiles(page, pageSize)
	if err != nil {
		return DuplicateFileReportDto{}, fmt.Errorf("error getting duplicate files: %w", err)
	}

	report := DuplicateFileReportDto{
		Files:      make([]DuplicateFileDto, len(duplicateFiles.Items)),
		Pagination: duplicateFiles.Pagination,
	}

	for i, file := range duplicateFiles.Items {
		report.TotalFiles += file.Copies
		report.TotalSize += file.Size
		report.Files[i] = DuplicateFileDto{
			Name:   file.Name,
			Size:   file.Size,
			Copies: file.Copies,
			Paths:  strings.Split(file.Paths, ","),
		}
	}

	return report, nil
}

func (s *Service) UpsertMetadata(tx *sql.Tx, fileDto FileDto) (FileDto, error) {
	var err error

	// O "type switch" com atribuição de variável 'm'
	// 'm' será a variável que conterá o valor tipado dentro de cada 'case'
	switch m := fileDto.Metadata.(type) {
	case ImageMetadataModel:
		// Agora 'm' já é do tipo ImageMetadataModel, sem a necessidade de conversão
		upsertedMetadata, upsertErr := s.MetadataRepository.UpsertImageMetadata(tx, m)
		if upsertErr != nil {
			err = upsertErr
			break
		}
		fileDto.Metadata = upsertedMetadata

	case AudioMetadataModel:
		// 'm' é do tipo AudioMetadataModel aqui
		upsertedMetadata, upsertErr := s.MetadataRepository.UpsertAudioMetadata(tx, m)
		if upsertErr != nil {
			err = upsertErr
			break
		}
		fileDto.Metadata = upsertedMetadata

	case VideoMetadataModel:
		// 'm' é do tipo VideoMetadataModel aqui
		upsertedMetadata, upsertErr := s.MetadataRepository.UpsertVideoMetadata(tx, m)
		if upsertErr != nil {
			err = upsertErr
			break
		}
		fileDto.Metadata = upsertedMetadata

	default:
		// Caso o Metadata seja nil ou de um tipo não esperado, retorna o DTO original
		// e um erro, se desejar, ou simplesmente 'nil'.
		return fileDto, nil
	}

	return fileDto, err
}
