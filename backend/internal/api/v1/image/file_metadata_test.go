package image

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/image"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

type imageSummaryServiceMock struct {
	requestedFileIDs []int
	summary          ImageSummaryDto
	lookupFailure    error
}

func (m *imageSummaryServiceMock) GetImageSummary(fileID int) (ImageSummaryDto, error) {
	m.requestedFileIDs = append(m.requestedFileIDs, fileID)
	return m.summary, m.lookupFailure
}

func performImageSummaryRequest(service *imageSummaryServiceMock, route string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewImageSummaryHandler(service, &imageLoggerMock{})
	router := gin.New()
	router.GET("/image/metadata/:file_id", handler.GetImageSummaryHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
	return recorder
}

func TestImageSummaryHandlerDecodesFileIDAndReturnsSummary(t *testing.T) {
	service := &imageSummaryServiceMock{summary: ImageSummaryDto{Width: 1920}}

	recorder := performImageSummaryRequest(service, "/image/metadata/42")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if len(service.requestedFileIDs) != 1 || service.requestedFileIDs[0] != 42 {
		t.Fatalf("handler must decode file_id 42, got %v", service.requestedFileIDs)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if _, hasKey := body["width"]; !hasKey {
		t.Fatalf("response must expose the width key, got %v", body)
	}
}

func TestImageSummaryHandlerErrorPaths(t *testing.T) {
	cases := []struct {
		name          string
		route         string
		lookupFailure error
		code          int
	}{
		{name: "invalid id", route: "/image/metadata/abc", code: http.StatusBadRequest},
		{name: "not found", route: "/image/metadata/1", lookupFailure: sql.ErrNoRows, code: http.StatusNotFound},
		{name: "internal", route: "/image/metadata/1", lookupFailure: errors.New("db down"), code: http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := &imageSummaryServiceMock{lookupFailure: tc.lookupFailure}
			recorder := performImageSummaryRequest(service, tc.route)
			if recorder.Code != tc.code {
				t.Fatalf("expected %d, got %d", tc.code, recorder.Code)
			}
			if tc.code == http.StatusBadRequest && len(service.requestedFileIDs) != 0 {
				t.Fatalf("service must not be called with an invalid id")
			}
		})
	}
}

type imageSummaryRepositoryMock struct {
	summary ImageSummaryDto
	failure error
}

func (m *imageSummaryRepositoryMock) GetImageSummaryByFileID(fileID int) (ImageSummaryDto, error) {
	return m.summary, m.failure
}

func TestImageSummaryServiceDelegatesToRepository(t *testing.T) {
	expected := ImageSummaryDto{Width: 1920}
	service := NewImageSummaryService(&imageSummaryRepositoryMock{summary: expected})
	summary, err := service.GetImageSummary(1)
	if err != nil || summary != expected {
		t.Fatalf("unexpected summary %+v err=%v", summary, err)
	}

	failing := NewImageSummaryService(&imageSummaryRepositoryMock{failure: sql.ErrNoRows})
	if _, err := failing.GetImageSummary(1); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestImageSummaryRepositoryBindsFileID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewImageSummaryRepository(database.NewDbContext(db))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetImageSummaryByFileIDQuery)).
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"width", "height", "make", "model", "lens_model", "datetime_original", "exposure_time", "f_number", "iso", "focal_length"}).AddRow(1920, 1080, "Canon", "R5", "RF50", "2026:01:01", 0.01, 1.8, 200.0, 50.0))
	mock.ExpectRollback()
	summary, err := repo.GetImageSummaryByFileID(9)
	if err != nil || summary.Width != 1920 || summary.Make != "Canon" || summary.ISO != 200 {
		t.Fatalf("unexpected summary %+v err=%v", summary, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetImageSummaryByFileIDQuery)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"width", "height", "make", "model", "lens_model", "datetime_original", "exposure_time", "f_number", "iso", "focal_length"}))
	mock.ExpectRollback()
	if _, err := repo.GetImageSummaryByFileID(10); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows to survive wrapping, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestPostgres_ImageSummaryByFileID(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_image_summary_it")
	moment := time.Now().UTC().Truncate(time.Second)
	var fileID int
	err := ctx.ExecTx(func(tx *sql.Tx) error {
		created, createErr := files.NewRepository(ctx).CreateFile(tx, files.FileModel{
			Name: "a.jpg", Path: "/srv/a.jpg", ParentPath: "/srv", Format: ".jpg",
			UpdatedAt: moment, CreatedAt: moment, Type: files.File,
		})
		if createErr != nil {
			return createErr
		}
		fileID = created.ID
		_, upsertErr := NewRepository(ctx).UpsertImageMetadata(tx, MetadataModel{FileId: fileID, Path: "/srv/a.jpg", Width: 1920, Height: 1080, Make: "Canon", ISO: 200})
		return upsertErr
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	repo := NewImageSummaryRepository(ctx)
	summary, err := repo.GetImageSummaryByFileID(fileID)
	if err != nil || summary.Width != 1920 || summary.Make != "Canon" || summary.ISO != 200 {
		t.Fatalf("unexpected summary %+v err=%v", summary, err)
	}
	if _, err := repo.GetImageSummaryByFileID(fileID + 1000); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for a file without metadata, got %v", err)
	}
}
