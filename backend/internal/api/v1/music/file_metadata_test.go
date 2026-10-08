package music

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
	queries "nas-go/api/pkg/database/queries/music"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

type musicSummaryServiceMock struct {
	requestedFileIDs []int
	summary          AudioSummaryDto
	lookupFailure    error
}

func (m *musicSummaryServiceMock) GetAudioSummary(fileID int) (AudioSummaryDto, error) {
	m.requestedFileIDs = append(m.requestedFileIDs, fileID)
	return m.summary, m.lookupFailure
}

func performAudioSummaryRequest(service *musicSummaryServiceMock, route string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewAudioSummaryHandler(service, &musicLoggerMock{})
	router := gin.New()
	router.GET("/music/metadata/:file_id", handler.GetAudioSummaryHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
	return recorder
}

func TestAudioSummaryHandlerDecodesFileIDAndReturnsSummary(t *testing.T) {
	service := &musicSummaryServiceMock{summary: AudioSummaryDto{Title: "T"}}

	recorder := performAudioSummaryRequest(service, "/music/metadata/42")

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
	if _, hasKey := body["title"]; !hasKey {
		t.Fatalf("response must expose the title key, got %v", body)
	}
}

func TestAudioSummaryHandlerErrorPaths(t *testing.T) {
	cases := []struct {
		name          string
		route         string
		lookupFailure error
		code          int
	}{
		{name: "invalid id", route: "/music/metadata/abc", code: http.StatusBadRequest},
		{name: "not found", route: "/music/metadata/1", lookupFailure: sql.ErrNoRows, code: http.StatusNotFound},
		{name: "internal", route: "/music/metadata/1", lookupFailure: errors.New("db down"), code: http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := &musicSummaryServiceMock{lookupFailure: tc.lookupFailure}
			recorder := performAudioSummaryRequest(service, tc.route)
			if recorder.Code != tc.code {
				t.Fatalf("expected %d, got %d", tc.code, recorder.Code)
			}
			if tc.code == http.StatusBadRequest && len(service.requestedFileIDs) != 0 {
				t.Fatalf("service must not be called with an invalid id")
			}
		})
	}
}

type musicSummaryRepositoryMock struct {
	summary AudioSummaryDto
	failure error
}

func (m *musicSummaryRepositoryMock) GetAudioSummaryByFileID(fileID int) (AudioSummaryDto, error) {
	return m.summary, m.failure
}

func TestAudioSummaryServiceDelegatesToRepository(t *testing.T) {
	expected := AudioSummaryDto{Title: "T"}
	service := NewAudioSummaryService(&musicSummaryRepositoryMock{summary: expected})
	summary, err := service.GetAudioSummary(1)
	if err != nil || summary != expected {
		t.Fatalf("unexpected summary %+v err=%v", summary, err)
	}

	failing := NewAudioSummaryService(&musicSummaryRepositoryMock{failure: sql.ErrNoRows})
	if _, err := failing.GetAudioSummary(1); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestAudioSummaryRepositoryBindsFileID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := NewAudioSummaryRepository(database.NewDbContext(db))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAudioSummaryByFileIDQuery)).
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"title", "artist", "album", "genre", "year", "track_number", "length", "bitrate", "sample_rate", "channels"}).AddRow("T", "A", "Al", "G", "2020", "3", 181.5, 320, 44100, 2))
	mock.ExpectRollback()
	summary, err := repo.GetAudioSummaryByFileID(9)
	if err != nil || summary.Title != "T" || summary.LengthSeconds != 181.5 || summary.Bitrate != 320 {
		t.Fatalf("unexpected summary %+v err=%v", summary, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAudioSummaryByFileIDQuery)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"title", "artist", "album", "genre", "year", "track_number", "length", "bitrate", "sample_rate", "channels"}))
	mock.ExpectRollback()
	if _, err := repo.GetAudioSummaryByFileID(10); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows to survive wrapping, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestPostgres_AudioSummaryByFileID(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_music_summary_it")
	moment := time.Now().UTC().Truncate(time.Second)
	var fileID int
	err := ctx.ExecTx(func(tx *sql.Tx) error {
		created, createErr := files.NewRepository(ctx).CreateFile(tx, files.FileModel{
			Name: "a.mp3", Path: "/srv/a.mp3", ParentPath: "/srv", Format: ".mp3",
			UpdatedAt: moment, CreatedAt: moment, Type: files.File,
		})
		if createErr != nil {
			return createErr
		}
		fileID = created.ID
		_, upsertErr := NewAudioMetadataRepository(ctx).UpsertAudioMetadata(tx, AudioMetadataModel{FileId: fileID, Path: "/srv/a.mp3", Title: "T", Length: 181.5, Bitrate: 320})
		return upsertErr
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	repo := NewAudioSummaryRepository(ctx)
	summary, err := repo.GetAudioSummaryByFileID(fileID)
	if err != nil || summary.Title != "T" || summary.LengthSeconds != 181.5 || summary.Bitrate != 320 {
		t.Fatalf("unexpected summary %+v err=%v", summary, err)
	}
	if _, err := repo.GetAudioSummaryByFileID(fileID + 1000); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for a file without metadata, got %v", err)
	}
}
