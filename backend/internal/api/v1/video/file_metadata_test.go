package video

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
	queries "nas-go/api/pkg/database/queries/video"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

type videoSummaryServiceMock struct {
	requestedFileIDs []int
	summary          VideoSummaryDto
	lookupFailure    error
}

func (m *videoSummaryServiceMock) GetVideoSummary(fileID int) (VideoSummaryDto, error) {
	m.requestedFileIDs = append(m.requestedFileIDs, fileID)
	return m.summary, m.lookupFailure
}

func performVideoSummaryRequest(service *videoSummaryServiceMock, route string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewVideoSummaryHandler(service, &videoLoggerMock{})
	router := gin.New()
	router.GET("/video/metadata/:file_id", handler.GetVideoSummaryHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
	return recorder
}

func TestVideoSummaryHandlerDecodesFileIDAndReturnsSummary(t *testing.T) {
	service := &videoSummaryServiceMock{summary: VideoSummaryDto{Duration: "120.5"}}

	recorder := performVideoSummaryRequest(service, "/video/metadata/42")

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
	if _, hasKey := body["duration"]; !hasKey {
		t.Fatalf("response must expose the duration key, got %v", body)
	}
}

func TestVideoSummaryHandlerErrorPaths(t *testing.T) {
	cases := []struct {
		name          string
		route         string
		lookupFailure error
		code          int
	}{
		{name: "invalid id", route: "/video/metadata/abc", code: http.StatusBadRequest},
		{name: "not found", route: "/video/metadata/1", lookupFailure: sql.ErrNoRows, code: http.StatusNotFound},
		{name: "internal", route: "/video/metadata/1", lookupFailure: errors.New("db down"), code: http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := &videoSummaryServiceMock{lookupFailure: tc.lookupFailure}
			recorder := performVideoSummaryRequest(service, tc.route)
			if recorder.Code != tc.code {
				t.Fatalf("expected %d, got %d", tc.code, recorder.Code)
			}
			if tc.code == http.StatusBadRequest && len(service.requestedFileIDs) != 0 {
				t.Fatalf("service must not be called with an invalid id")
			}
		})
	}
}

type videoSummaryRepositoryMock struct {
	summary VideoSummaryDto
	failure error
}

func (m *videoSummaryRepositoryMock) GetVideoSummaryByFileID(fileID int) (VideoSummaryDto, error) {
	return m.summary, m.failure
}

func TestVideoSummaryServiceDelegatesToRepository(t *testing.T) {
	expected := VideoSummaryDto{Duration: "120.5"}
	service := NewVideoSummaryService(&videoSummaryRepositoryMock{summary: expected})
	summary, err := service.GetVideoSummary(1)
	if err != nil || summary != expected {
		t.Fatalf("unexpected summary %+v err=%v", summary, err)
	}

	failing := NewVideoSummaryService(&videoSummaryRepositoryMock{failure: sql.ErrNoRows})
	if _, err := failing.GetVideoSummary(1); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestVideoSummaryRepositoryBindsFileID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewVideoSummaryRepository(database.NewDbContext(db))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideoSummaryByFileIDQuery)).
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"duration", "width", "height", "codec_name", "frame_rate", "bit_rate", "audio_codec", "format_name"}).AddRow("120.5", 1280, 720, "h264", 24.0, "1000", "aac", "mov"))
	mock.ExpectRollback()
	summary, err := repo.GetVideoSummaryByFileID(9)
	if err != nil || summary.Duration != "120.5" || summary.CodecName != "h264" || summary.Height != 720 {
		t.Fatalf("unexpected summary %+v err=%v", summary, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideoSummaryByFileIDQuery)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"duration", "width", "height", "codec_name", "frame_rate", "bit_rate", "audio_codec", "format_name"}))
	mock.ExpectRollback()
	if _, err := repo.GetVideoSummaryByFileID(10); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows to survive wrapping, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestPostgres_VideoSummaryByFileID(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_video_summary_it")
	moment := time.Now().UTC().Truncate(time.Second)
	var fileID int
	err := ctx.ExecTx(func(tx *sql.Tx) error {
		created, createErr := files.NewRepository(ctx).CreateFile(tx, files.FileModel{
			Name: "a.mkv", Path: "/srv/a.mkv", ParentPath: "/srv", Format: ".mkv",
			UpdatedAt: moment, CreatedAt: moment, Type: files.File,
		})
		if createErr != nil {
			return createErr
		}
		fileID = created.ID
		_, upsertErr := NewVideoMetadataRepository(ctx).UpsertVideoMetadata(tx, VideoMetadataModel{FileId: fileID, Path: "/srv/a.mkv", Duration: "120.5", Height: 720, CodecName: "h264"})
		return upsertErr
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	repo := NewVideoSummaryRepository(ctx)
	summary, err := repo.GetVideoSummaryByFileID(fileID)
	if err != nil || summary.Duration != "120.5" || summary.CodecName != "h264" || summary.Height != 720 {
		t.Fatalf("unexpected summary %+v err=%v", summary, err)
	}
	if _, err := repo.GetVideoSummaryByFileID(fileID + 1000); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for a file without metadata, got %v", err)
	}
}
