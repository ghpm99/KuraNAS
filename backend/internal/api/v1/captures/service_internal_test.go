package captures

import (
	"bytes"
	"database/sql"
	"errors"
	"mime/multipart"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Repository mock
// ---------------------------------------------------------------------------

type repoMock struct {
	dbContext      *database.DbContext
	createFn      func(tx *sql.Tx, capture CaptureModel) (CaptureModel, error)
	getCapturesFn func(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureModel], error)
	getByIDFn     func(id int) (CaptureModel, error)
	deleteFn      func(tx *sql.Tx, id int) error
}

func (r *repoMock) GetDbContext() *database.DbContext { return r.dbContext }

func (r *repoMock) CreateCapture(tx *sql.Tx, capture CaptureModel) (CaptureModel, error) {
	if r.createFn != nil {
		return r.createFn(tx, capture)
	}
	capture.ID = 1
	return capture, nil
}

func (r *repoMock) GetCaptures(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureModel], error) {
	if r.getCapturesFn != nil {
		return r.getCapturesFn(filter, page, pageSize)
	}
	return utils.PaginationResponse[CaptureModel]{Items: []CaptureModel{}}, nil
}

func (r *repoMock) GetCaptureByID(id int) (CaptureModel, error) {
	if r.getByIDFn != nil {
		return r.getByIDFn(id)
	}
	return CaptureModel{ID: id, Name: "test"}, nil
}

func (r *repoMock) DeleteCapture(tx *sql.Tx, id int) error {
	if r.deleteFn != nil {
		return r.deleteFn(tx, id)
	}
	return nil
}

func newServiceForTest(t *testing.T, mock *repoMock) *Service {
	t.Helper()
	mock.dbContext = database.NewDbContext(nil)
	return &Service{Repository: mock}
}

func setEntryPointForTest(t *testing.T, dir string) {
	t.Helper()
	orig := config.AppConfig.EntryPoint
	config.AppConfig.EntryPoint = dir
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = orig
	})
}

func buildTestFileHeader(t *testing.T, filename string, content string) *multipart.FileHeader {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte(content))
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(1024 * 1024)
	if err != nil {
		t.Fatal(err)
	}
	return form.File["file"][0]
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestServiceUploadCapture(t *testing.T) {
	dir := t.TempDir()
	setEntryPointForTest(t, dir)

	mock := &repoMock{
		createFn: func(tx *sql.Tx, capture CaptureModel) (CaptureModel, error) {
			capture.ID = 1
			return capture, nil
		},
	}
	service := newServiceForTest(t, mock)
	file := buildTestFileHeader(t, "video.ts", "fake-ts-data")

	dto := CreateCaptureDto{
		Name:      "my_show",
		MediaType: "hls",
		MimeType:  "video/mp2t",
	}

	result, err := service.UploadCapture(file, dto)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != 1 {
		t.Fatalf("expected ID 1, got %d", result.ID)
	}
	if result.Name != "my_show" {
		t.Fatalf("expected name my_show, got %s", result.Name)
	}

	captureDir := filepath.Join(dir, "capturas", "my_show")
	if _, err := os.Stat(captureDir); os.IsNotExist(err) {
		t.Fatal("expected capturas directory to be created")
	}

	savedPath := filepath.Join(captureDir, "video.ts")
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		t.Fatal("expected file to be saved")
	}
}

func TestServiceUploadCaptureRepoError(t *testing.T) {
	dir := t.TempDir()
	setEntryPointForTest(t, dir)

	mock := &repoMock{
		createFn: func(tx *sql.Tx, capture CaptureModel) (CaptureModel, error) {
			return CaptureModel{}, errors.New("db error")
		},
	}
	service := newServiceForTest(t, mock)
	file := buildTestFileHeader(t, "video.ts", "data")

	dto := CreateCaptureDto{Name: "fail_test", MediaType: "hls"}
	_, err := service.UploadCapture(file, dto)
	if err == nil {
		t.Fatal("expected error from repo")
	}

	savedPath := filepath.Join(dir, "capturas", "fail_test", "video.ts")
	if _, err := os.Stat(savedPath); !os.IsNotExist(err) {
		t.Fatal("expected file to be cleaned up on repo error")
	}
}

func TestServiceGetCaptures(t *testing.T) {
	now := time.Now()
	mock := &repoMock{
		getCapturesFn: func(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureModel], error) {
			return utils.PaginationResponse[CaptureModel]{
				Items: []CaptureModel{
					{ID: 1, Name: "test", CreatedAt: now},
					{ID: 2, Name: "other", CreatedAt: now},
				},
				Pagination: utils.Pagination{Page: 1, PageSize: 10},
			}, nil
		},
	}
	service := newServiceForTest(t, mock)

	result, err := service.GetCaptures(CaptureFilter{}, 1, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
}

func TestServiceGetCapturesError(t *testing.T) {
	mock := &repoMock{
		getCapturesFn: func(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureModel], error) {
			return utils.PaginationResponse[CaptureModel]{}, errors.New("db error")
		},
	}
	service := newServiceForTest(t, mock)

	_, err := service.GetCaptures(CaptureFilter{}, 1, 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceGetCaptureByID(t *testing.T) {
	mock := &repoMock{
		getByIDFn: func(id int) (CaptureModel, error) {
			return CaptureModel{ID: id, Name: "found"}, nil
		},
	}
	service := newServiceForTest(t, mock)

	result, err := service.GetCaptureByID(5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != 5 || result.Name != "found" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestServiceGetCaptureByIDError(t *testing.T) {
	mock := &repoMock{
		getByIDFn: func(id int) (CaptureModel, error) {
			return CaptureModel{}, errors.New("not found")
		},
	}
	service := newServiceForTest(t, mock)

	_, err := service.GetCaptureByID(99)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceDeleteCapture(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "video.ts")
	os.WriteFile(filePath, []byte("data"), 0644)

	mock := &repoMock{
		getByIDFn: func(id int) (CaptureModel, error) {
			return CaptureModel{ID: id, FilePath: filePath}, nil
		},
	}
	service := newServiceForTest(t, mock)

	err := service.DeleteCapture(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatal("expected file to be deleted")
	}
}

func TestServiceDeleteCaptureGetError(t *testing.T) {
	mock := &repoMock{
		getByIDFn: func(id int) (CaptureModel, error) {
			return CaptureModel{}, errors.New("not found")
		},
	}
	service := newServiceForTest(t, mock)

	err := service.DeleteCapture(99)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceDeleteCaptureRepoError(t *testing.T) {
	mock := &repoMock{
		getByIDFn: func(id int) (CaptureModel, error) {
			return CaptureModel{ID: id, FilePath: "/tmp/nonexistent"}, nil
		},
		deleteFn: func(tx *sql.Tx, id int) error {
			return errors.New("db error")
		},
	}
	service := newServiceForTest(t, mock)

	err := service.DeleteCapture(1)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// Tests — sanitizeFileName
// ---------------------------------------------------------------------------

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal_name", "normal_name"},
		{"file/with/slashes", "file_with_slashes"},
		{"file:with:colons", "file_with_colons"},
		{"", "unnamed"},
		{"   ", "unnamed"},
		{"<danger>", "_danger_"},
	}

	for _, tc := range tests {
		result := sanitizeFileName(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeFileName(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// ---------------------------------------------------------------------------
// Tests — buildCaptureDir
// ---------------------------------------------------------------------------

func TestBuildCaptureDir(t *testing.T) {
	orig := config.AppConfig.EntryPoint
	config.AppConfig.EntryPoint = "/data"
	defer func() { config.AppConfig.EntryPoint = orig }()

	result := buildCaptureDir("My Show")
	expected := filepath.Join("/data", "capturas", "My Show")
	if result != expected {
		t.Errorf("buildCaptureDir = %q, want %q", result, expected)
	}
}

// ---------------------------------------------------------------------------
// Tests — saveUploadedFile error
// ---------------------------------------------------------------------------

func TestSaveUploadedFileInvalidPath(t *testing.T) {
	header := &multipart.FileHeader{
		Filename: "test.mp4",
		Size:     0,
		Header:   textproto.MIMEHeader{},
	}

	err := saveUploadedFile(header, "/nonexistent/path/file.mp4")
	if err == nil {
		t.Fatal("expected error for invalid file header")
	}
}
