package files

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"nas-go/api/internal/roots"
	queries "nas-go/api/pkg/database/queries/files"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

type locationServiceMock struct {
	filesHandlerServiceMock
	requestedLocationIDs []int
	requestedDiskPaths   []string
	location             FileLocationDto
	file                 FileDto
	lookupError          error
}

func (m *locationServiceMock) GetFileLocation(id int) (FileLocationDto, error) {
	m.requestedLocationIDs = append(m.requestedLocationIDs, id)
	return m.location, m.lookupError
}

func (m *locationServiceMock) GetActiveFileByDiskPath(diskPath string) (FileDto, error) {
	m.requestedDiskPaths = append(m.requestedDiskPaths, diskPath)
	return m.file, m.lookupError
}

func newLocationRouter(service *locationServiceMock) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.GET("/files/location/:id", handler.GetFileLocationHandler)
	router.GET("/files/by-disk-path", handler.GetFileByDiskPathHandler)
	return router
}

func performLocationRequest(router *gin.Engine, route string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
	return recorder
}

func TestFileLocationHandlerDecodesIDAndReturnsLocation(t *testing.T) {
	service := &locationServiceMock{location: FileLocationDto{
		FileID: 12, Tier: TierCold, LogicalPath: "/docs/a.txt", DiskPath: "/cold/docs/a.txt",
		LogicalDiskPath: "/ssd/docs/a.txt", RootLabel: "main", ExistsOnDisk: true,
	}}
	recorder := performLocationRequest(newLocationRouter(service), "/files/location/12")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if len(service.requestedLocationIDs) != 1 || service.requestedLocationIDs[0] != 12 {
		t.Fatalf("handler must decode id 12, got %v", service.requestedLocationIDs)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	expectedKeys := []string{"file_id", "tier", "logical_path", "disk_path", "logical_disk_path", "root_label", "exists_on_disk"}
	for _, key := range expectedKeys {
		if _, present := body[key]; !present {
			t.Fatalf("missing key %q in %v", key, body)
		}
	}
	if body["disk_path"] != "/cold/docs/a.txt" || body["tier"] != TierCold || body["exists_on_disk"] != true {
		t.Fatalf("unexpected body %v", body)
	}
}

func TestFileLocationHandlerErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		route       string
		lookupError error
		code        int
	}{
		{name: "invalid id", route: "/files/location/abc", code: http.StatusBadRequest},
		{name: "not found", route: "/files/location/9", lookupError: sql.ErrNoRows, code: http.StatusNotFound},
		{name: "internal", route: "/files/location/9", lookupError: errors.New("boom"), code: http.StatusInternalServerError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := &locationServiceMock{lookupError: tc.lookupError}
			recorder := performLocationRequest(newLocationRouter(service), tc.route)
			if recorder.Code != tc.code {
				t.Fatalf("expected %d, got %d body=%s", tc.code, recorder.Code, recorder.Body.String())
			}
			var body map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] == "" {
				t.Fatalf("expected error body, got %s", recorder.Body.String())
			}
		})
	}
}

func TestFileByDiskPathHandlerDecodesQueryParameter(t *testing.T) {
	service := &locationServiceMock{file: FileDto{ID: 3, Name: "a.txt", Path: "/ssd/docs/a.txt", ParentPath: "/ssd/docs"}}
	diskPath := `D:\Cold Store\docs\a b&c.txt`
	recorder := performLocationRequest(newLocationRouter(service), "/files/by-disk-path?path="+url.QueryEscape(diskPath))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if len(service.requestedDiskPaths) != 1 || service.requestedDiskPaths[0] != diskPath {
		t.Fatalf("handler must decode the path query parameter, got %v", service.requestedDiskPaths)
	}
	var body FileDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body.ID != 3 {
		t.Fatalf("expected FileDto body, got %s err=%v", recorder.Body.String(), err)
	}
}

func TestFileByDiskPathHandlerErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		route       string
		lookupError error
		code        int
	}{
		{name: "missing param", route: "/files/by-disk-path", code: http.StatusBadRequest},
		{name: "blank param", route: "/files/by-disk-path?path=%20%20", code: http.StatusBadRequest},
		{name: "not found", route: "/files/by-disk-path?path=/x", lookupError: sql.ErrNoRows, code: http.StatusNotFound},
		{name: "internal", route: "/files/by-disk-path?path=/x", lookupError: errors.New("boom"), code: http.StatusInternalServerError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := &locationServiceMock{lookupError: tc.lookupError}
			recorder := performLocationRequest(newLocationRouter(service), tc.route)
			if recorder.Code != tc.code {
				t.Fatalf("expected %d, got %d body=%s", tc.code, recorder.Code, recorder.Body.String())
			}
			var body map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] == "" {
				t.Fatalf("expected error body, got %s", recorder.Body.String())
			}
			if tc.code == http.StatusBadRequest && len(service.requestedDiskPaths) != 0 {
				t.Fatalf("service must not be called without a path")
			}
		})
	}
}

func TestServiceGetFileLocationForHotAndColdFiles(t *testing.T) {
	roots.Set([]roots.Root{{ID: 1, Path: "/ssd", Label: "main", Enabled: true}})
	t.Cleanup(roots.Reset)

	hotDir := t.TempDir()
	hotPath := filepath.Join(hotDir, "hot.txt")
	if err := os.WriteFile(hotPath, []byte("x"), 0o600); err != nil {
		t.Fatalf("write hot file: %v", err)
	}
	coldModel := sampleModel(2, "cold.txt", File)
	coldModel.Path = "/ssd/docs/cold.txt"
	coldModel.PhysicalPath = sql.NullString{String: "/missing-cold/docs/cold.txt", Valid: true}
	hotModel := sampleModel(1, "hot.txt", File)
	hotModel.Path = hotPath
	deletedModel := sampleModel(3, "gone.txt", File)
	deletedModel.DeletedAt = sql.NullTime{Time: time.Now(), Valid: true}

	service := newFilesServiceForTest(t, &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			switch id {
			case 1:
				return hotModel, true, nil
			case 2:
				return coldModel, true, nil
			case 3:
				return deletedModel, true, nil
			}
			return FileModel{}, false, nil
		},
	})

	cold, err := service.GetFileLocation(2)
	if err != nil {
		t.Fatalf("cold location: %v", err)
	}
	if cold.Tier != TierCold || cold.DiskPath != "/missing-cold/docs/cold.txt" || cold.LogicalDiskPath != "/ssd/docs/cold.txt" ||
		cold.LogicalPath != "/docs/cold.txt" || cold.RootLabel != "main" || cold.ExistsOnDisk || cold.FileID != 2 {
		t.Fatalf("unexpected cold location %+v", cold)
	}

	hot, err := service.GetFileLocation(1)
	if err != nil {
		t.Fatalf("hot location: %v", err)
	}
	if hot.Tier != TierHot || hot.DiskPath != hotPath || hot.LogicalDiskPath != hotPath || !hot.ExistsOnDisk || hot.RootLabel != "" {
		t.Fatalf("unexpected hot location %+v", hot)
	}

	for _, missingID := range []int{3, 404} {
		if _, err := service.GetFileLocation(missingID); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("id %d: expected sql.ErrNoRows, got %v", missingID, err)
		}
	}
}

func TestServiceGetFileLocationPropagatesRepositoryError(t *testing.T) {
	service := newFilesServiceForTest(t, &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) { return FileModel{}, false, errors.New("db down") },
	})
	if _, err := service.GetFileLocation(1); err == nil || errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}

func TestServiceGetActiveFileByDiskPathNormalizesPath(t *testing.T) {
	var receivedPaths []string
	service := newFilesServiceForTest(t, &filesRepoMock{
		getActiveByPathOrPhysicalFn: func(path string) (FileModel, bool, error) {
			receivedPaths = append(receivedPaths, path)
			if path == filepath.Clean("/cold/docs/a.txt") {
				model := sampleModel(5, "a.txt", File)
				model.PhysicalPath = sql.NullString{String: path, Valid: true}
				return model, true, nil
			}
			return FileModel{}, false, nil
		},
	})

	file, err := service.GetActiveFileByDiskPath("  /cold//docs/./a.txt/  ")
	if err != nil || file.ID != 5 || file.Tier != TierCold {
		t.Fatalf("unexpected file %+v err=%v", file, err)
	}
	if receivedPaths[0] != filepath.Clean("/cold/docs/a.txt") {
		t.Fatalf("path must be cleaned before the lookup, got %q", receivedPaths[0])
	}

	if _, err := service.GetActiveFileByDiskPath("/nope"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if _, err := service.GetActiveFileByDiskPath("   "); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("blank path must not match anything, got %v", err)
	}

	failing := newFilesServiceForTest(t, &filesRepoMock{
		getActiveByPathOrPhysicalFn: func(path string) (FileModel, bool, error) { return FileModel{}, false, errors.New("db down") },
	})
	if _, err := failing.GetActiveFileByDiskPath("/x"); err == nil || errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}

func TestRepositoryGetActiveFileByPathOrPhysicalPath(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetActiveFileByPathOrPhysicalPathQuery)).
		WithArgs("/cold/a").
		WillReturnRows(addFileRow(sqlmock.NewRows(fileRowColumns()), 7, "a", "/ssd/a"))
	mock.ExpectRollback()
	file, found, err := repo.GetActiveFileByPathOrPhysicalPath("/cold/a")
	if err != nil || !found || file.ID != 7 {
		t.Fatalf("expected file 7, got %+v found=%v err=%v", file, found, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetActiveFileByPathOrPhysicalPathQuery)).
		WithArgs("/none").
		WillReturnRows(sqlmock.NewRows(fileRowColumns()))
	mock.ExpectRollback()
	if _, found, err := repo.GetActiveFileByPathOrPhysicalPath("/none"); err != nil || found {
		t.Fatalf("expected not found, found=%v err=%v", found, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetActiveFileByPathOrPhysicalPathQuery)).
		WithArgs("/err").
		WillReturnError(errors.New("query failed"))
	mock.ExpectRollback()
	if _, _, err := repo.GetActiveFileByPathOrPhysicalPath("/err"); err == nil {
		t.Fatalf("expected query error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
