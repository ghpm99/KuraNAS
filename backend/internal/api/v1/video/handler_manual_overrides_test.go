package video

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type manualOverridesServiceMock struct {
	videoHandlerServiceMock
	classificationResult VideoClassificationDto
	episodeResult        VideoEpisodeDto
	serviceError         error

	receivedFileID         int
	receivedClassification *string
	receivedSeason         *int
	receivedEpisode        *int
}

func (mock *manualOverridesServiceMock) SetVideoClassification(fileID int, classification *string) (VideoClassificationDto, error) {
	mock.receivedFileID = fileID
	mock.receivedClassification = classification
	return mock.classificationResult, mock.serviceError
}

func (mock *manualOverridesServiceMock) SetVideoEpisode(fileID int, season *int, episode *int) (VideoEpisodeDto, error) {
	mock.receivedFileID = fileID
	mock.receivedSeason = season
	mock.receivedEpisode = episode
	return mock.episodeResult, mock.serviceError
}

func putManualOverride(serviceMock *manualOverridesServiceMock, path string, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(serviceMock, &videoFilesServiceMock{}, &videoRecentServiceMock{}, &videoLoggerMock{})
	router := gin.New()
	router.PUT("/video/library/files/:file_id/classification", handler.SetVideoClassificationHandler)
	router.PUT("/video/library/files/:file_id/episode", handler.SetVideoEpisodeHandler)

	request := httptest.NewRequest(http.MethodPut, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestSetVideoClassificationHandlerDecodesBodyAndReturnsEffectiveClassification(t *testing.T) {
	serviceMock := &manualOverridesServiceMock{classificationResult: VideoClassificationDto{FileID: 7, Classification: "series", IsOverride: true}}

	recorder := putManualOverride(serviceMock, "/video/library/files/7/classification", `{"classification":"series"}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if serviceMock.receivedFileID != 7 || serviceMock.receivedClassification == nil || *serviceMock.receivedClassification != "series" {
		t.Fatalf("service received file=%d classification=%v", serviceMock.receivedFileID, serviceMock.receivedClassification)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("body: %v", err)
	}
	if body["file_id"] != float64(7) || body["classification"] != "series" || body["is_override"] != true {
		t.Fatalf("body = %v", body)
	}
}

func TestSetVideoClassificationHandlerTreatsNullAsClear(t *testing.T) {
	serviceMock := &manualOverridesServiceMock{}

	recorder := putManualOverride(serviceMock, "/video/library/files/7/classification", `{"classification":null}`)

	if recorder.Code != http.StatusOK || serviceMock.receivedClassification != nil {
		t.Fatalf("status = %d, classification = %v", recorder.Code, serviceMock.receivedClassification)
	}
}

func TestSetVideoClassificationHandlerRejectsInvalidValueWithTranslatedField(t *testing.T) {
	useTranslationLocale(t, "pt-BR")
	serviceMock := &manualOverridesServiceMock{serviceError: &VideoFieldValidationError{Field: "classification", MessageKey: "ERROR_VIDEO_INVALID_CLASSIFICATION"}}

	recorder := putManualOverride(serviceMock, "/video/library/files/7/classification", `{"classification":"anime"}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("body: %v", err)
	}
	if body["field"] != "classification" || body["error"] != "Classificação de vídeo inválida" {
		t.Fatalf("body = %v", body)
	}
}

func TestSetVideoClassificationHandlerRejectsMalformedRequests(t *testing.T) {
	for _, testCase := range []struct{ path, body string }{
		{"/video/library/files/abc/classification", `{"classification":"movie"}`},
		{"/video/library/files/7/classification", `{"classification":`},
		{"/video/library/files/7/classification", `{"classification":5}`},
	} {
		recorder := putManualOverride(&manualOverridesServiceMock{}, testCase.path, testCase.body)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s %s: status = %d, want 400", testCase.path, testCase.body, recorder.Code)
		}
	}
}

func TestManualOverrideHandlersReturnNotFoundForUnknownVideo(t *testing.T) {
	useTranslationLocale(t, "pt-BR")
	for _, path := range []string{"/video/library/files/999/classification", "/video/library/files/999/episode"} {
		serviceMock := &manualOverridesServiceMock{serviceError: sql.ErrNoRows}

		recorder := putManualOverride(serviceMock, path, `{"season":1,"episode":1,"classification":"movie"}`)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("%s: status = %d, want 404", path, recorder.Code)
		}
	}
}

func TestSetVideoEpisodeHandlerDecodesNumbersAndReturnsEffectiveNumbering(t *testing.T) {
	season, episode := 2, 3
	serviceMock := &manualOverridesServiceMock{episodeResult: VideoEpisodeDto{FileID: 9, Season: &season, Episode: &episode}}

	recorder := putManualOverride(serviceMock, "/video/library/files/9/episode", `{"season":2,"episode":3}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if serviceMock.receivedFileID != 9 || *serviceMock.receivedSeason != 2 || *serviceMock.receivedEpisode != 3 {
		t.Fatalf("service received file=%d season=%v episode=%v", serviceMock.receivedFileID, serviceMock.receivedSeason, serviceMock.receivedEpisode)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("body: %v", err)
	}
	if body["file_id"] != float64(9) || body["season"] != float64(2) || body["episode"] != float64(3) {
		t.Fatalf("body = %v", body)
	}
}

func TestSetVideoEpisodeHandlerTreatsBothNullAsClear(t *testing.T) {
	serviceMock := &manualOverridesServiceMock{}

	recorder := putManualOverride(serviceMock, "/video/library/files/9/episode", `{"season":null,"episode":null}`)

	if recorder.Code != http.StatusOK || serviceMock.receivedSeason != nil || serviceMock.receivedEpisode != nil {
		t.Fatalf("status = %d, season = %v, episode = %v", recorder.Code, serviceMock.receivedSeason, serviceMock.receivedEpisode)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("body: %v", err)
	}
	if _, hasSeason := body["season"]; !hasSeason {
		t.Fatalf("season key must be present as null: %v", body)
	}
}

func TestSetVideoEpisodeHandlerRejectsInvalidNumbersWithField(t *testing.T) {
	useTranslationLocale(t, "pt-BR")
	serviceMock := &manualOverridesServiceMock{serviceError: &VideoFieldValidationError{Field: "episode", MessageKey: "ERROR_VIDEO_INVALID_EPISODE"}}

	recorder := putManualOverride(serviceMock, "/video/library/files/9/episode", `{"season":1,"episode":0}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("body: %v", err)
	}
	if body["field"] != "episode" || body["error"] != "Número de episódio inválido" {
		t.Fatalf("body = %v", body)
	}
}
