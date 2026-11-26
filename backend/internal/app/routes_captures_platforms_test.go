package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/internal/api/v1/captures"
	"nas-go/api/pkg/logger"

	"github.com/gin-gonic/gin"
)

type platformsRouteServiceStub struct {
	captures.ServiceInterface
	isGetByIDCalled bool
}

func (s *platformsRouteServiceStub) GetCapturePlatforms() ([]string, error) {
	return []string{"crunchyroll"}, nil
}

func (s *platformsRouteServiceStub) GetCaptureByID(id int) (captures.CaptureDto, error) {
	s.isGetByIDCalled = true
	return captures.CaptureDto{}, nil
}

type platformsRouteLoggerStub struct{ logger.LoggerServiceInterface }

func (platformsRouteLoggerStub) CreateLog(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	return log, nil
}
func (platformsRouteLoggerStub) CompleteWithSuccessLog(log logger.LoggerModel) error { return nil }

func TestCapturesPlatformsRouteIsNotShadowedByCaptureIDRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &platformsRouteServiceStub{}
	router := gin.New()
	RegisterCapturesRoutes(router.Group("/api/v1"), &AppContext{
		Captures: &CapturesContext{Handler: captures.NewHandler(service, platformsRouteLoggerStub{})},
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/captures/platforms", nil))

	if recorder.Code != http.StatusOK || recorder.Body.String() != `{"items":["crunchyroll"]}` {
		t.Fatalf("unexpected response %d %s", recorder.Code, recorder.Body.String())
	}
	if service.isGetByIDCalled {
		t.Fatal("/captures/platforms reached the /captures/:id handler")
	}
}
