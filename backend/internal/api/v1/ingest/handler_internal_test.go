package ingest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	fetchID       int
	fetchErr      error
	targets       []TargetDto
	presets       []PresetDto
	capturedFetch *FetchRequestDto
}

func (s *fakeService) Fetch(req FetchRequestDto) (int, error) {
	s.capturedFetch = &req
	return s.fetchID, s.fetchErr
}
func (s *fakeService) ListTargets() []TargetDto { return s.targets }
func (s *fakeService) ListPresets() []PresetDto { return s.presets }

func newTestRouter(service ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service)
	router := gin.New()
	router.POST("/ingest/fetch", handler.FetchHandler)
	router.GET("/ingest/targets", handler.GetTargetsHandler)
	router.GET("/ingest/presets", handler.GetPresetsHandler)
	return router
}

func doRequest(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestFetchHandlerSuccess(t *testing.T) {
	router := newTestRouter(&fakeService{fetchID: 7})
	rec := doRequest(router, http.MethodPost, "/ingest/fetch", `{"url":"https://x.test/v","preset":"audio_mp3","target_root":"/srv","subfolder":"m"}`)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
	var resp FetchResponseDto
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if resp.JobID != 7 {
		t.Fatalf("expected job id 7, got %d", resp.JobID)
	}
}

// TestFetchHandlerDecodesPayload pins the request seam: it captures the
// FetchRequestDto the handler decodes (service/ingest.ts → POST /downloads/fetch)
// and asserts every field. A json tag drift fails here instead of silently
// dropping a field of the download request in production.
func TestFetchHandlerDecodesPayload(t *testing.T) {
	service := &fakeService{fetchID: 7}
	router := newTestRouter(service)

	rec := doRequest(router, http.MethodPost, "/ingest/fetch", `{"url":"https://x.test/v","preset":"audio_mp3","target_root":"/srv","subfolder":"musicas"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	if service.capturedFetch == nil {
		t.Fatal("service did not receive the fetch request")
	}
	got := *service.capturedFetch
	want := FetchRequestDto{URL: "https://x.test/v", Preset: "audio_mp3", TargetRoot: "/srv", Subfolder: "musicas"}
	if got != want {
		t.Fatalf("fetch payload decoded into %#v, want %#v", got, want)
	}
}

func TestFetchHandlerBadJSON(t *testing.T) {
	router := newTestRouter(&fakeService{})
	rec := doRequest(router, http.MethodPost, "/ingest/fetch", `{not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestFetchHandlerErrorMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"invalid url", ErrInvalidURL, http.StatusBadRequest},
		{"invalid preset", ErrInvalidPreset, http.StatusBadRequest},
		{"invalid target", ErrInvalidTarget, http.StatusBadRequest},
		{"invalid subfolder", ErrInvalidSubfolder, http.StatusBadRequest},
		{"jobs unavailable", ErrJobsUnavailable, http.StatusServiceUnavailable},
		{"unexpected", errString("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := newTestRouter(&fakeService{fetchErr: tc.err})
			rec := doRequest(router, http.MethodPost, "/ingest/fetch", `{"url":"https://x.test/v","preset":"audio_mp3","target_root":"/srv"}`)
			if rec.Code != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, rec.Code)
			}
		})
	}
}

func TestTargetsAndPresetsHandlers(t *testing.T) {
	router := newTestRouter(&fakeService{
		targets: []TargetDto{{Label: "Midia", Path: "/srv/midia"}},
		presets: []PresetDto{{Key: "audio_mp3", Label: "DOWNLOAD_PRESET_AUDIO_MP3"}},
	})

	rec := doRequest(router, http.MethodGet, "/ingest/targets", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "/srv/midia") {
		t.Fatalf("targets: code %d body %s", rec.Code, rec.Body.String())
	}

	rec = doRequest(router, http.MethodGet, "/ingest/presets", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "audio_mp3") {
		t.Fatalf("presets: code %d body %s", rec.Code, rec.Body.String())
	}
}

type errString string

func (e errString) Error() string { return string(e) }
