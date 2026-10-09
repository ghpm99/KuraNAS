package image

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
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
	service := &imageSummaryServiceMock{summary: ImageSummaryDto{Width: 1920, Caption: "a dog", Tags: []string{"dog", "park"}, OCRText: "stop"}}

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
	for _, key := range []string{"width", "caption", "tags", "ocr_text"} {
		if _, hasKey := body[key]; !hasKey {
			t.Fatalf("response must expose the %s key, got %v", key, body)
		}
	}
	if tags, isList := body["tags"].([]any); !isList || len(tags) != 2 {
		t.Fatalf("tags must be a list of two entries, got %v", body["tags"])
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
	expected := ImageSummaryDto{Width: 1920, Caption: "a dog", Tags: []string{"dog"}, OCRText: "stop"}
	service := NewImageSummaryService(&imageSummaryRepositoryMock{summary: expected})
	summary, err := service.GetImageSummary(1)
	if err != nil || !reflect.DeepEqual(summary, expected) {
		t.Fatalf("unexpected summary %+v err=%v", summary, err)
	}

	failing := NewImageSummaryService(&imageSummaryRepositoryMock{failure: sql.ErrNoRows})
	if _, err := failing.GetImageSummary(1); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

var imageSummaryColumns = []string{"width", "height", "make", "model", "lens_model", "datetime_original", "exposure_time", "f_number", "iso", "focal_length", "software", "image_description", "taken_at", "gps_latitude", "gps_longitude", "classification_confidence", "classification_suggested_name", "ai_caption", "ai_tags", "ai_ocr_text"}

func TestImageSummaryRepositoryBindsFileID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewImageSummaryRepository(database.NewDbContext(db))
	takenAt := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetImageSummaryByFileIDQuery)).
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows(imageSummaryColumns).AddRow(1920, 1080, "Canon", "R5", "RF50", "2026:01:01", 0.01, 1.8, 200.0, 50.0, "Lightroom", "A trip", takenAt, -23.55, -46.63, 0.8, "sunset-beach", "a beach", "{beach,sunset}", "EXIT"))
	mock.ExpectRollback()
	summary, err := repo.GetImageSummaryByFileID(9)
	if err != nil || summary.Width != 1920 || summary.Make != "Canon" || summary.ISO != 200 {
		t.Fatalf("unexpected summary %+v err=%v", summary, err)
	}
	if summary.Software != "Lightroom" || summary.Description != "A trip" || summary.TakenAt == nil || !summary.TakenAt.Equal(takenAt) {
		t.Fatalf("unexpected extra fields %+v", summary)
	}
	if summary.GPSLatitude == nil || *summary.GPSLatitude != -23.55 || summary.GPSLongitude == nil || *summary.GPSLongitude != -46.63 {
		t.Fatalf("unexpected gps %+v", summary)
	}
	if summary.Caption != "a beach" || !reflect.DeepEqual(summary.Tags, []string{"beach", "sunset"}) || summary.OCRText != "EXIT" {
		t.Fatalf("unexpected content fields %+v", summary)
	}
	if summary.ClassificationConfidence != 0.8 || summary.SuggestedName != "sunset-beach" {
		t.Fatalf("unexpected ai fields %+v", summary)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetImageSummaryByFileIDQuery)).
		WithArgs(11).
		WillReturnRows(sqlmock.NewRows(imageSummaryColumns).AddRow(1, 1, "", "", "", "", 0, 0, 0, 0, "", "", nil, nil, nil, 0, "", "", "{}", ""))
	mock.ExpectRollback()
	withoutGPS, err := repo.GetImageSummaryByFileID(11)
	if err != nil || withoutGPS.GPSLatitude != nil || withoutGPS.GPSLongitude != nil || withoutGPS.TakenAt != nil {
		t.Fatalf("absent gps and date must stay nil, got %+v err=%v", withoutGPS, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetImageSummaryByFileIDQuery)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows(imageSummaryColumns))
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

func TestPostgres_ImageSummaryGPSTreatsZeroZeroAsAbsent(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_image_summary_gps_it")
	moment := time.Now().UTC().Truncate(time.Second)
	idsByName := map[string]int{}
	err := ctx.ExecTx(func(tx *sql.Tx) error {
		for _, seed := range []struct {
			name      string
			latitude  float64
			longitude float64
		}{{"rio.jpg", -22.9, -43.2}, {"zero.jpg", 0, 0}, {"equator.jpg", 0, -43.2}} {
			created, createErr := files.NewRepository(ctx).CreateFile(tx, files.FileModel{
				Name: seed.name, Path: "/srv/" + seed.name, ParentPath: "/srv", Format: ".jpg",
				UpdatedAt: moment, CreatedAt: moment, Type: files.File,
			})
			if createErr != nil {
				return createErr
			}
			idsByName[seed.name] = created.ID
			if _, upsertErr := NewRepository(ctx).UpsertImageMetadata(tx, MetadataModel{
				FileId: created.ID, Path: "/srv/" + seed.name, GPSLatitude: seed.latitude, GPSLongitude: seed.longitude,
				Software: "Editor", Classification: ClassificationModel{Confidence: 0.5, SuggestedName: "suggestion"},
			}); upsertErr != nil {
				return upsertErr
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	repo := NewImageSummaryRepository(ctx)
	rio, err := repo.GetImageSummaryByFileID(idsByName["rio.jpg"])
	if err != nil || rio.GPSLatitude == nil || rio.GPSLongitude == nil || *rio.GPSLongitude > -43.1 || *rio.GPSLongitude < -43.3 {
		t.Fatalf("expected rio coordinates, got %+v err=%v", rio, err)
	}
	if rio.Software != "Editor" || rio.SuggestedName != "suggestion" || rio.ClassificationConfidence != 0.5 {
		t.Fatalf("unexpected extra fields %+v", rio)
	}
	zero, err := repo.GetImageSummaryByFileID(idsByName["zero.jpg"])
	if err != nil || zero.GPSLatitude != nil || zero.GPSLongitude != nil {
		t.Fatalf("exact 0,0 must be absent, got %+v err=%v", zero, err)
	}
	equator, err := repo.GetImageSummaryByFileID(idsByName["equator.jpg"])
	if err != nil || equator.GPSLatitude == nil || *equator.GPSLatitude != 0 {
		t.Fatalf("a real coordinate with a zero axis must be kept, got %+v err=%v", equator, err)
	}
}
