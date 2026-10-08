package files

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"nas-go/api/internal/roots"
	queries "nas-go/api/pkg/database/queries/files"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

type ancestorsServiceMock struct {
	filesHandlerServiceMock
	requestedIDs  []int
	ancestors     []FileAncestorDto
	lookupFailure error
}

func (m *ancestorsServiceMock) GetFileAncestors(id int) ([]FileAncestorDto, error) {
	m.requestedIDs = append(m.requestedIDs, id)
	return m.ancestors, m.lookupFailure
}

func performAncestorsRequest(service *ancestorsServiceMock, route string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.GET("/files/ancestors/:id", handler.GetFileAncestorsHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
	return recorder
}

func TestAncestorsHandlerDecodesIDAndReturnsList(t *testing.T) {
	service := &ancestorsServiceMock{ancestors: []FileAncestorDto{
		{ID: 1, Name: "main", Path: "/", Type: Directory},
		{ID: 4, Name: "docs", Path: "/docs", Type: Directory},
	}}
	recorder := performAncestorsRequest(service, "/files/ancestors/9")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if len(service.requestedIDs) != 1 || service.requestedIDs[0] != 9 {
		t.Fatalf("handler must decode id 9, got %v", service.requestedIDs)
	}
	var body []map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || len(body) != 2 {
		t.Fatalf("expected two ancestors, got %s err=%v", recorder.Body.String(), err)
	}
	for _, field := range []string{"id", "name", "path", "type"} {
		if _, present := body[1][field]; !present {
			t.Fatalf("ancestor must expose %q, got %v", field, body[1])
		}
	}
	if body[1]["id"] != float64(4) || body[1]["path"] != "/docs" {
		t.Fatalf("unexpected ancestor %v", body[1])
	}
}

func TestAncestorsHandlerErrorPaths(t *testing.T) {
	tests := []struct {
		name          string
		route         string
		lookupFailure error
		code          int
	}{
		{name: "invalid id", route: "/files/ancestors/abc", code: http.StatusBadRequest},
		{name: "not found", route: "/files/ancestors/1", lookupFailure: sql.ErrNoRows, code: http.StatusNotFound},
		{name: "internal", route: "/files/ancestors/1", lookupFailure: errors.New("boom"), code: http.StatusInternalServerError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := &ancestorsServiceMock{lookupFailure: tc.lookupFailure}
			recorder := performAncestorsRequest(service, tc.route)
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

func ancestorModel(id int, name string, path string) FileModel {
	model := sampleModel(id, name, Directory)
	model.Path = path
	return model
}

func TestServiceGetFileAncestorsOrdersFromRootAndLabelsRoot(t *testing.T) {
	roots.Set([]roots.Root{
		{ID: 1, Path: "/ssd", Label: "main", Enabled: true},
		{ID: 2, Path: "/disk2", Label: "Midia", Enabled: true},
	})
	t.Cleanup(roots.Reset)

	indexed := []FileModel{
		ancestorModel(30, "fotos", "/disk2/fotos"),
		ancestorModel(10, "disk2", "/disk2"),
		ancestorModel(40, "2024", "/disk2/fotos/2024"),
	}
	var requestedPaths []string
	service := newFilesServiceForTest(t, &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			return ancestorModel(50, "viagem", "/disk2/fotos/2024/viagem"), true, nil
		},
		getActiveFilesByPathsFn: func(paths []string) ([]FileModel, error) {
			requestedPaths = paths
			return indexed, nil
		},
	})

	ancestors, err := service.GetFileAncestors(50)
	if err != nil {
		t.Fatalf("ancestors: %v", err)
	}
	expectedPaths := []string{"/disk2", "/disk2/fotos", "/disk2/fotos/2024"}
	if len(requestedPaths) != len(expectedPaths) {
		t.Fatalf("expected lookup of %v, got %v", expectedPaths, requestedPaths)
	}
	for index, expectedPath := range expectedPaths {
		if requestedPaths[index] != expectedPath {
			t.Fatalf("expected lookup of %v, got %v", expectedPaths, requestedPaths)
		}
	}
	expected := []FileAncestorDto{
		{ID: 10, Name: "Midia", Path: "/Midia", Type: Directory},
		{ID: 30, Name: "fotos", Path: "/Midia/fotos", Type: Directory},
		{ID: 40, Name: "2024", Path: "/Midia/fotos/2024", Type: Directory},
	}
	if len(ancestors) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, ancestors)
	}
	for index := range expected {
		if ancestors[index] != expected[index] {
			t.Fatalf("expected %v, got %v", expected, ancestors)
		}
	}
}

func TestServiceGetFileAncestorsPrimaryRootUsesSlashPath(t *testing.T) {
	roots.Set([]roots.Root{{ID: 1, Path: "/ssd", Label: "main", Enabled: true}})
	t.Cleanup(roots.Reset)

	service := newFilesServiceForTest(t, &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			return ancestorModel(3, "b", "/ssd/a/b"), true, nil
		},
		getActiveFilesByPathsFn: func(paths []string) ([]FileModel, error) {
			return []FileModel{ancestorModel(2, "a", "/ssd/a"), ancestorModel(1, "ssd", "/ssd")}, nil
		},
	})

	ancestors, err := service.GetFileAncestors(3)
	if err != nil || len(ancestors) != 2 {
		t.Fatalf("expected two ancestors, got %v err=%v", ancestors, err)
	}
	if ancestors[0].Path != "/" || ancestors[0].Name != "main" || ancestors[1].Path != "/a" {
		t.Fatalf("unexpected ancestors %v", ancestors)
	}
}

func TestServiceGetFileAncestorsOfRootIsEmptyAndSkipsLookup(t *testing.T) {
	roots.Set([]roots.Root{{ID: 1, Path: "/ssd", Label: "main", Enabled: true}})
	t.Cleanup(roots.Reset)

	lookedUp := false
	service := newFilesServiceForTest(t, &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			return ancestorModel(1, "ssd", "/ssd"), true, nil
		},
		getActiveFilesByPathsFn: func(paths []string) ([]FileModel, error) {
			lookedUp = true
			return nil, nil
		},
	})

	ancestors, err := service.GetFileAncestors(1)
	if err != nil || ancestors == nil || len(ancestors) != 0 || lookedUp {
		t.Fatalf("root must have no ancestors, got %v err=%v lookedUp=%v", ancestors, err, lookedUp)
	}
}

func TestServiceGetFileAncestorsErrors(t *testing.T) {
	deleted := ancestorModel(2, "x", "/ssd/x")
	deleted.DeletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	service := newFilesServiceForTest(t, &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			if id == 2 {
				return deleted, true, nil
			}
			if id == 3 {
				return ancestorModel(3, "b", "/ssd/a/b"), true, nil
			}
			return FileModel{}, false, nil
		},
		getActiveFilesByPathsFn: func(paths []string) ([]FileModel, error) {
			return nil, errors.New("db down")
		},
	})

	for _, missingID := range []int{2, 404} {
		if _, err := service.GetFileAncestors(missingID); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("id %d: expected sql.ErrNoRows, got %v", missingID, err)
		}
	}
	if _, err := service.GetFileAncestors(3); err == nil || errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}

func TestRepositoryGetActiveFilesByPaths(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetActiveFilesByPathsQuery)).
		WillReturnRows(addFileRow(sqlmock.NewRows(fileRowColumns()), 7, "a", "/ssd/a"))
	mock.ExpectRollback()
	found, err := repo.GetActiveFilesByPaths([]string{"/ssd", "/ssd/a"})
	if err != nil || len(found) != 1 || found[0].ID != 7 {
		t.Fatalf("expected file 7, got %+v err=%v", found, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetActiveFilesByPathsQuery)).
		WillReturnError(errors.New("query failed"))
	mock.ExpectRollback()
	if _, err := repo.GetActiveFilesByPaths([]string{"/x"}); err == nil {
		t.Fatalf("expected query error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
