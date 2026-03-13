package files

import (
	"database/sql"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type filesHandlerServiceMock struct{}

func (m *filesHandlerServiceMock) CreateFile(fileDto FileDto) (FileDto, error) { return fileDto, nil }
func (m *filesHandlerServiceMock) GetFileByNameAndPath(name string, path string) (FileDto, error) {
	return FileDto{Name: name, Path: path}, nil
}
func (m *filesHandlerServiceMock) GetFileById(id int) (FileDto, error) {
	return FileDto{
		ID:         id,
		Name:       "file",
		Path:       "/tmp/missing.mp3",
		ParentPath: "/tmp",
		Format:     ".mp3",
		Type:       File,
	}, nil
}
func (m *filesHandlerServiceMock) GetFiles(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	return utils.PaginationResponse[FileDto]{
		Items: []FileDto{{ID: 1, Name: "a", Path: "/tmp/a", ParentPath: "/tmp"}},
		Pagination: utils.Pagination{
			Page: page, PageSize: pageSize,
		},
	}, nil
}
func (m *filesHandlerServiceMock) UpdateFile(file FileDto) (bool, error) { return true, nil }
func (m *filesHandlerServiceMock) ScanFilesTask(data string)             {}
func (m *filesHandlerServiceMock) ScanDirTask(data string)               {}
func (m *filesHandlerServiceMock) UpdateCheckSum(fileId int) error       { return nil }
func (m *filesHandlerServiceMock) CreateUploadProcessJob(paths []string) (int, error) {
	return 1, nil
}
func (m *filesHandlerServiceMock) GetFileThumbnail(fileDto FileDto, width, height int) ([]byte, error) {
	return []byte("png"), nil
}
func (m *filesHandlerServiceMock) GetVideoThumbnail(fileDto FileDto, width, height int) ([]byte, error) {
	return []byte("png"), nil
}
func (m *filesHandlerServiceMock) GetVideoPreviewGif(fileDto FileDto, width, height int) ([]byte, error) {
	return []byte("gif"), nil
}
func (m *filesHandlerServiceMock) GetFileBlobById(fileId int) (FileBlob, error) {
	return FileBlob{ID: fileId, Blob: []byte("data"), Format: ".txt"}, nil
}
func (m *filesHandlerServiceMock) GetTotalSpaceUsed() (int, error) { return 123, nil }
func (m *filesHandlerServiceMock) GetTotalFiles() (int, error)     { return 9, nil }
func (m *filesHandlerServiceMock) GetTotalDirectory() (int, error) { return 3, nil }
func (m *filesHandlerServiceMock) GetReportSizeByFormat() ([]SizeReportDto, error) {
	return []SizeReportDto{{Format: ".txt", Total: 1, Size: 10, Percentage: 100}}, nil
}
func (m *filesHandlerServiceMock) GetTopFilesBySize(limit int) ([]FileDto, error) {
	return []FileDto{{ID: 1, Name: "big"}}, nil
}
func (m *filesHandlerServiceMock) GetDuplicateFiles(page int, pageSize int) (DuplicateFileReportDto, error) {
	return DuplicateFileReportDto{
		Files: []DuplicateFileDto{{Name: "dup", Size: 10, Copies: 2, Paths: []string{"/a", "/b"}}},
		Pagination: utils.Pagination{
			Page: page, PageSize: pageSize,
		},
	}, nil
}
func (m *filesHandlerServiceMock) UpsertMetadata(tx *sql.Tx, file FileDto) (FileDto, error) {
	return file, nil
}
func (m *filesHandlerServiceMock) GetImages(page int, pageSize int, groupBy ImageGroupBy) (utils.PaginationResponse[FileDto], error) {
	return utils.PaginationResponse[FileDto]{Items: []FileDto{{ID: 1}}}, nil
}
func (m *filesHandlerServiceMock) GetMusic(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	return utils.PaginationResponse[FileDto]{Items: []FileDto{{ID: 1}}}, nil
}
func (m *filesHandlerServiceMock) GetVideos(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	return utils.PaginationResponse[FileDto]{Items: []FileDto{{ID: 1}}}, nil
}
func (m *filesHandlerServiceMock) CheckFileExists(fileId int) bool              { return false }
func (m *filesHandlerServiceMock) CheckFileExistsByPath(path string) bool       { return false }
func (m *filesHandlerServiceMock) DeleteFile(file FileDto, bySystem bool) error { return nil }
func (m *filesHandlerServiceMock) GetMusicArtists(page int, pageSize int) (utils.PaginationResponse[MusicArtistDto], error) {
	return utils.PaginationResponse[MusicArtistDto]{Items: []MusicArtistDto{{Artist: "a", TrackCount: 1, AlbumCount: 1}}}, nil
}
func (m *filesHandlerServiceMock) GetMusicByArtist(artist string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	return utils.PaginationResponse[FileDto]{Items: []FileDto{{Name: artist}}}, nil
}
func (m *filesHandlerServiceMock) GetMusicAlbums(page int, pageSize int) (utils.PaginationResponse[MusicAlbumDto], error) {
	return utils.PaginationResponse[MusicAlbumDto]{Items: []MusicAlbumDto{{Album: "x", Artist: "a", Year: "2025", TrackCount: 1}}}, nil
}
func (m *filesHandlerServiceMock) GetMusicByAlbum(album string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	return utils.PaginationResponse[FileDto]{Items: []FileDto{{Name: album}}}, nil
}
func (m *filesHandlerServiceMock) GetMusicGenres(page int, pageSize int) (utils.PaginationResponse[MusicGenreDto], error) {
	return utils.PaginationResponse[MusicGenreDto]{Items: []MusicGenreDto{{Genre: "g", TrackCount: 1}}}, nil
}
func (m *filesHandlerServiceMock) GetMusicByGenre(genre string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	return utils.PaginationResponse[FileDto]{Items: []FileDto{{Name: genre}}}, nil
}
func (m *filesHandlerServiceMock) GetMusicFolders(page int, pageSize int) (utils.PaginationResponse[MusicFolderDto], error) {
	return utils.PaginationResponse[MusicFolderDto]{Items: []MusicFolderDto{{Folder: "/m", TrackCount: 1}}}, nil
}
func (m *filesHandlerServiceMock) UploadFiles(targetPath string, files []*multipart.FileHeader) (UploadFilesResult, error) {
	return UploadFilesResult{}, nil
}
func (m *filesHandlerServiceMock) CreateFolder(parentPath string, name string) (string, error) {
	return "", nil
}
func (m *filesHandlerServiceMock) MovePath(sourcePath string, destinationPath string) (string, error) {
	return destinationPath, nil
}
func (m *filesHandlerServiceMock) DeletePath(path string) error { return nil }
func (m *filesHandlerServiceMock) RenamePath(sourcePath string, newName string) (string, error) {
	return filepath.Join(filepath.Dir(sourcePath), newName), nil
}
func (m *filesHandlerServiceMock) CopyPath(sourcePath string, destinationPath string) (string, error) {
	return destinationPath, nil
}

type filesRecentServiceMock struct{}

func (m *filesRecentServiceMock) RegisterAccess(ip string, fileID int, keep int) error { return nil }
func (m *filesRecentServiceMock) GetRecentFiles(page int, pageSize int) ([]RecentFileDto, error) {
	return []RecentFileDto{{ID: 1, FileID: 1, IPAddress: "127.0.0.1", AccessedAt: time.Now()}}, nil
}
func (m *filesRecentServiceMock) DeleteRecentFile(ip string, fileID int) error { return nil }
func (m *filesRecentServiceMock) GetRecentAccessByFileID(fileID int) ([]RecentFileDto, error) {
	return []RecentFileDto{{ID: 2, FileID: fileID, IPAddress: "127.0.0.1", AccessedAt: time.Now()}}, nil
}

type filesHandlerServiceFuncMock struct {
	filesHandlerServiceMock
	getFilesFn           func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileDto], error)
	getFileByIdFn        func(id int) (FileDto, error)
	updateFileFn         func(file FileDto) (bool, error)
	getFileBlobByIdFn    func(fileId int) (FileBlob, error)
	getTotalSpaceUsedFn  func() (int, error)
	getTotalFilesFn      func() (int, error)
	getTotalDirectoryFn  func() (int, error)
	getReportSizeByFmtFn func() ([]SizeReportDto, error)
	getTopFilesBySizeFn  func(limit int) ([]FileDto, error)
	getDuplicateFilesFn  func(page int, pageSize int) (DuplicateFileReportDto, error)
	getImagesFn          func(page int, pageSize int, groupBy ImageGroupBy) (utils.PaginationResponse[FileDto], error)
	getMusicFn           func(page int, pageSize int) (utils.PaginationResponse[FileDto], error)
	getVideosFn          func(page int, pageSize int) (utils.PaginationResponse[FileDto], error)
	getMusicArtistsFn    func(page int, pageSize int) (utils.PaginationResponse[MusicArtistDto], error)
	getMusicByArtistFn   func(artist string, page int, pageSize int) (utils.PaginationResponse[FileDto], error)
	getMusicAlbumsFn     func(page int, pageSize int) (utils.PaginationResponse[MusicAlbumDto], error)
	getMusicByAlbumFn    func(album string, page int, pageSize int) (utils.PaginationResponse[FileDto], error)
	getMusicGenresFn     func(page int, pageSize int) (utils.PaginationResponse[MusicGenreDto], error)
	getMusicByGenreFn    func(genre string, page int, pageSize int) (utils.PaginationResponse[FileDto], error)
	getMusicFoldersFn    func(page int, pageSize int) (utils.PaginationResponse[MusicFolderDto], error)
}

func (m *filesHandlerServiceFuncMock) GetFiles(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	if m.getFilesFn != nil {
		return m.getFilesFn(filter, page, pageSize)
	}
	return m.filesHandlerServiceMock.GetFiles(filter, page, pageSize)
}
func (m *filesHandlerServiceFuncMock) GetFileById(id int) (FileDto, error) {
	if m.getFileByIdFn != nil {
		return m.getFileByIdFn(id)
	}
	return m.filesHandlerServiceMock.GetFileById(id)
}
func (m *filesHandlerServiceFuncMock) UpdateFile(file FileDto) (bool, error) {
	if m.updateFileFn != nil {
		return m.updateFileFn(file)
	}
	return m.filesHandlerServiceMock.UpdateFile(file)
}
func (m *filesHandlerServiceFuncMock) GetFileBlobById(fileId int) (FileBlob, error) {
	if m.getFileBlobByIdFn != nil {
		return m.getFileBlobByIdFn(fileId)
	}
	return m.filesHandlerServiceMock.GetFileBlobById(fileId)
}
func (m *filesHandlerServiceFuncMock) GetTotalSpaceUsed() (int, error) {
	if m.getTotalSpaceUsedFn != nil {
		return m.getTotalSpaceUsedFn()
	}
	return m.filesHandlerServiceMock.GetTotalSpaceUsed()
}
func (m *filesHandlerServiceFuncMock) GetTotalFiles() (int, error) {
	if m.getTotalFilesFn != nil {
		return m.getTotalFilesFn()
	}
	return m.filesHandlerServiceMock.GetTotalFiles()
}
func (m *filesHandlerServiceFuncMock) GetTotalDirectory() (int, error) {
	if m.getTotalDirectoryFn != nil {
		return m.getTotalDirectoryFn()
	}
	return m.filesHandlerServiceMock.GetTotalDirectory()
}
func (m *filesHandlerServiceFuncMock) GetReportSizeByFormat() ([]SizeReportDto, error) {
	if m.getReportSizeByFmtFn != nil {
		return m.getReportSizeByFmtFn()
	}
	return m.filesHandlerServiceMock.GetReportSizeByFormat()
}
func (m *filesHandlerServiceFuncMock) GetTopFilesBySize(limit int) ([]FileDto, error) {
	if m.getTopFilesBySizeFn != nil {
		return m.getTopFilesBySizeFn(limit)
	}
	return m.filesHandlerServiceMock.GetTopFilesBySize(limit)
}
func (m *filesHandlerServiceFuncMock) GetDuplicateFiles(page int, pageSize int) (DuplicateFileReportDto, error) {
	if m.getDuplicateFilesFn != nil {
		return m.getDuplicateFilesFn(page, pageSize)
	}
	return m.filesHandlerServiceMock.GetDuplicateFiles(page, pageSize)
}
func (m *filesHandlerServiceFuncMock) GetImages(page int, pageSize int, groupBy ImageGroupBy) (utils.PaginationResponse[FileDto], error) {
	if m.getImagesFn != nil {
		return m.getImagesFn(page, pageSize, groupBy)
	}
	return m.filesHandlerServiceMock.GetImages(page, pageSize, groupBy)
}
func (m *filesHandlerServiceFuncMock) GetMusic(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	if m.getMusicFn != nil {
		return m.getMusicFn(page, pageSize)
	}
	return m.filesHandlerServiceMock.GetMusic(page, pageSize)
}
func (m *filesHandlerServiceFuncMock) GetVideos(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	if m.getVideosFn != nil {
		return m.getVideosFn(page, pageSize)
	}
	return m.filesHandlerServiceMock.GetVideos(page, pageSize)
}
func (m *filesHandlerServiceFuncMock) GetMusicArtists(page int, pageSize int) (utils.PaginationResponse[MusicArtistDto], error) {
	if m.getMusicArtistsFn != nil {
		return m.getMusicArtistsFn(page, pageSize)
	}
	return m.filesHandlerServiceMock.GetMusicArtists(page, pageSize)
}
func (m *filesHandlerServiceFuncMock) GetMusicByArtist(artist string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	if m.getMusicByArtistFn != nil {
		return m.getMusicByArtistFn(artist, page, pageSize)
	}
	return m.filesHandlerServiceMock.GetMusicByArtist(artist, page, pageSize)
}
func (m *filesHandlerServiceFuncMock) GetMusicAlbums(page int, pageSize int) (utils.PaginationResponse[MusicAlbumDto], error) {
	if m.getMusicAlbumsFn != nil {
		return m.getMusicAlbumsFn(page, pageSize)
	}
	return m.filesHandlerServiceMock.GetMusicAlbums(page, pageSize)
}
func (m *filesHandlerServiceFuncMock) GetMusicByAlbum(album string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	if m.getMusicByAlbumFn != nil {
		return m.getMusicByAlbumFn(album, page, pageSize)
	}
	return m.filesHandlerServiceMock.GetMusicByAlbum(album, page, pageSize)
}
func (m *filesHandlerServiceFuncMock) GetMusicGenres(page int, pageSize int) (utils.PaginationResponse[MusicGenreDto], error) {
	if m.getMusicGenresFn != nil {
		return m.getMusicGenresFn(page, pageSize)
	}
	return m.filesHandlerServiceMock.GetMusicGenres(page, pageSize)
}
func (m *filesHandlerServiceFuncMock) GetMusicByGenre(genre string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	if m.getMusicByGenreFn != nil {
		return m.getMusicByGenreFn(genre, page, pageSize)
	}
	return m.filesHandlerServiceMock.GetMusicByGenre(genre, page, pageSize)
}
func (m *filesHandlerServiceFuncMock) GetMusicFolders(page int, pageSize int) (utils.PaginationResponse[MusicFolderDto], error) {
	if m.getMusicFoldersFn != nil {
		return m.getMusicFoldersFn(page, pageSize)
	}
	return m.filesHandlerServiceMock.GetMusicFolders(page, pageSize)
}

type filesRecentServiceFuncMock struct {
	filesRecentServiceMock
	getRecentFilesFn  func(page int, pageSize int) ([]RecentFileDto, error)
	getRecentByFileFn func(fileID int) ([]RecentFileDto, error)
}

func (m *filesRecentServiceFuncMock) GetRecentFiles(page int, pageSize int) ([]RecentFileDto, error) {
	if m.getRecentFilesFn != nil {
		return m.getRecentFilesFn(page, pageSize)
	}
	return m.filesRecentServiceMock.GetRecentFiles(page, pageSize)
}
func (m *filesRecentServiceFuncMock) GetRecentAccessByFileID(fileID int) ([]RecentFileDto, error) {
	if m.getRecentByFileFn != nil {
		return m.getRecentByFileFn(fileID)
	}
	return m.filesRecentServiceMock.GetRecentAccessByFileID(fileID)
}

type filesLoggerMock struct{}

func (m *filesLoggerMock) CreateLog(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	return logger.LoggerModel{}, nil
}
func (m *filesLoggerMock) GetLogByID(id int) (logger.LoggerModel, error) {
	return logger.LoggerModel{}, nil
}
func (m *filesLoggerMock) GetLogs(page, pageSize int) ([]logger.LoggerModel, error) {
	return nil, nil
}
func (m *filesLoggerMock) UpdateLog(log logger.LoggerModel) error { return nil }
func (m *filesLoggerMock) CompleteWithSuccessLog(log logger.LoggerModel) error {
	return nil
}
func (m *filesLoggerMock) CompleteWithErrorLog(log logger.LoggerModel, err error) error {
	return nil
}

func newFilesHandlerRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/files", handler.GetFilesHandler)
	router.GET("/files/path", handler.GetFilesByPathHandler)
	router.GET("/files/children/:id", handler.GetChildrenByIdHandler)
	router.POST("/files/update", handler.UpdateFilesHandler)
	router.GET("/files/tree", handler.GetFilesTreeHandler)
	router.GET("/files/thumbnail/:id", handler.GetFileThumbnailHandler)
	router.GET("/files/video-thumbnail/:id", handler.GetVideoThumbnailHandler)
	router.GET("/files/video-preview/:id", handler.GetVideoPreviewHandler)
	router.GET("/files/blob/:id", handler.GetBlobFileHandler)
	router.GET("/files/recent", handler.GetRecentFilesHandler)
	router.GET("/files/recent/:id", handler.GetRecentAccessByFileHandler)
	router.POST("/files/starred/:id", handler.StarreFileHandler)
	router.GET("/files/total-space-used", handler.GetTotalSpaceUsedHandler)
	router.GET("/files/total-files", handler.GetTotalFilesHandler)
	router.GET("/files/total-directory", handler.GetTotalDirectoryHandler)
	router.GET("/files/report-size-by-format", handler.GetReportSizeByFormatHandler)
	router.GET("/files/top-files-by-size", handler.GetTopFilesBySizeHandler)
	router.GET("/files/duplicate-files", handler.GetDuplicateFilesHandler)
	router.GET("/files/images", handler.GetImagesHandler)
	router.GET("/files/music", handler.GetMusicHandler)
	router.GET("/files/videos", handler.GetVideosHandler)
	router.GET("/files/music/artists", handler.GetMusicArtistsHandler)
	router.GET("/files/music/artists/:name", handler.GetMusicByArtistHandler)
	router.GET("/files/music/albums", handler.GetMusicAlbumsHandler)
	router.GET("/files/music/albums/:name", handler.GetMusicByAlbumHandler)
	router.GET("/files/music/genres", handler.GetMusicGenresHandler)
	router.GET("/files/music/genres/:name", handler.GetMusicByGenreHandler)
	router.GET("/files/music/folders", handler.GetMusicFoldersHandler)
	router.GET("/files/stream/:id", handler.StreamAudioHandler)
	router.GET("/files/video-stream/:id", handler.StreamVideoHandler)
	return router
}

func TestFilesHandlerManyEndpoints(t *testing.T) {
	handler := NewHandler(&filesHandlerServiceMock{}, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := newFilesHandlerRouter(handler)

	tests := []struct {
		method string
		path   string
		body   string
		code   int
	}{
		{method: http.MethodGet, path: "/files", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/path?path=/tmp", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/children/1", code: http.StatusOK},
		{method: http.MethodPost, path: "/files/update", body: "data=/tmp", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/tree", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/thumbnail/1", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/video-thumbnail/1", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/video-preview/1", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/blob/1", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/recent", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/recent/1", code: http.StatusOK},
		{method: http.MethodPost, path: "/files/starred/1", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/total-space-used", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/total-files", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/total-directory", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/report-size-by-format", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/top-files-by-size?limit=3", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/duplicate-files", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/images", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/music", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/videos", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/music/artists", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/music/artists/n1", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/music/albums", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/music/albums/a1", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/music/genres", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/music/genres/g1", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/music/folders", code: http.StatusOK},
		{method: http.MethodGet, path: "/files/stream/1", code: http.StatusNotFound},
		{method: http.MethodGet, path: "/files/video-stream/1", code: http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected status %d, got %d. body=%s", tc.code, w.Code, w.Body.String())
			}
		})
	}
}

func TestFilesHandlerUpdateRequiresData(t *testing.T) {
	handler := NewHandler(&filesHandlerServiceMock{}, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.POST("/files/update", handler.UpdateFilesHandler)

	req := httptest.NewRequest(http.MethodPost, "/files/update", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on missing data, got %d", w.Code)
	}
}

type filesStreamServiceMock struct {
	filesHandlerServiceMock
	filePath string
	format   string
}

func (m *filesStreamServiceMock) GetFileById(id int) (FileDto, error) {
	return FileDto{
		ID:         id,
		Name:       "stream",
		Path:       m.filePath,
		ParentPath: filepath.Dir(m.filePath),
		Format:     m.format,
		Type:       File,
	}, nil
}

func (m *filesStreamServiceMock) CheckFileExistsByPath(path string) bool {
	return path == m.filePath
}

func (m *filesStreamServiceMock) GetFileThumbnail(fileDto FileDto, width, height int) ([]byte, error) {
	return nil, ErrFileMissingDisk
}

func (m *filesStreamServiceMock) GetVideoThumbnail(fileDto FileDto, width, height int) ([]byte, error) {
	return nil, ErrFileMissingDisk
}

func (m *filesStreamServiceMock) GetVideoPreviewGif(fileDto FileDto, width, height int) ([]byte, error) {
	return nil, errors.New("preview failed")
}

func TestFilesHandlerStreamsAndErrorBranches(t *testing.T) {
	tmpDir := t.TempDir()
	audioPath := filepath.Join(tmpDir, "a.mp3")
	videoPath := filepath.Join(tmpDir, "v.mp4")
	if err := os.WriteFile(audioPath, []byte("abcdefghijklmnopqrstuvwxyz"), 0644); err != nil {
		t.Fatalf("failed to create audio file: %v", err)
	}
	if err := os.WriteFile(videoPath, []byte("0123456789abcdefghijklmnopqrstuvwxyz"), 0644); err != nil {
		t.Fatalf("failed to create video file: %v", err)
	}

	audioService := &filesStreamServiceMock{filePath: audioPath, format: ".mp3"}
	videoService := &filesStreamServiceMock{filePath: videoPath, format: ".mp4"}

	audioHandler := NewHandler(audioService, &filesRecentServiceMock{}, &filesLoggerMock{})
	videoHandler := NewHandler(videoService, &filesRecentServiceMock{}, &filesLoggerMock{})

	audioRouter := gin.New()
	audioRouter.GET("/files/stream/:id", audioHandler.StreamAudioHandler)
	audioRouter.GET("/files/thumbnail/:id", audioHandler.GetFileThumbnailHandler)

	videoRouter := gin.New()
	videoRouter.GET("/files/video-stream/:id", videoHandler.StreamVideoHandler)
	videoRouter.GET("/files/video-thumbnail/:id", videoHandler.GetVideoThumbnailHandler)
	videoRouter.GET("/files/video-preview/:id", videoHandler.GetVideoPreviewHandler)

	req := httptest.NewRequest(http.MethodGet, "/files/stream/1", nil)
	req.Header.Set("Range", "bytes=0-5")
	w := httptest.NewRecorder()
	audioRouter.ServeHTTP(w, req)
	if w.Code != http.StatusPartialContent {
		t.Fatalf("expected partial content, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/files/video-stream/1", nil)
	req.Header.Set("Range", "bytes=0-10")
	w = httptest.NewRecorder()
	videoRouter.ServeHTTP(w, req)
	if w.Code != http.StatusPartialContent {
		t.Fatalf("expected partial content for video, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/files/thumbnail/1", nil)
	w = httptest.NewRecorder()
	audioRouter.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing thumbnail source, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/files/video-thumbnail/1", nil)
	w = httptest.NewRecorder()
	videoRouter.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing video thumbnail source, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/files/video-preview/1", nil)
	w = httptest.NewRecorder()
	videoRouter.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for video preview generic error, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/files/stream/1", nil)
	w = httptest.NewRecorder()
	audioRouter.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected full audio stream 200 without range, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/files/video-stream/1", nil)
	w = httptest.NewRecorder()
	videoRouter.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected full video stream 200 without range, got %d", w.Code)
	}
}

func TestFilesHandlerGetChildrenByIdNotFound(t *testing.T) {
	service := &filesHandlerServiceFuncMock{
		getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			if filter.ID.HasValue {
				return utils.PaginationResponse[FileDto]{Items: []FileDto{}}, nil
			}
			return utils.PaginationResponse[FileDto]{Items: []FileDto{{ID: 1, Path: "/tmp"}}}, nil
		},
	}
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.GET("/files/children/:id", handler.GetChildrenByIdHandler)

	req := httptest.NewRequest(http.MethodGet, "/files/children/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing parent file, got %d", w.Code)
	}
}

func TestFilesHandlerGetFilesTreeWithParentFilter(t *testing.T) {
	expectedParentPath := "/tmp/parent"
	service := &filesHandlerServiceFuncMock{
		getFileByIdFn: func(id int) (FileDto, error) {
			return FileDto{ID: id, Path: expectedParentPath}, nil
		},
		getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			if !filter.ParentPath.HasValue || filter.ParentPath.Value != expectedParentPath {
				t.Fatalf("expected ParentPath filter %q, got %+v", expectedParentPath, filter.ParentPath)
			}
			return utils.PaginationResponse[FileDto]{Items: []FileDto{}}, nil
		},
	}
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.GET("/files/tree", handler.GetFilesTreeHandler)

	req := httptest.NewRequest(http.MethodGet, "/files/tree?file_parent=123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestFilesHandlerErrorResponses(t *testing.T) {
	errBoom := errors.New("boom")
	service := &filesHandlerServiceFuncMock{
		getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			return utils.PaginationResponse[FileDto]{}, errBoom
		},
		getFileByIdFn: func(id int) (FileDto, error) {
			return FileDto{}, errBoom
		},
		updateFileFn: func(file FileDto) (bool, error) {
			return false, errBoom
		},
		getFileBlobByIdFn: func(fileId int) (FileBlob, error) {
			return FileBlob{}, errBoom
		},
		getTotalSpaceUsedFn: func() (int, error) { return 0, errBoom },
		getTotalFilesFn:     func() (int, error) { return 0, errBoom },
		getTotalDirectoryFn: func() (int, error) { return 0, errBoom },
		getReportSizeByFmtFn: func() ([]SizeReportDto, error) {
			return nil, errBoom
		},
		getTopFilesBySizeFn: func(limit int) ([]FileDto, error) { return nil, errBoom },
		getDuplicateFilesFn: func(page int, pageSize int) (DuplicateFileReportDto, error) {
			return DuplicateFileReportDto{}, errBoom
		},
		getImagesFn: func(page int, pageSize int, groupBy ImageGroupBy) (utils.PaginationResponse[FileDto], error) {
			return utils.PaginationResponse[FileDto]{}, errBoom
		},
		getMusicFn: func(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			return utils.PaginationResponse[FileDto]{}, errBoom
		},
		getVideosFn: func(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			return utils.PaginationResponse[FileDto]{}, errBoom
		},
		getMusicArtistsFn: func(page int, pageSize int) (utils.PaginationResponse[MusicArtistDto], error) {
			return utils.PaginationResponse[MusicArtistDto]{}, errBoom
		},
		getMusicByArtistFn: func(artist string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			return utils.PaginationResponse[FileDto]{}, errBoom
		},
		getMusicAlbumsFn: func(page int, pageSize int) (utils.PaginationResponse[MusicAlbumDto], error) {
			return utils.PaginationResponse[MusicAlbumDto]{}, errBoom
		},
		getMusicByAlbumFn: func(album string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			return utils.PaginationResponse[FileDto]{}, errBoom
		},
		getMusicGenresFn: func(page int, pageSize int) (utils.PaginationResponse[MusicGenreDto], error) {
			return utils.PaginationResponse[MusicGenreDto]{}, errBoom
		},
		getMusicByGenreFn: func(genre string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			return utils.PaginationResponse[FileDto]{}, errBoom
		},
		getMusicFoldersFn: func(page int, pageSize int) (utils.PaginationResponse[MusicFolderDto], error) {
			return utils.PaginationResponse[MusicFolderDto]{}, errBoom
		},
	}
	recentService := &filesRecentServiceFuncMock{
		getRecentFilesFn: func(page int, pageSize int) ([]RecentFileDto, error) {
			return nil, errBoom
		},
		getRecentByFileFn: func(fileID int) ([]RecentFileDto, error) {
			return nil, errBoom
		},
	}
	handler := NewHandler(service, recentService, &filesLoggerMock{})
	router := newFilesHandlerRouter(handler)

	tests := []struct {
		method string
		path   string
		code   int
	}{
		{method: http.MethodGet, path: "/files", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/path?path=/tmp", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/tree", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/children/1", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/blob/1", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/recent", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/recent/1", code: http.StatusInternalServerError},
		{method: http.MethodPost, path: "/files/starred/1", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/total-space-used", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/total-files", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/total-directory", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/report-size-by-format", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/top-files-by-size?limit=5", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/duplicate-files", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/images", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/images?group_by=invalid", code: http.StatusBadRequest},
		{method: http.MethodGet, path: "/files/music", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/videos", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/music/artists", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/music/artists/x", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/music/albums", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/music/albums/x", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/music/genres", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/music/genres/x", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/music/folders", code: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/files/thumbnail/1", code: http.StatusBadRequest},
		{method: http.MethodGet, path: "/files/video-thumbnail/1", code: http.StatusBadRequest},
		{method: http.MethodGet, path: "/files/video-preview/1", code: http.StatusBadRequest},
		{method: http.MethodGet, path: "/files/stream/1", code: http.StatusNotFound},
		{method: http.MethodGet, path: "/files/video-stream/1", code: http.StatusNotFound},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != tc.code {
			t.Fatalf("path %s expected %d got %d body=%s", tc.path, tc.code, w.Code, w.Body.String())
		}
	}
}
