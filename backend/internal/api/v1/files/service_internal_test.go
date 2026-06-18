package files

import (
	"database/sql"
	"errors"
	"image"
	"image/color"
	"image/png"
	jobsapi "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
	"nas-go/api/internal/roots"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setProgramFilesForTest(t *testing.T) string {
	t.Helper()

	programFiles := filepath.Join(t.TempDir(), "ProgramFiles")
	t.Setenv("ProgramFiles", programFiles)
	return programFiles
}

func ensureTestIcons(t *testing.T) {
	t.Helper()

	setProgramFilesForTest(t)

	iconDir := config.GetBuildConfig("IconPath")
	if err := os.MkdirAll(iconDir, 0755); err != nil {
		t.Fatalf("failed to create icon dir: %v", err)
	}

	writeIcon := func(name string) {
		t.Helper()
		path := filepath.Join(iconDir, name+".png")
		if _, err := os.Stat(path); err == nil {
			return
		}

		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("failed to create icon file %s: %v", path, err)
		}
		defer f.Close()

		img := image.NewRGBA(image.Rect(0, 0, 2, 2))
		img.Set(0, 0, color.RGBA{R: 255, A: 255})
		img.Set(1, 0, color.RGBA{G: 255, A: 255})
		img.Set(0, 1, color.RGBA{B: 255, A: 255})
		img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		if err := png.Encode(f, img); err != nil {
			t.Fatalf("failed to encode icon %s: %v", path, err)
		}
	}

	for _, name := range []string{"folder", "unknown", "mp4", "mp3", "pdf"} {
		writeIcon(name)
	}
}

type filesRepoMock struct {
	db *database.DbContext

	createFileFn               func(transaction *sql.Tx, file FileModel) (FileModel, error)
	deleteFileByIDFn           func(transaction *sql.Tx, id int) error
	getFileByIDFn              func(id int) (FileModel, bool, error)
	getFilesByNameAndPathFn    func(name string, path string, limit int) ([]FileModel, error)
	getActiveChildrenFn        func(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	getActiveFilesByPathFn     func(path string, page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	getActiveFilesFn           func(page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	getFilesByPathPrefixFn     func(prefix string, page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	getFileStatByPathFn        func(path string) (FileStat, bool, error)
	updateFileFn               func(transaction *sql.Tx, file FileModel) (bool, error)
	updateDescendantPathsFn    func(transaction *sql.Tx, oldPath string, newPath string) (int64, error)
	markDeletedSubtreeFn       func(transaction *sql.Tx, path string, deletedAt time.Time) (int64, error)
	restoreSubtreeFn           func(transaction *sql.Tx, path string) (int64, error)
	getDirectoryContentCountFn func(fileId int, parentPath string) (int, error)
	getCountByTypeFn           func(fileType FileType) (int, error)
	getTotalSpaceUsedFn        func() (int, error)
	getReportSizeByFormatFn    func() ([]SizeReportModel, error)
	getTopFilesBySizeFn        func(limit int) ([]FileModel, error)
	getDuplicateFilesFn        func(page int, pageSize int) (utils.PaginationResponse[DuplicateFilesModel], error)
}

func (m *filesRepoMock) GetDbContext() *database.DbContext { return m.db }
func (m *filesRepoMock) CreateFile(transaction *sql.Tx, file FileModel) (FileModel, error) {
	if m.createFileFn != nil {
		return m.createFileFn(transaction, file)
	}
	return file, nil
}
func (m *filesRepoMock) DeleteFileByID(transaction *sql.Tx, id int) error {
	if m.deleteFileByIDFn != nil {
		return m.deleteFileByIDFn(transaction, id)
	}
	return nil
}
func (m *filesRepoMock) GetFileByID(id int) (FileModel, bool, error) {
	if m.getFileByIDFn != nil {
		return m.getFileByIDFn(id)
	}
	return FileModel{}, false, nil
}
func (m *filesRepoMock) GetFilesByNameAndPath(name string, path string, limit int) ([]FileModel, error) {
	if m.getFilesByNameAndPathFn != nil {
		return m.getFilesByNameAndPathFn(name, path, limit)
	}
	return nil, nil
}
func (m *filesRepoMock) GetActiveChildrenByParentPath(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.getActiveChildrenFn != nil {
		return m.getActiveChildrenFn(parentPath, category, page, pageSize)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}
func (m *filesRepoMock) GetActiveFilesByPath(path string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.getActiveFilesByPathFn != nil {
		return m.getActiveFilesByPathFn(path, page, pageSize)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}
func (m *filesRepoMock) GetActiveFiles(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.getActiveFilesFn != nil {
		return m.getActiveFilesFn(page, pageSize)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}
func (m *filesRepoMock) GetFilesByPathPrefix(prefix string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.getFilesByPathPrefixFn != nil {
		return m.getFilesByPathPrefixFn(prefix, page, pageSize)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}
func (m *filesRepoMock) GetFileStatByPath(path string) (FileStat, bool, error) {
	if m.getFileStatByPathFn != nil {
		return m.getFileStatByPathFn(path)
	}
	return FileStat{}, false, nil
}
func (m *filesRepoMock) UpdateFile(transaction *sql.Tx, file FileModel) (bool, error) {
	if m.updateFileFn != nil {
		return m.updateFileFn(transaction, file)
	}
	return true, nil
}
func (m *filesRepoMock) UpdateDescendantPaths(transaction *sql.Tx, oldPath string, newPath string) (int64, error) {
	if m.updateDescendantPathsFn != nil {
		return m.updateDescendantPathsFn(transaction, oldPath, newPath)
	}
	return 0, nil
}
func (m *filesRepoMock) MarkDeletedSubtree(transaction *sql.Tx, path string, deletedAt time.Time) (int64, error) {
	if m.markDeletedSubtreeFn != nil {
		return m.markDeletedSubtreeFn(transaction, path, deletedAt)
	}
	return 0, nil
}
func (m *filesRepoMock) RestoreSubtree(transaction *sql.Tx, path string) (int64, error) {
	if m.restoreSubtreeFn != nil {
		return m.restoreSubtreeFn(transaction, path)
	}
	return 0, nil
}
func (m *filesRepoMock) GetDirectoryContentCount(fileId int, parentPath string) (int, error) {
	if m.getDirectoryContentCountFn != nil {
		return m.getDirectoryContentCountFn(fileId, parentPath)
	}
	return 0, nil
}
func (m *filesRepoMock) GetCountByType(fileType FileType) (int, error) {
	if m.getCountByTypeFn != nil {
		return m.getCountByTypeFn(fileType)
	}
	return 0, nil
}
func (m *filesRepoMock) GetTotalSpaceUsed() (int, error) {
	if m.getTotalSpaceUsedFn != nil {
		return m.getTotalSpaceUsedFn()
	}
	return 0, nil
}
func (m *filesRepoMock) GetReportSizeByFormat() ([]SizeReportModel, error) {
	if m.getReportSizeByFormatFn != nil {
		return m.getReportSizeByFormatFn()
	}
	return nil, nil
}
func (m *filesRepoMock) GetTopFilesBySize(limit int) ([]FileModel, error) {
	if m.getTopFilesBySizeFn != nil {
		return m.getTopFilesBySizeFn(limit)
	}
	return nil, nil
}
func (m *filesRepoMock) GetDuplicateFiles(page int, pageSize int) (utils.PaginationResponse[DuplicateFilesModel], error) {
	if m.getDuplicateFilesFn != nil {
		return m.getDuplicateFilesFn(page, pageSize)
	}
	return utils.PaginationResponse[DuplicateFilesModel]{Items: []DuplicateFilesModel{}}, nil
}

type filesJobsRepoMock struct {
	jobsapi.RepositoryInterface
	db           *database.DbContext
	createdJobs  []jobsapi.JobModel
	createdSteps []jobsapi.StepModel
	createJobFn  func(tx *sql.Tx, job jobsapi.JobModel) (jobsapi.JobModel, error)
	createStepFn func(tx *sql.Tx, step jobsapi.StepModel) (jobsapi.StepModel, error)
}

func (m *filesJobsRepoMock) GetDbContext() *database.DbContext {
	return m.db
}

func (m *filesJobsRepoMock) CreateJob(tx *sql.Tx, job jobsapi.JobModel) (jobsapi.JobModel, error) {
	if m.createJobFn != nil {
		return m.createJobFn(tx, job)
	}
	job.ID = len(m.createdJobs) + 1
	m.createdJobs = append(m.createdJobs, job)
	return job, nil
}

func (m *filesJobsRepoMock) CreateStep(tx *sql.Tx, step jobsapi.StepModel) (jobsapi.StepModel, error) {
	if m.createStepFn != nil {
		return m.createStepFn(tx, step)
	}
	step.ID = len(m.createdSteps) + 1
	m.createdSteps = append(m.createdSteps, step)
	return step, nil
}

func newFilesServiceForTest(t *testing.T, repo *filesRepoMock) *Service {
	t.Helper()
	repo.db = database.NewDbContext(nil)
	return &Service{
		Repository: repo,
		Tasks:      make(chan utils.Task, 4),
	}
}

func newFilesJobsRepoMockForTest(t *testing.T) *filesJobsRepoMock {
	t.Helper()
	return &filesJobsRepoMock{
		db: database.NewDbContext(nil),
	}
}

func sampleModel(id int, name string, typ FileType) FileModel {
	return FileModel{
		ID:         id,
		Name:       name,
		Path:       "/tmp/" + name,
		ParentPath: "/tmp",
		Type:       typ,
		Format:     ".txt",
		Size:       10,
		UpdatedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}
}

func TestFileService_GetAndFind(t *testing.T) {
	repo := &filesRepoMock{
		getFilesByNameAndPathFn: func(name string, path string, limit int) ([]FileModel, error) {
			switch name {
			case "none":
				return nil, nil
			case "multi":
				return []FileModel{sampleModel(1, "a", File), sampleModel(2, "b", File)}, nil
			default:
				return []FileModel{sampleModel(1, "one", File)}, nil
			}
		},
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			if id == 123 {
				return sampleModel(123, "one", File), true, nil
			}
			return FileModel{}, false, nil
		},
	}
	s := newFilesServiceForTest(t, repo)

	if _, err := s.GetFileByNameAndPath("none", "/tmp/none"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	// Múltiplas linhas (arquivo recriado deixa a antiga soft-deleted convivendo
	// com a nova): pickActiveFile prefere a linha ativa em vez de dar erro.
	multi, err := s.GetFileByNameAndPath("multi", "/tmp/multi")
	if err != nil {
		t.Fatalf("expected active file, got error %v", err)
	}
	if multi.ID != 1 {
		t.Fatalf("expected active file id 1, got %d", multi.ID)
	}
	file, err := s.GetFileByNameAndPath("one", "/tmp/one")
	if err != nil || file.ID == 0 {
		t.Fatalf("expected one file, err=%v", err)
	}
	idFile, err := s.GetFileById(123)
	if err != nil || idFile.ID != 123 {
		t.Fatalf("expected id 123, got %+v err=%v", idFile, err)
	}
}

func TestFileService_CreateUploadProcessJob(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "video.mp4")
	if err := os.WriteFile(filePath, []byte("video"), 0644); err != nil {
		t.Fatalf("failed to create upload file: %v", err)
	}

	jobsRepo := newFilesJobsRepoMockForTest(t)
	service := newFilesServiceForTest(t, &filesRepoMock{})
	service.JobsRepository = jobsRepo

	jobID, err := service.CreateUploadProcessJob([]string{filePath})
	if err != nil {
		t.Fatalf("expected upload job creation success, got %v", err)
	}
	if jobID <= 0 {
		t.Fatalf("expected valid job id, got %d", jobID)
	}
	if len(jobsRepo.createdJobs) != 1 {
		t.Fatalf("expected one created job, got %d", len(jobsRepo.createdJobs))
	}
	if jobsRepo.createdJobs[0].Type != "upload_process" {
		t.Fatalf("expected upload_process job, got %s", jobsRepo.createdJobs[0].Type)
	}
	if jobsRepo.createdJobs[0].Priority != "high" {
		t.Fatalf("expected high priority job, got %s", jobsRepo.createdJobs[0].Priority)
	}

	hasPersist := false
	hasMetadata := false
	hasChecksum := false
	hasThumbnail := false
	hasPlaylist := false
	for _, step := range jobsRepo.createdSteps {
		switch step.Type {
		case "persist":
			hasPersist = true
		case "metadata":
			hasMetadata = true
		case "checksum":
			hasChecksum = true
		case "thumbnail":
			hasThumbnail = true
		case "playlist_index":
			hasPlaylist = true
		}
	}

	if !hasPersist || !hasMetadata || !hasChecksum || !hasThumbnail || !hasPlaylist {
		t.Fatalf("expected upload steps persist/metadata/checksum/thumbnail/playlist_index, got %+v", jobsRepo.createdSteps)
	}
}

func TestFileService_CreateCaptureProcessJob(t *testing.T) {
	jobsRepo := newFilesJobsRepoMockForTest(t)
	service := newFilesServiceForTest(t, &filesRepoMock{})
	service.JobsRepository = jobsRepo

	jobID, err := service.CreateCaptureProcessJob(42)
	if err != nil {
		t.Fatalf("expected capture job creation success, got %v", err)
	}
	if jobID <= 0 {
		t.Fatalf("expected valid job id, got %d", jobID)
	}
	if len(jobsRepo.createdJobs) != 1 {
		t.Fatalf("expected one created job, got %d", len(jobsRepo.createdJobs))
	}
	if jobsRepo.createdJobs[0].Type != "capture_process" {
		t.Fatalf("expected capture_process job, got %s", jobsRepo.createdJobs[0].Type)
	}
	if len(jobsRepo.createdSteps) != 1 || jobsRepo.createdSteps[0].Type != "capture_promote" {
		t.Fatalf("expected a single capture_promote step, got %+v", jobsRepo.createdSteps)
	}
}

func TestFileService_CreateCaptureProcessJobRequiresID(t *testing.T) {
	service := newFilesServiceForTest(t, &filesRepoMock{})
	service.JobsRepository = newFilesJobsRepoMockForTest(t)

	if _, err := service.CreateCaptureProcessJob(0); err == nil {
		t.Fatal("expected error for missing capture id")
	}
}

func TestFileService_GetChildrenAndDirectoryCount(t *testing.T) {
	repo := &filesRepoMock{
		getActiveChildrenFn: func(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{
				Items: []FileModel{
					sampleModel(1, "dir", Directory),
					sampleModel(2, "file", File),
				},
				Pagination: utils.Pagination{Page: 1, PageSize: 10},
			}, nil
		},
		getDirectoryContentCountFn: func(fileId int, parentPath string) (int, error) {
			if fileId == 1 {
				return 4, nil
			}
			return 0, errors.New("not directory")
		},
	}
	s := newFilesServiceForTest(t, repo)

	result, err := s.GetChildrenByParentPath("/tmp", AllCategory, 1, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected two items")
	}
	if result.Items[0].DirectoryContentCount != 4 {
		t.Fatalf("expected directory content count 4")
	}
}

func TestFileService_DecomposedListings(t *testing.T) {
	t.Run("GetFilesByPath converts and counts directories", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getActiveFilesByPathFn: func(path string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
				return utils.PaginationResponse[FileModel]{
					Items: []FileModel{sampleModel(1, "dir", Directory)},
				}, nil
			},
			getDirectoryContentCountFn: func(fileId int, parentPath string) (int, error) {
				return 3, nil
			},
		})

		out, err := s.GetFilesByPath("/tmp/dir", 1, 10)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if len(out.Items) != 1 || out.Items[0].DirectoryContentCount != 3 {
			t.Fatalf("expected one directory with count 3, got %+v", out.Items)
		}
	})

	t.Run("GetFilesByPath propagates repository error", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getActiveFilesByPathFn: func(path string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
				return utils.PaginationResponse[FileModel]{}, errors.New("path lookup failed")
			},
		})

		if _, err := s.GetFilesByPath("/tmp/dir", 1, 10); err == nil {
			t.Fatalf("expected GetFilesByPath error")
		}
	})

	t.Run("GetActiveFilesPage converts the page", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getActiveFilesFn: func(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
				return utils.PaginationResponse[FileModel]{
					Items: []FileModel{sampleModel(2, "file", File)},
				}, nil
			},
		})

		out, err := s.GetActiveFilesPage(1, 10)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if len(out.Items) != 1 || out.Items[0].ID != 2 {
			t.Fatalf("expected file 2, got %+v", out.Items)
		}
	})

	t.Run("GetActiveFilesPage propagates repository error", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getActiveFilesFn: func(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
				return utils.PaginationResponse[FileModel]{}, errors.New("listing failed")
			},
		})

		if _, err := s.GetActiveFilesPage(1, 10); err == nil {
			t.Fatalf("expected GetActiveFilesPage error")
		}
	})

	t.Run("GetFilesByPathPrefix converts without directory counts", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getFilesByPathPrefixFn: func(prefix string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
				return utils.PaginationResponse[FileModel]{
					Items: []FileModel{sampleModel(3, "dir", Directory)},
				}, nil
			},
			getDirectoryContentCountFn: func(fileId int, parentPath string) (int, error) {
				t.Fatalf("prefix walk must not count directory contents")
				return 0, nil
			},
		})

		out, err := s.GetFilesByPathPrefix("/tmp", 1, 10)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if len(out.Items) != 1 || out.Items[0].ID != 3 {
			t.Fatalf("expected file 3, got %+v", out.Items)
		}
	})

	t.Run("GetFilesByPathPrefix propagates repository error", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getFilesByPathPrefixFn: func(prefix string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
				return utils.PaginationResponse[FileModel]{}, errors.New("prefix walk failed")
			},
		})

		if _, err := s.GetFilesByPathPrefix("/tmp", 1, 10); err == nil {
			t.Fatalf("expected GetFilesByPathPrefix error")
		}
	})

	t.Run("GetFileStatByPath passes through", func(t *testing.T) {
		now := time.Now()
		s := newFilesServiceForTest(t, &filesRepoMock{
			getFileStatByPathFn: func(path string) (FileStat, bool, error) {
				return FileStat{Size: 42, UpdatedAt: now}, true, nil
			},
		})

		stat, found, err := s.GetFileStatByPath("/tmp/file")
		if err != nil || !found {
			t.Fatalf("expected stat found, got found=%v err=%v", found, err)
		}
		if stat.Size != 42 || !stat.UpdatedAt.Equal(now) {
			t.Fatalf("expected stat passthrough, got %+v", stat)
		}
	})
}

func TestFileService_CreateUpdateAndMetadata(t *testing.T) {
	repo := &filesRepoMock{
		createFileFn: func(transaction *sql.Tx, file FileModel) (FileModel, error) {
			file.ID = 77
			return file, nil
		},
		updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
			return true, nil
		},
	}
	s := newFilesServiceForTest(t, repo)

	file, err := s.CreateFile(FileDto{
		Name:       "f",
		Path:       "/tmp/f",
		ParentPath: "/tmp",
		Type:       File,
	})
	if err != nil {
		t.Fatalf("expected create success, got %v", err)
	}
	if file.ID != 77 {
		t.Fatalf("expected id 77")
	}

	ok, err := s.UpdateFile(FileDto{
		ID:         77,
		Name:       "f",
		Path:       "/tmp/f",
		ParentPath: "/tmp",
		Type:       File,
	})
	if err != nil || !ok {
		t.Fatalf("expected update success, ok=%v err=%v", ok, err)
	}

}

func TestFileService_ScanAndExistsAndBlob(t *testing.T) {
	repo := &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			if id == 999 {
				return FileModel{}, false, nil
			}
			return sampleModel(10, "blob.txt", File), true, nil
		},
	}
	s := newFilesServiceForTest(t, repo)

	s.ScanFilesTask("x")
	s.ScanDirTask("/tmp")
	if len(s.Tasks) < 2 {
		t.Fatalf("expected tasks to be enqueued")
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "blob.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if !s.CheckFileExistsByPath(tmpFile) {
		t.Fatalf("expected file to exist")
	}
	if s.CheckFileExistsByPath(filepath.Join(tmpDir, "missing")) {
		t.Fatalf("expected missing file to not exist")
	}

	// Rebind repo return path to existing file for blob retrieval.
	repo.getFileByIDFn = func(id int) (FileModel, bool, error) {
		return FileModel{
			ID:         10,
			Name:       "blob.txt",
			Path:       tmpFile,
			ParentPath: tmpDir,
			Type:       File,
			Format:     ".txt",
			UpdatedAt:  time.Now(),
			CreatedAt:  time.Now(),
		}, true, nil
	}

	blob, err := s.GetFileBlobById(10)
	if err != nil || len(blob.Blob) == 0 {
		t.Fatalf("expected blob bytes, err=%v", err)
	}
	if !s.CheckFileExists(10) {
		t.Fatalf("expected CheckFileExists true for existing file")
	}
	repo.getFileByIDFn = func(id int) (FileModel, bool, error) {
		return FileModel{}, false, nil
	}
	if s.CheckFileExists(999) {
		t.Fatalf("expected CheckFileExists false for missing file")
	}
}

func TestFileService_ReportsAndWrappers(t *testing.T) {
	repo := &filesRepoMock{
		getTotalSpaceUsedFn: func() (int, error) { return 1000, nil },
		getCountByTypeFn: func(fileType FileType) (int, error) {
			if fileType == File {
				return 50, nil
			}
			return 20, nil
		},
		getReportSizeByFormatFn: func() ([]SizeReportModel, error) {
			return []SizeReportModel{
				{Format: ".mp3", Total: 3, Size: 300},
				{Format: ".jpg", Total: 2, Size: 200},
			}, nil
		},
		getTopFilesBySizeFn: func(limit int) ([]FileModel, error) {
			return []FileModel{sampleModel(1, "top", File)}, nil
		},
		getDuplicateFilesFn: func(page int, pageSize int) (utils.PaginationResponse[DuplicateFilesModel], error) {
			return utils.PaginationResponse[DuplicateFilesModel]{
				Items: []DuplicateFilesModel{{Name: "d", Size: 10, Copies: 2, Paths: "/a,/b"}},
				Pagination: utils.Pagination{
					Page: 1, PageSize: 10,
				},
			}, nil
		},
	}
	s := newFilesServiceForTest(t, repo)

	if total, _ := s.GetTotalSpaceUsed(); total != 1000 {
		t.Fatalf("expected total space 1000")
	}
	if filesCount, _ := s.GetTotalFiles(); filesCount != 50 {
		t.Fatalf("expected total files 50")
	}
	if dirsCount, _ := s.GetTotalDirectory(); dirsCount != 20 {
		t.Fatalf("expected total dirs 20")
	}

	report, err := s.GetReportSizeByFormat()
	if err != nil || len(report) == 0 {
		t.Fatalf("expected report entries, err=%v", err)
	}
	top, err := s.GetTopFilesBySize(1)
	if err != nil || len(top) != 1 {
		t.Fatalf("expected one top file, err=%v", err)
	}
	dups, err := s.GetDuplicateFiles(1, 10)
	if err != nil || dups.TotalFiles != 2 {
		t.Fatalf("expected duplicates report, err=%v", err)
	}

}

func TestFileService_DeleteAndChecksumBranches(t *testing.T) {
	repo := &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			return sampleModel(id, "x", File), true, nil
		},
	}
	s := newFilesServiceForTest(t, repo)

	alreadyDeleted := FileDto{
		ID:        1,
		Name:      "x",
		Path:      "/tmp/x",
		DeletedAt: utils.Optional[time.Time]{HasValue: true, Value: time.Now()},
	}
	if err := s.DeleteFile(alreadyDeleted, false); err == nil {
		t.Fatalf("expected already deleted error")
	}

	recent := FileDto{
		ID:              2,
		Name:            "y",
		Path:            "/tmp/y",
		LastInteraction: utils.Optional[time.Time]{HasValue: true, Value: time.Now()},
	}
	if err := s.DeleteFile(recent, true); err == nil {
		t.Fatalf("expected recently accessed deletion error")
	}
	if err := s.DeleteFile(recent, false); err != nil {
		t.Fatalf("expected manual delete to ignore recent-access rule, got %v", err)
	}

	// Force default branch in UpdateCheckSum.
	repo.getFileByIDFn = func(id int) (FileModel, bool, error) {
		return sampleModel(id, "unknown", FileType(99)), true, nil
	}
	if err := s.UpdateCheckSum(9); err == nil {
		t.Fatalf("expected unknown file type error")
	}
}

func TestFileService_ChecksumThumbnailAndDeleteSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "data.txt")
	if err := os.WriteFile(filePath, []byte("checksum"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	repo := &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			switch id {
			case 1:
				return FileModel{
					ID:         1,
					Name:       "data.txt",
					Path:       filePath,
					ParentPath: tmpDir,
					Type:       File,
					Format:     ".txt",
					UpdatedAt:  time.Now(),
					CreatedAt:  time.Now(),
				}, true, nil
			case 2:
				return FileModel{
					ID:         2,
					Name:       "dir",
					Path:       tmpDir,
					ParentPath: filepath.Dir(tmpDir),
					Type:       Directory,
					UpdatedAt:  time.Now(),
					CreatedAt:  time.Now(),
				}, true, nil
			default:
				return FileModel{}, false, nil
			}
		},
		getActiveChildrenFn: func(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{
				Items: []FileModel{
					{
						ID:         3,
						Name:       "child1",
						Path:       filepath.Join(tmpDir, "c1"),
						ParentPath: tmpDir,
						Type:       File,
						CheckSum:   "abcd",
						UpdatedAt:  time.Now(),
						CreatedAt:  time.Now(),
					},
				},
				Pagination: utils.Pagination{Page: 1, PageSize: 1000, HasNext: false},
			}, nil
		},
	}
	s := newFilesServiceForTest(t, repo)

	if err := s.UpdateCheckSum(1); err != nil {
		t.Fatalf("expected file checksum update success, got %v", err)
	}
	if err := s.UpdateCheckSum(2); err != nil {
		t.Fatalf("expected dir checksum update success, got %v", err)
	}
	queuedTasks := 0
drainLoop:
	for {
		select {
		case task := <-s.Tasks:
			if task.Type == utils.UpdateCheckSum {
				queuedTasks++
			}
		default:
			break drainLoop
		}
	}
	if queuedTasks < 2 {
		t.Fatalf("expected at least two queued checksum tasks, got %d", queuedTasks)
	}

	oldAccess := time.Now().Add(-48 * time.Hour)
	err := s.DeleteFile(FileDto{
		ID:              4,
		Name:            "old.txt",
		Path:            filePath,
		Type:            File,
		LastInteraction: utils.Optional[time.Time]{HasValue: true, Value: oldAccess},
	}, true)
	if err != nil {
		t.Fatalf("expected DeleteFile success for old access file, got %v", err)
	}

	thumbData, err := s.GetFileThumbnail(FileDto{
		ID:              5,
		Name:            "missing.txt",
		Path:            filepath.Join(tmpDir, "missing.txt"),
		ParentPath:      tmpDir,
		Type:            File,
		Format:          ".txt",
		LastInteraction: utils.Optional[time.Time]{HasValue: true, Value: oldAccess},
	}, 100, 100)
	if err == nil || thumbData != nil {
		t.Fatalf("expected GetFileThumbnail to fail on missing file with nil data")
	}

}

func TestFileService_ThumbnailFallbacks(t *testing.T) {
	ensureTestIcons(t)

	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "not-image.txt")
	if err := os.WriteFile(existingFile, []byte("plain text"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	s := newFilesServiceForTest(t, &filesRepoMock{})

	dirThumb, err := s.GetFileThumbnail(FileDto{
		ID:   101,
		Path: tmpDir,
		Type: Directory,
	}, -10, 120)
	if err != nil {
		t.Fatalf("expected directory thumbnail success, got %v", err)
	}
	if len(dirThumb) == 0 {
		t.Fatalf("expected non-empty directory thumbnail")
	}

	fileThumb, err := s.GetFileThumbnail(FileDto{
		ID:     102,
		Path:   existingFile,
		Type:   File,
		Format: ".txt",
	}, 4096, 120)
	if err != nil {
		t.Fatalf("expected file fallback thumbnail success, got %v", err)
	}
	if len(fileThumb) == 0 {
		t.Fatalf("expected non-empty file thumbnail")
	}

}

func TestFileService_GetFileThumbnailCacheHit(t *testing.T) {
	setProgramFilesForTest(t)
	s := newFilesServiceForTest(t, &filesRepoMock{})
	cacheDir := config.GetBuildConfig("ThumbnailPath")
	cacheFile := filepath.Join(cacheDir, "42_320.png")
	cached := []byte("cached-png")

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := os.WriteFile(cacheFile, cached, 0644); err != nil {
		t.Fatalf("failed to write cache file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(cacheFile) })

	data, err := s.GetFileThumbnail(FileDto{
		ID:   42,
		Path: "/path/not/used/on-cache-hit",
		Type: File,
	}, 0, 100)
	if err != nil {
		t.Fatalf("expected cache hit without error, got %v", err)
	}
	if string(data) != string(cached) {
		t.Fatalf("expected cached data, got %q", string(data))
	}
}

func TestFileService_GetFileThumbnailMissingFileDeleteFailure(t *testing.T) {
	tmpDir := t.TempDir()
	s := newFilesServiceForTest(t, &filesRepoMock{
		updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
			return false, errors.New("update failure")
		},
	})

	_, err := s.GetFileThumbnail(FileDto{
		ID:     99,
		Path:   filepath.Join(tmpDir, "missing.txt"),
		Type:   File,
		Format: ".txt",
	}, 100, 100)
	if err == nil {
		t.Fatalf("expected error for missing file with delete failure")
	}
	if !errors.Is(err, ErrDatabase) {
		t.Fatalf("expected ErrDatabase wrapping, got %v", err)
	}
}

func TestFileService_ErrorBranches(t *testing.T) {
	s := newFilesServiceForTest(t, &filesRepoMock{
		getActiveChildrenFn: func(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{}, errors.New("get children failed")
		},
		createFileFn: func(transaction *sql.Tx, file FileModel) (FileModel, error) {
			return FileModel{}, errors.New("create failed")
		},
		updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
			return false, errors.New("update failed")
		},
	})

	if _, err := s.GetChildrenByParentPath("/tmp", AllCategory, 1, 10); err == nil {
		t.Fatalf("expected GetChildrenByParentPath error")
	}
	if _, err := s.CreateFile(FileDto{Name: "x", Path: "/tmp/x", ParentPath: "/tmp", Type: File}); err == nil {
		t.Fatalf("expected CreateFile error")
	}
	if _, err := s.UpdateFile(FileDto{ID: 1, Name: "x", Path: "/tmp/x", ParentPath: "/tmp", Type: File}); err == nil {
		t.Fatalf("expected UpdateFile error")
	}
}

func TestFileService_GetFileBlobByIdReadError(t *testing.T) {
	s := newFilesServiceForTest(t, &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			return sampleModel(999, "missing.bin", File), true, nil
		},
	})

	if _, err := s.GetFileBlobById(999); err == nil {
		t.Fatalf("expected GetFileBlobById read error")
	}
}

func TestFileService_AdditionalErrorAndEdgeBranches(t *testing.T) {
	t.Run("listing with directory count error falls back to zero", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getActiveChildrenFn: func(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
				return utils.PaginationResponse[FileModel]{
					Items: []FileModel{sampleModel(1, "dir", Directory)},
				}, nil
			},
			getDirectoryContentCountFn: func(fileId int, parentPath string) (int, error) {
				return 0, errors.New("count failed")
			},
		})

		out, err := s.GetChildrenByParentPath("/tmp", AllCategory, 1, 10)
		if err != nil {
			t.Fatalf("expected GetChildrenByParentPath success, got %v", err)
		}
		if out.Items[0].DirectoryContentCount != 0 {
			t.Fatalf("expected directory count fallback to 0")
		}
	})

	t.Run("GetFileById no rows and found row", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getFileByIDFn: func(id int) (FileModel, bool, error) {
				if id == 10 {
					return FileModel{}, false, nil
				}
				return sampleModel(id, "a", File), true, nil
			},
		})

		if _, err := s.GetFileById(10); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
		found, err := s.GetFileById(11)
		if err != nil || found.ID != 11 {
			t.Fatalf("expected file 11, got %+v err=%v", found, err)
		}
	})

	t.Run("UpdateCheckSum enqueues task for valid file type", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "x.txt")
		if err := os.WriteFile(filePath, []byte("abc"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		s := newFilesServiceForTest(t, &filesRepoMock{
			getFileByIDFn: func(id int) (FileModel, bool, error) {
				return FileModel{
					ID:         id,
					Name:       "x.txt",
					Path:       filePath,
					ParentPath: tmpDir,
					Type:       File,
					Format:     ".txt",
					UpdatedAt:  time.Now(),
					CreatedAt:  time.Now(),
				}, true, nil
			},
		})

		if err := s.UpdateCheckSum(1); err != nil {
			t.Fatalf("expected UpdateCheckSum success, got %v", err)
		}

		select {
		case task := <-s.Tasks:
			if task.Type != utils.UpdateCheckSum {
				t.Fatalf("expected UpdateCheckSum task type, got %v", task.Type)
			}
		default:
			t.Fatalf("expected checksum task enqueued")
		}
	})

	t.Run("UpdateCheckSum propagates get-file error", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getFileByIDFn: func(id int) (FileModel, bool, error) {
				return FileModel{}, false, errors.New("repository down")
			},
		})

		if err := s.UpdateCheckSum(1); err == nil {
			t.Fatalf("expected UpdateCheckSum to propagate fetch error")
		}
	})

	t.Run("GetReportSizeByFormat propagates repository error", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getReportSizeByFormatFn: func() ([]SizeReportModel, error) {
				return nil, errors.New("report failed")
			},
		})

		if _, err := s.GetReportSizeByFormat(); err == nil {
			t.Fatalf("expected report error")
		}
	})

	t.Run("CheckFileExists returns false on repository error", func(t *testing.T) {
		s := newFilesServiceForTest(t, &filesRepoMock{
			getFileByIDFn: func(id int) (FileModel, bool, error) {
				return FileModel{}, false, errors.New("query failed")
			},
		})

		if s.CheckFileExists(10) {
			t.Fatalf("expected CheckFileExists false on repository error")
		}
	})

	t.Run("DeleteFile propagates update errors", func(t *testing.T) {
		sErr := newFilesServiceForTest(t, &filesRepoMock{
			updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
				return false, errors.New("update failed")
			},
		})

		err := sErr.DeleteFile(FileDto{
			ID:              1,
			Name:            "x",
			Path:            "/tmp/x",
			Type:            File,
			LastInteraction: utils.Optional[time.Time]{HasValue: true, Value: time.Now().Add(-48 * time.Hour)},
		}, true)
		if err == nil {
			t.Fatalf("expected DeleteFile update error")
		}

		sFalse := newFilesServiceForTest(t, &filesRepoMock{
			updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
				return false, nil
			},
		})

		err = sFalse.DeleteFile(FileDto{
			ID:              2,
			Name:            "y",
			Path:            "/tmp/y",
			Type:            File,
			LastInteraction: utils.Optional[time.Time]{HasValue: true, Value: time.Now().Add(-48 * time.Hour)},
		}, true)
		if err == nil {
			t.Fatalf("expected DeleteFile not-found error")
		}
	})
}

// TestPickActiveFile exercita diretamente a função pura de resolução, sem mocks:
// 0 linhas -> ErrNoRows; 1 linha -> ela; várias -> prefere a linha não deletada
// (arquivo recriado deixa a antiga soft-deleted); todas deletadas -> a primeira.
func TestPickActiveFile(t *testing.T) {
	deleted := func(id int) FileDto {
		return FileDto{ID: id, DeletedAt: utils.Optional[time.Time]{HasValue: true, Value: time.Now()}}
	}
	active := func(id int) FileDto {
		return FileDto{ID: id}
	}

	t.Run("no rows returns ErrNoRows", func(t *testing.T) {
		if _, err := pickActiveFile(nil); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
	})

	t.Run("single row returns it", func(t *testing.T) {
		got, err := pickActiveFile([]FileDto{active(7)})
		if err != nil || got.ID != 7 {
			t.Fatalf("expected id 7 no error, got %d err=%v", got.ID, err)
		}
	})

	t.Run("multiple prefers the active row over the soft-deleted one", func(t *testing.T) {
		// itens chegam ordenados por id DESC; a linha nova (ativa) pode vir depois da antiga deletada.
		got, err := pickActiveFile([]FileDto{deleted(2), active(1)})
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}
		if got.ID != 1 {
			t.Fatalf("expected active id 1, got %d", got.ID)
		}
	})

	t.Run("multiple all deleted falls back to the first", func(t *testing.T) {
		got, err := pickActiveFile([]FileDto{deleted(9), deleted(8)})
		if err != nil || got.ID != 9 {
			t.Fatalf("expected first id 9 no error, got %d err=%v", got.ID, err)
		}
	})
}

func TestGetRootNodesListsRevivesAndSelfHeals(t *testing.T) {
	previousEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = previousEntryPoint
		roots.Reset()
	})
	config.AppConfig.EntryPoint = ""

	indexedRoot := t.TempDir()                             // already has an active row
	freshRoot := t.TempDir()                               // no row yet → created on the fly
	missingRoot := filepath.Join(t.TempDir(), "unplugged") // not on disk → skipped
	disabledRoot := t.TempDir()                            // disabled → never listed

	roots.Set([]roots.Root{
		{ID: 1, Path: indexedRoot, Label: "Principal", Enabled: true},
		{ID: 2, Path: freshRoot, Label: "Midia", Enabled: true},
		{ID: 3, Path: missingRoot, Label: "Externo", Enabled: true},
		{ID: 4, Path: disabledRoot, Label: "Oculto", Enabled: false},
	})

	created := []FileModel{}
	repo := &filesRepoMock{
		getFilesByNameAndPathFn: func(name string, path string, limit int) ([]FileModel, error) {
			if path == indexedRoot {
				return []FileModel{{ID: 11, Name: name, Path: path, ParentPath: filepath.Dir(path), Type: Directory}}, nil
			}
			return nil, nil
		},
		createFileFn: func(transaction *sql.Tx, file FileModel) (FileModel, error) {
			file.ID = 22
			created = append(created, file)
			return file, nil
		},
		getDirectoryContentCountFn: func(fileId int, parentPath string) (int, error) { return 3, nil },
	}
	service := &Service{Repository: repo}

	nodes, err := service.GetRootNodes()
	if err != nil {
		t.Fatalf("GetRootNodes: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes (indexed + fresh), got %+v", nodes)
	}
	if nodes[0].ID != 11 || nodes[0].Name != "Principal" || nodes[0].Path != indexedRoot {
		t.Fatalf("unexpected first node: %+v", nodes[0])
	}
	if nodes[1].ID != 22 || nodes[1].Name != "Midia" || nodes[1].Path != freshRoot {
		t.Fatalf("unexpected second node: %+v", nodes[1])
	}
	if nodes[0].DirectoryContentCount != 3 {
		t.Fatalf("expected content count filled, got %+v", nodes[0])
	}
	if len(created) != 1 || created[0].Path != freshRoot || created[0].Type != Directory {
		t.Fatalf("expected the fresh root row to be self-created, got %+v", created)
	}
}

func TestGetRootNodesRevivesSoftDeletedRow(t *testing.T) {
	previousEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = previousEntryPoint
		roots.Reset()
	})
	config.AppConfig.EntryPoint = ""

	rootPath := t.TempDir()
	roots.Set([]roots.Root{{ID: 1, Path: rootPath, Label: "Principal", Enabled: true}})

	updates := []FileModel{}
	repo := &filesRepoMock{
		getFilesByNameAndPathFn: func(name string, path string, limit int) ([]FileModel, error) {
			return []FileModel{{
				ID:        7,
				Name:      name,
				Path:      path,
				Type:      Directory,
				DeletedAt: sql.NullTime{Valid: true, Time: time.Now()},
			}}, nil
		},
		updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
			updates = append(updates, file)
			return true, nil
		},
	}
	service := &Service{Repository: repo}

	nodes, err := service.GetRootNodes()
	if err != nil {
		t.Fatalf("GetRootNodes: %v", err)
	}
	if len(nodes) != 1 || nodes[0].ID != 7 || nodes[0].DeletedAt.HasValue {
		t.Fatalf("expected the soft-deleted row revived, got %+v", nodes)
	}
	if len(updates) != 1 || updates[0].DeletedAt.Valid {
		t.Fatalf("expected an update clearing deleted_at, got %+v", updates)
	}
}
