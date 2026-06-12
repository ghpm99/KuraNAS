package files

import (
	"database/sql"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nas-go/api/internal/roots"
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
func (m *filesHandlerServiceMock) listingPage(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	return utils.PaginationResponse[FileDto]{
		Items: []FileDto{{ID: 1, Name: "a", Path: "/tmp/a", ParentPath: "/tmp"}},
		Pagination: utils.Pagination{
			Page: page, PageSize: pageSize,
		},
	}, nil
}
func (m *filesHandlerServiceMock) GetChildrenByParentPath(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	return m.listingPage(page, pageSize)
}
func (m *filesHandlerServiceMock) GetRootNodes() ([]FileDto, error) {
	page, err := m.listingPage(1, 1)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
func (m *filesHandlerServiceMock) GetFilesByPath(path string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	return m.listingPage(page, pageSize)
}
func (m *filesHandlerServiceMock) GetActiveFilesPage(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	return m.listingPage(page, pageSize)
}
func (m *filesHandlerServiceMock) GetFilesByPathPrefix(prefix string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	return m.listingPage(page, pageSize)
}
func (m *filesHandlerServiceMock) GetFileStatByPath(path string) (FileStat, bool, error) {
	return FileStat{}, false, nil
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
func (m *filesHandlerServiceMock) CheckFileExists(fileId int) bool              { return false }
func (m *filesHandlerServiceMock) CheckFileExistsByPath(path string) bool       { return false }
func (m *filesHandlerServiceMock) DeleteFile(file FileDto, bySystem bool) error { return nil }
func (m *filesHandlerServiceMock) RestoreSubtree(path string) error             { return nil }
func (m *filesHandlerServiceMock) UploadFiles(targetFolderID int, files []*multipart.FileHeader) (UploadFilesResult, error) {
	return UploadFilesResult{}, nil
}
func (m *filesHandlerServiceMock) CreateFolder(parentID *int, name string) (string, error) {
	return "", nil
}
func (m *filesHandlerServiceMock) MoveFile(sourceID int, destinationFolderID *int, destinationPath string) (string, error) {
	return "", nil
}
func (m *filesHandlerServiceMock) DeleteFileFromDisk(id int, permanent bool) error { return nil }
func (m *filesHandlerServiceMock) SetTrashBin(trashBin TrashBinInterface)          {}
func (m *filesHandlerServiceMock) RenameFile(id int, newName string) (string, error) {
	return newName, nil
}
func (m *filesHandlerServiceMock) CopyFile(sourceID int, destinationFolderID *int, destinationPath string, newName string) (string, error) {
	return "", nil
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
	getChildrenFn        func(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileDto], error)
	getFilesByPathFn     func(path string, page int, pageSize int) (utils.PaginationResponse[FileDto], error)
	getActiveFilesFn     func(page int, pageSize int) (utils.PaginationResponse[FileDto], error)
	getFileByIdFn        func(id int) (FileDto, error)
	updateFileFn         func(file FileDto) (bool, error)
	getFileBlobByIdFn    func(fileId int) (FileBlob, error)
	getTotalSpaceUsedFn  func() (int, error)
	getTotalFilesFn      func() (int, error)
	getTotalDirectoryFn  func() (int, error)
	getReportSizeByFmtFn func() ([]SizeReportDto, error)
	getTopFilesBySizeFn  func(limit int) ([]FileDto, error)
	getDuplicateFilesFn  func(page int, pageSize int) (DuplicateFileReportDto, error)
}

func (m *filesHandlerServiceFuncMock) GetChildrenByParentPath(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	if m.getChildrenFn != nil {
		return m.getChildrenFn(parentPath, category, page, pageSize)
	}
	return m.filesHandlerServiceMock.GetChildrenByParentPath(parentPath, category, page, pageSize)
}
func (m *filesHandlerServiceFuncMock) GetFilesByPath(path string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	if m.getFilesByPathFn != nil {
		return m.getFilesByPathFn(path, page, pageSize)
	}
	return m.filesHandlerServiceMock.GetFilesByPath(path, page, pageSize)
}
func (m *filesHandlerServiceFuncMock) GetActiveFilesPage(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	if m.getActiveFilesFn != nil {
		return m.getActiveFilesFn(page, pageSize)
	}
	return m.filesHandlerServiceMock.GetActiveFilesPage(page, pageSize)
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

func TestFilesHandlerThumbnailMissingSource(t *testing.T) {
	tmpDir := t.TempDir()
	service := &filesStreamServiceMock{filePath: filepath.Join(tmpDir, "f.txt"), format: ".txt"}
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})

	router := gin.New()
	router.GET("/files/thumbnail/:id", handler.GetFileThumbnailHandler)

	req := httptest.NewRequest(http.MethodGet, "/files/thumbnail/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing thumbnail source, got %d", w.Code)
	}
}

func TestFilesHandlerGetChildrenByIdNotFound(t *testing.T) {
	service := &filesHandlerServiceFuncMock{
		getFileByIdFn: func(id int) (FileDto, error) {
			return FileDto{}, sql.ErrNoRows
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
		getChildrenFn: func(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			if parentPath != expectedParentPath {
				t.Fatalf("expected parent path %q, got %q", expectedParentPath, parentPath)
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
		getChildrenFn: func(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			return utils.PaginationResponse[FileDto]{}, errBoom
		},
		getFilesByPathFn: func(path string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			return utils.PaginationResponse[FileDto]{}, errBoom
		},
		getActiveFilesFn: func(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
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
		{method: http.MethodGet, path: "/files/thumbnail/1", code: http.StatusBadRequest},
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

func TestGetFilesTreeHandlerMultiRootLevelZero(t *testing.T) {
	t.Cleanup(roots.Reset)
	rootA := t.TempDir()
	rootB := t.TempDir()
	roots.Set([]roots.Root{
		{ID: 1, Path: rootA, Label: "Principal", Enabled: true},
		{ID: 2, Path: rootB, Label: "Midia", Enabled: true},
	})

	handler := NewHandler(&filesHandlerServiceMock{}, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.GET("/files/tree", handler.GetFilesTreeHandler)

	req := httptest.NewRequest(http.MethodGet, "/files/tree", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}

	var payload struct {
		Items []FileDto `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// The mock returns one root node; the handler must serve exactly the
	// GetRootNodes result instead of the legacy children listing.
	if len(payload.Items) != 1 || payload.Items[0].ID != 1 {
		t.Fatalf("expected the root-node listing, got %+v", payload.Items)
	}

	// Asking for a specific folder keeps the legacy children behavior even
	// with multiple roots.
	req = httptest.NewRequest(http.MethodGet, "/files/tree?file_parent=9", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for child listing, got %d", w.Code)
	}
}
