package files

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/icons"
	"nas-go/api/pkg/img"
	"nas-go/api/pkg/utils"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Service struct {
	Repository     RepositoryInterface
	JobsRepository jobs.RepositoryInterface
	Tasks          chan utils.Task
}

func NewService(
	repository RepositoryInterface,
	jobsRepository jobs.RepositoryInterface,
	tasksChannel chan utils.Task,
) ServiceInterface {
	return &Service{
		Repository:     repository,
		JobsRepository: jobsRepository,
		Tasks:          tasksChannel,
	}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) (err error) {
	return database.ExecOptionalTx(s.Repository.GetDbContext(), fn)
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
		fileDto.ID = result.ID

		fileDtoResult, err = result.ToDto()
		return
	})
	if err != nil {
		return FileDto{}, fmt.Errorf("error creating file: %w", err)
	}
	return
}

// GetFileStatByPath exposes the lightweight per-path lookup used by the diff
// scan to decide whether a file on disk changed since it was last indexed.
func (s *Service) GetFileStatByPath(path string) (FileStat, bool, error) {
	return s.Repository.GetFileStatByPath(path)
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
		return 0
	}

	return contentCount
}

// toDtoPageWithCounts converts a model page to the DTO shape served by the
// listing endpoints, filling DirectoryContentCount for directories.
func (s *Service) toDtoPageWithCounts(models utils.PaginationResponse[FileModel]) (utils.PaginationResponse[FileDto], error) {
	page, err := ParsePaginationToDto(&models)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}
	for index := range page.Items {
		if page.Items[index].Type == Directory {
			page.Items[index].DirectoryContentCount = s.getDirectoryContentCount(page.Items[index])
		}
	}
	return page, nil
}

// GetChildrenByParentPath lists the active children of a directory (the tree),
// optionally narrowed by category (all / starred / recent).
func (s *Service) GetChildrenByParentPath(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	models, err := s.Repository.GetActiveChildrenByParentPath(parentPath, category, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}
	return s.toDtoPageWithCounts(models)
}

// GetFilesByPath returns the active row(s) at an exact path.
func (s *Service) GetFilesByPath(path string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	models, err := s.Repository.GetActiveFilesByPath(path, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}
	return s.toDtoPageWithCounts(models)
}

// GetActiveFilesPage lists all active files, paginated.
func (s *Service) GetActiveFilesPage(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	models, err := s.Repository.GetActiveFiles(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}
	return s.toDtoPageWithCounts(models)
}

// GetFilesByPathPrefix walks a subtree in any soft-delete state (reconciliation
// feed) — no per-directory content counts, the consumer only needs the rows.
func (s *Service) GetFilesByPathPrefix(prefix string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	models, err := s.Repository.GetFilesByPathPrefix(prefix, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}
	return ParsePaginationToDto(&models)
}

func (s *Service) GetFileByNameAndPath(name string, path string) (FileDto, error) {
	// Sees soft-deleted rows too: pickActiveFile prefers the active row, but
	// persist/revive flows need the deleted one when it is all there is.
	models, err := s.Repository.GetFilesByNameAndPath(name, path, 5)
	if err != nil {
		return FileDto{}, fmt.Errorf("error fetching files: %w", err)
	}

	items := make([]FileDto, 0, len(models))
	for index := range models {
		fileDto, dtoErr := models[index].ToDto()
		if dtoErr != nil {
			return FileDto{}, dtoErr
		}
		items = append(items, fileDto)
	}
	return pickActiveFile(items)
}

func (s *Service) GetFileById(id int) (FileDto, error) {
	// Sees soft-deleted rows too: internal flows look rows up by id even
	// while soft-deleted (e.g. restore).
	model, found, err := s.Repository.GetFileByID(id)
	if err != nil {
		return FileDto{}, fmt.Errorf("error fetching file: %w", err)
	}
	if !found {
		return FileDto{}, sql.ErrNoRows
	}
	return model.ToDto()
}

// pickActiveFile resolve buscas que podem casar mais de uma linha. Um arquivo recriado
// no mesmo caminho deixa a linha soft-deleted antiga convivendo com a nova, então a
// busca por name+path retorna 2 registros. Antes isso virava erro e travava os jobs de
// metadata/thumbnail/checksum (deixando arquivos sem miniatura); agora preferimos a
// linha ativa (não deletada) — os itens já chegam ordenados por id DESC, então pegamos
// a mais recente.
func pickActiveFile(items []FileDto) (FileDto, error) {
	switch len(items) {
	case 0:
		return FileDto{}, sql.ErrNoRows
	case 1:
		return items[0], nil
	default:
		for _, f := range items {
			if !f.DeletedAt.HasValue {
				return f, nil
			}
		}
		return items[0], nil
	}
}

func (service *Service) UpdateFile(fileDto FileDto) (result bool, err error) {
	err = service.withTransaction(func(tx *sql.Tx) (err error) {
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
		Data: "File scan",
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

type uploadJobPayload struct {
	FileID int      `json:"file_id,omitempty"`
	Path   string   `json:"path,omitempty"`
	File   *FileDto `json:"file,omitempty"`
}

func (s *Service) CreateUploadProcessJob(paths []string) (int, error) {
	if s.JobsRepository == nil {
		return 0, fmt.Errorf("jobs repository is not configured")
	}
	if len(paths) == 0 {
		return 0, fmt.Errorf("paths are required")
	}

	var createdJob jobs.JobModel
	err := database.ExecOptionalTx(s.JobsRepository.GetDbContext(), func(tx *sql.Tx) error {
		scopeJSON, scopeErr := json.Marshal(map[string]any{"paths": paths})
		if scopeErr != nil {
			return fmt.Errorf("marshal upload scope: %w", scopeErr)
		}

		jobModel, createErr := s.JobsRepository.CreateJob(tx, jobs.JobModel{
			Type:            "upload_process",
			Priority:        "high",
			Scope:           scopeJSON,
			Status:          "queued",
			CancelRequested: false,
		})
		if createErr != nil {
			return createErr
		}
		createdJob = jobModel

		for _, path := range paths {
			fileInfo, statErr := os.Stat(path)
			if statErr != nil {
				return statErr
			}

			fileDto := FileDto{
				Path:       path,
				ParentPath: filepath.Dir(path),
			}
			if parseErr := fileDto.ParseFileInfoToFileDto(fileInfo); parseErr != nil {
				return parseErr
			}

			persistPayload, marshalPersistErr := json.Marshal(uploadJobPayload{Path: path, File: &fileDto})
			if marshalPersistErr != nil {
				return marshalPersistErr
			}

			persistStep, createPersistErr := s.JobsRepository.CreateStep(tx, jobs.StepModel{
				JobID:       createdJob.ID,
				Type:        "persist",
				Status:      "queued",
				DependsOn:   []byte("[]"),
				Attempts:    0,
				MaxAttempts: 3,
				Progress:    0,
				Payload:     persistPayload,
			})
			if createPersistErr != nil {
				return createPersistErr
			}

			commonPayload, marshalCommonErr := json.Marshal(uploadJobPayload{Path: path})
			if marshalCommonErr != nil {
				return marshalCommonErr
			}
			dependsOnPersist, dependsErr := json.Marshal([]int{persistStep.ID})
			if dependsErr != nil {
				return dependsErr
			}

			if _, createStepErr := s.JobsRepository.CreateStep(tx, jobs.StepModel{
				JobID:       createdJob.ID,
				Type:        "metadata",
				Status:      "queued",
				DependsOn:   dependsOnPersist,
				Attempts:    0,
				MaxAttempts: 3,
				Progress:    0,
				Payload:     commonPayload,
			}); createStepErr != nil {
				return createStepErr
			}

			if _, createStepErr := s.JobsRepository.CreateStep(tx, jobs.StepModel{
				JobID:       createdJob.ID,
				Type:        "checksum",
				Status:      "queued",
				DependsOn:   dependsOnPersist,
				Attempts:    0,
				MaxAttempts: 3,
				Progress:    0,
				Payload:     commonPayload,
			}); createStepErr != nil {
				return createStepErr
			}

			formatType := utils.GetFormatTypeByExtension(fileDto.Format)
			if formatType.Type == utils.FormatTypeImage || formatType.Type == utils.FormatTypeVideo {
				thumbnailPayload, thumbnailPayloadErr := json.Marshal(uploadJobPayload{Path: path})
				if thumbnailPayloadErr != nil {
					return thumbnailPayloadErr
				}
				if _, createStepErr := s.JobsRepository.CreateStep(tx, jobs.StepModel{
					JobID:       createdJob.ID,
					Type:        "thumbnail",
					Status:      "queued",
					DependsOn:   dependsOnPersist,
					Attempts:    0,
					MaxAttempts: 3,
					Progress:    0,
					Payload:     thumbnailPayload,
				}); createStepErr != nil {
					return createStepErr
				}
			}

			if formatType.Type == utils.FormatTypeVideo {
				playlistPayload, playlistPayloadErr := json.Marshal(uploadJobPayload{Path: path})
				if playlistPayloadErr != nil {
					return playlistPayloadErr
				}
				if _, createStepErr := s.JobsRepository.CreateStep(tx, jobs.StepModel{
					JobID:       createdJob.ID,
					Type:        "playlist_index",
					Status:      "queued",
					DependsOn:   dependsOnPersist,
					Attempts:    0,
					MaxAttempts: 3,
					Progress:    0,
					Payload:     playlistPayload,
				}); createStepErr != nil {
					return createStepErr
				}
			}

		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return createdJob.ID, nil
}

func (s *Service) UpdateCheckSum(fileId int) error {

	fileDto, err := s.GetFileById(fileId)

	if err != nil {
		return err
	}

	switch fileDto.Type {
	case File, Directory:
		s.Tasks <- utils.Task{
			Type: utils.UpdateCheckSum,
			Data: fileId,
		}
		return nil
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
		return fmt.Errorf("error updating file: %v", err)
	}

	return nil

}

func (s *Service) updateDirectoryCheckSum(fileDto FileDto) error {

	var page = 1
	var hasNext = true
	var checkSumFiles []string

	for hasNext {

		filesInDirectory, err := s.Repository.GetActiveChildrenByParentPath(fileDto.Path, AllCategory, page, 1000)

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
		return fmt.Errorf("no directory updated")
	}

	return nil
}

func (s *Service) GetFileThumbnail(fileDto FileDto, width, height int) ([]byte, error) {
	if width <= 0 {
		width = 320
	}
	if width > 2048 {
		width = 2048
	}

	cacheDir := config.GetBuildConfig("ThumbnailPath")
	cacheKey := fmt.Sprintf("%d_%d.png", fileDto.ID, width)
	cachePath := filepath.Join(cacheDir, cacheKey)

	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}

	var thumbnailImg image.Image

	if fileDto.Type == Directory {
		iconImg, err := icons.FolderIcon()
		if err != nil {
			return nil, err
		}
		thumbnailImg = img.Thumbnail(iconImg, uint(width), uint(height))
	} else {
		exists := s.CheckFileExistsByPath(fileDto.Path)
		if !exists {
			err := s.DeleteFile(fileDto, true)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrDatabase, err)
			}
			return nil, fmt.Errorf("%w: %s", ErrFileMissingDisk, fileDto.Path)
		}

		srcImg, format, err := img.OpenImageFromFile(fileDto.Path)
		if err != nil {
			switch strings.ToLower(fileDto.Format) {
			case ".pdf":
				iconImg, _ := icons.PdfIcon()
				thumbnailImg = img.Thumbnail(iconImg, uint(width), uint(height))
			case ".mp3", ".flac", ".wav", ".ogg", ".m4a":
				iconImg, _ := icons.Mp3Icon()
				thumbnailImg = img.Thumbnail(iconImg, uint(width), uint(height))
			case ".mp4", ".avi", ".mkv", ".mov", ".webm":
				iconImg, _ := icons.Mp4Icon()
				thumbnailImg = img.Thumbnail(iconImg, uint(width), uint(height))
			default:
				iconImg, _ := icons.Icon()
				thumbnailImg = img.Thumbnail(iconImg, uint(width), uint(height))
			}
		} else {
			thumbnailImg = img.Thumbnail(srcImg, uint(width), uint(height))
			_ = format
		}
	}

	data, err := img.EncodePNG(thumbnailImg)
	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	_ = os.MkdirAll(cacheDir, 0755)
	_ = os.WriteFile(cachePath, data, 0644)

	return data, nil
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

func (s *Service) CheckFileExists(fileId int) bool {
	file, err := s.GetFileById(fileId)

	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		return false
	}

	return s.CheckFileExistsByPath(file.Path)
}

func (s *Service) CheckFileExistsByPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func (s *Service) DeleteFile(file FileDto, bySystem bool) error {
	if file.DeletedAt.HasValue {
		return fmt.Errorf("file already marked for deletion")
	}
	if bySystem && (!file.LastInteraction.HasValue || file.LastInteraction.Value.Add(24*time.Hour).After(time.Now())) {
		return fmt.Errorf("file was recently accessed, cannot be deleted")
	}

	if !file.DeletedAt.HasValue {
		file.DeletedAt = utils.Optional[time.Time]{HasValue: true, Value: time.Now()}
	}

	success, err := s.UpdateFile(file)
	if err != nil {
		return fmt.Errorf("error updating file before deletion: %w", err)
	}

	if !success {
		return fmt.Errorf("file not found for deletion")
	}
	return nil
}
