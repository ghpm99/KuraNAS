package files

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	queries "nas-go/api/pkg/database/queries/files"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

type folderStatsServiceMock struct {
	filesHandlerServiceMock
	requestedIDs  []int
	stats         FolderStatsDto
	lookupFailure error
}

func (m *folderStatsServiceMock) GetFolderStats(id int) (FolderStatsDto, error) {
	m.requestedIDs = append(m.requestedIDs, id)
	return m.stats, m.lookupFailure
}

func performFolderStatsRequest(service *folderStatsServiceMock, route string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.GET("/files/folder-stats/:id", handler.GetFolderStatsHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
	return recorder
}

func TestFolderStatsHandlerDecodesIDAndReturnsStats(t *testing.T) {
	service := &folderStatsServiceMock{stats: FolderStatsDto{FileCount: 3, FolderCount: 2, TotalSizeBytes: 900}}

	recorder := performFolderStatsRequest(service, "/files/folder-stats/12")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if len(service.requestedIDs) != 1 || service.requestedIDs[0] != 12 {
		t.Fatalf("handler must decode id 12, got %v", service.requestedIDs)
	}
	var body map[string]int64
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["file_count"] != 3 || body["folder_count"] != 2 || body["total_size_bytes"] != 900 {
		t.Fatalf("unexpected body %v", body)
	}
}

func TestFolderStatsHandlerErrorPaths(t *testing.T) {
	cases := []struct {
		name          string
		route         string
		lookupFailure error
		code          int
	}{
		{name: "invalid id", route: "/files/folder-stats/abc", code: http.StatusBadRequest},
		{name: "not found", route: "/files/folder-stats/1", lookupFailure: sql.ErrNoRows, code: http.StatusNotFound},
		{name: "internal", route: "/files/folder-stats/1", lookupFailure: errors.New("db down"), code: http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := &folderStatsServiceMock{lookupFailure: tc.lookupFailure}
			recorder := performFolderStatsRequest(service, tc.route)
			if recorder.Code != tc.code {
				t.Fatalf("expected %d, got %d", tc.code, recorder.Code)
			}
			if tc.code == http.StatusBadRequest && len(service.requestedIDs) != 0 {
				t.Fatalf("service must not be called with an invalid id")
			}
		})
	}
}

func TestServiceGetFolderStatsQueriesDescendantPrefixOfActiveFolder(t *testing.T) {
	folderModel := sampleModel(1, "docs", Directory)
	folderModel.Path = "/srv/docs"
	fileModel := sampleModel(2, "a.txt", File)
	deletedFolderModel := sampleModel(3, "gone", Directory)
	deletedFolderModel.DeletedAt = sql.NullTime{Time: time.Now(), Valid: true}

	var receivedPrefixes []string
	service := newFilesServiceForTest(t, &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			switch id {
			case 1:
				return folderModel, true, nil
			case 2:
				return fileModel, true, nil
			case 3:
				return deletedFolderModel, true, nil
			}
			return FileModel{}, false, nil
		},
		getFolderStatsFn: func(prefix string) (FolderStatsDto, error) {
			receivedPrefixes = append(receivedPrefixes, prefix)
			return FolderStatsDto{FileCount: 4, FolderCount: 1, TotalSizeBytes: 77}, nil
		},
	})

	stats, err := service.GetFolderStats(1)
	if err != nil || stats.FileCount != 4 || stats.FolderCount != 1 || stats.TotalSizeBytes != 77 {
		t.Fatalf("unexpected stats %+v err=%v", stats, err)
	}
	if len(receivedPrefixes) != 1 || receivedPrefixes[0] != "/srv/docs"+string(filepath.Separator) {
		t.Fatalf("prefix must end with the separator, got %v", receivedPrefixes)
	}

	for _, rejectedID := range []int{2, 3, 404} {
		if _, err := service.GetFolderStats(rejectedID); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("id %d: expected sql.ErrNoRows, got %v", rejectedID, err)
		}
	}
	if len(receivedPrefixes) != 1 {
		t.Fatalf("repository must not be queried for rejected ids")
	}
}

func TestServiceGetFolderStatsPropagatesRepositoryErrors(t *testing.T) {
	folderModel := sampleModel(1, "docs", Directory)
	failingStats := newFilesServiceForTest(t, &filesRepoMock{
		getFileByIDFn:    func(id int) (FileModel, bool, error) { return folderModel, true, nil },
		getFolderStatsFn: func(string) (FolderStatsDto, error) { return FolderStatsDto{}, errors.New("db down") },
	})
	if _, err := failingStats.GetFolderStats(1); err == nil || errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected wrapped stats error, got %v", err)
	}

	failingLookup := newFilesServiceForTest(t, &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) { return FileModel{}, false, errors.New("db down") },
	})
	if _, err := failingLookup.GetFolderStats(1); err == nil || errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected wrapped lookup error, got %v", err)
	}
}

func TestRepositoryGetFolderStatsBindsPrefixAndTypes(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetFolderStatsQuery)).
		WithArgs("/srv/docs/", File, Directory).
		WillReturnRows(sqlmock.NewRows([]string{"file_count", "folder_count", "total_size_bytes"}).AddRow(5, 2, 1234))
	mock.ExpectRollback()
	stats, err := repo.GetFolderStats("/srv/docs/")
	if err != nil || stats.FileCount != 5 || stats.FolderCount != 2 || stats.TotalSizeBytes != 1234 {
		t.Fatalf("unexpected stats %+v err=%v", stats, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetFolderStatsQuery)).
		WithArgs("/boom/", File, Directory).
		WillReturnError(errors.New("query failed"))
	mock.ExpectRollback()
	if _, err := repo.GetFolderStats("/boom/"); err == nil {
		t.Fatalf("expected query error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
