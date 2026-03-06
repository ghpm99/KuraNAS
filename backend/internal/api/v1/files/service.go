package files

import (
	"bytes"
	"database/sql"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/icons"
	"nas-go/api/pkg/img"
	"nas-go/api/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Service struct {
	Repository         RepositoryInterface
	MetadataRepository MetadataRepositoryInterface
	Tasks              chan utils.Task
	UploadScheduler    UploadProcessSchedulerInterface
}

func NewService(
	repository RepositoryInterface,
	metadataRepository MetadataRepositoryInterface,
	tasksChannel chan utils.Task,
	uploadScheduler UploadProcessSchedulerInterface,
) ServiceInterface {
	return &Service{
		Repository:         repository,
		MetadataRepository: metadataRepository,
		Tasks:              tasksChannel,
		UploadScheduler:    uploadScheduler,
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
		fileDto.ID = result.ID

		metadata, err := s.UpsertMetadata(tx, fileDto)
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
		return FileDto{}, fmt.Errorf("error fetching files: %w", err)
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
		return FileDto{}, fmt.Errorf("error fetching file: %w", err)
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
		Data: "File scan",
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

func (s *Service) ScheduleUploadProcess(uploadedPaths []string) (UploadProcessResult, error) {
	if len(uploadedPaths) == 0 {
		return UploadProcessResult{}, ErrNoUploadedFiles
	}

	if s.UploadScheduler == nil {
		return UploadProcessResult{}, ErrUploadSchedulerUnavailable
	}

	result, err := s.UploadScheduler.ScheduleUploadProcess(uploadedPaths)
	if err != nil {
		return UploadProcessResult{}, fmt.Errorf("schedule upload process: %w", err)
	}

	if result.JobID == "" {
		return UploadProcessResult{}, ErrUploadJobIDMissing
	}

	return result, nil
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

func (s *Service) GetVideoThumbnail(fileDto FileDto, width, height int) ([]byte, error) {
	if width <= 0 {
		width = 320
	}
	if height <= 0 {
		height = 180
	}
	if width > 2048 {
		width = 2048
	}
	if height > 2048 {
		height = 2048
	}

	cacheDir := filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video")
	_ = os.MkdirAll(cacheDir, 0755)
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("%d_%dx%d.png", fileDto.ID, width, height))

	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}

	if !s.CheckFileExistsByPath(fileDto.Path) {
		return nil, fmt.Errorf("%w: %s", ErrFileMissingDisk, fileDto.Path)
	}

	ffmpegErr := exec.Command(
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-ss", "00:00:03",
		"-i", fileDto.Path,
		"-frames:v", "1",
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black", width, height, width, height),
		cachePath,
	).Run()

	if ffmpegErr == nil {
		if data, err := os.ReadFile(cachePath); err == nil {
			return data, nil
		}
	}

	iconImg, _ := icons.Mp4Icon()
	thumb := img.Thumbnail(iconImg, uint(width), uint(height))
	fallback, err := img.EncodePNG(thumb)
	if err != nil {
		return nil, err
	}
	_ = os.WriteFile(cachePath, fallback, 0644)
	return fallback, nil
}

func (s *Service) GetVideoPreviewGif(fileDto FileDto, width, height int) ([]byte, error) {
	if width <= 0 {
		width = 320
	}
	if height <= 0 {
		height = 180
	}
	if width > 1024 {
		width = 1024
	}
	if height > 1024 {
		height = 1024
	}

	cacheDir := filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video")
	_ = os.MkdirAll(cacheDir, 0755)
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("%d_%dx%d_preview.gif", fileDto.ID, width, height))

	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}

	if !s.CheckFileExistsByPath(fileDto.Path) {
		return nil, fmt.Errorf("%w: %s", ErrFileMissingDisk, fileDto.Path)
	}

	// Curta prévia animada: ~2.5s, baixa taxa de frames para performance de cache e rede local.
	ffmpegErr := exec.Command(
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-ss", "00:00:03",
		"-t", "2.5",
		"-i", fileDto.Path,
		"-vf", fmt.Sprintf("fps=4,scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black", width, height, width, height),
		"-loop", "0",
		cachePath,
	).Run()

	if ffmpegErr == nil {
		if data, err := os.ReadFile(cachePath); err == nil {
			return data, nil
		}
	}

	iconImg, _ := icons.Mp4Icon()
	thumb := img.Thumbnail(iconImg, uint(width), uint(height))

	paletted := image.NewPaletted(thumb.Bounds(), palette.Plan9)
	draw.FloydSteinberg.Draw(paletted, thumb.Bounds(), thumb, image.Point{})

	g := &gif.GIF{
		Image:     []*image.Paletted{paletted},
		Delay:     []int{120},
		LoopCount: 0,
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		return nil, err
	}
	fallback := buf.Bytes()
	_ = os.WriteFile(cachePath, fallback, 0644)
	return fallback, nil
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

	switch m := fileDto.Metadata.(type) {
	case ImageMetadataModel:
		m.FileId = fileDto.ID
		upsertedMetadata, upsertErr := s.MetadataRepository.UpsertImageMetadata(tx, m)
		if upsertErr != nil {
			err = upsertErr
			break
		}
		fileDto.Metadata = upsertedMetadata

	case AudioMetadataModel:
		m.FileId = fileDto.ID
		upsertedMetadata, upsertErr := s.MetadataRepository.UpsertAudioMetadata(tx, m)
		if upsertErr != nil {
			err = upsertErr
			break
		}
		fileDto.Metadata = upsertedMetadata

	case VideoMetadataModel:
		m.FileId = fileDto.ID
		upsertedMetadata, upsertErr := s.MetadataRepository.UpsertVideoMetadata(tx, m)
		if upsertErr != nil {
			err = upsertErr
			break
		}
		fileDto.Metadata = upsertedMetadata

	default:
		return fileDto, nil
	}

	return fileDto, err
}

func (s *Service) GetImages(page int, pageSize int, groupBy ImageGroupBy) (utils.PaginationResponse[FileDto], error) {
	filesModel, err := s.Repository.GetImages(page, pageSize, groupBy)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}

	paginationResponse, err := ParsePaginationToDto(&filesModel)

	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}

	return paginationResponse, nil
}

func (s *Service) GetMusic(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	filesModel, err := s.Repository.GetMusic(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}

	paginationResponse, err := ParsePaginationToDto(&filesModel)

	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}

	return paginationResponse, nil
}

func (s *Service) GetVideos(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	filesModel, err := s.Repository.GetVideos(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}

	paginationResponse, err := ParsePaginationToDto(&filesModel)

	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}

	return paginationResponse, nil
}

func (s *Service) GetMusicArtists(page int, pageSize int) (utils.PaginationResponse[MusicArtistDto], error) {
	return s.Repository.GetMusicArtists(page, pageSize)
}

func (s *Service) GetMusicByArtist(artist string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	filesModel, err := s.Repository.GetMusicByArtist(artist, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}
	return ParsePaginationToDto(&filesModel)
}

func (s *Service) GetMusicAlbums(page int, pageSize int) (utils.PaginationResponse[MusicAlbumDto], error) {
	return s.Repository.GetMusicAlbums(page, pageSize)
}

func (s *Service) GetMusicByAlbum(album string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	filesModel, err := s.Repository.GetMusicByAlbum(album, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}
	return ParsePaginationToDto(&filesModel)
}

func (s *Service) GetMusicGenres(page int, pageSize int) (utils.PaginationResponse[MusicGenreDto], error) {
	return s.Repository.GetMusicGenres(page, pageSize)
}

func (s *Service) GetMusicByGenre(genre string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	filesModel, err := s.Repository.GetMusicByGenre(genre, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}
	return ParsePaginationToDto(&filesModel)
}

func (s *Service) GetMusicFolders(page int, pageSize int) (utils.PaginationResponse[MusicFolderDto], error) {
	return s.Repository.GetMusicFolders(page, pageSize)
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
