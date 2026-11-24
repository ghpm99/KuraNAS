package ingest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	fetchResponse FetchResponseDto
	fetchErr      error
	targets       []TargetDto
	presets       []PresetDto
	capturedFetch *FetchRequestDto
}

func (s *fakeService) Fetch(req FetchRequestDto) (FetchResponseDto, error) {
	s.capturedFetch = &req
	return s.fetchResponse, s.fetchErr
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
	router := newTestRouter(&fakeService{fetchResponse: FetchResponseDto{JobID: 7, DownloadID: 12}})
	rec := doRequest(router, http.MethodPost, "/ingest/fetch", `{"url":"https://x.test/v","preset":"audio_mp3","target_root":"/srv","subfolder":"m"}`)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
	var resp FetchResponseDto
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if resp.JobID != 7 || resp.DownloadID != 12 {
		t.Fatalf("expected job_id 7 and download_id 12, got %+v", resp)
	}
}

// TestFetchHandlerDecodesPayload pins the request seam: it captures the
// FetchRequestDto the handler decodes (service/ingest.ts → POST /ingest/fetch)
// and asserts every field. A json tag drift fails here instead of silently
// dropping a field of the download request in production.
func TestFetchHandlerDecodesPayload(t *testing.T) {
	service := &fakeService{fetchResponse: FetchResponseDto{JobID: 7, DownloadID: 12}}
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
		{"url too long", ErrURLTooLong, http.StatusBadRequest},
		{"invalid preset", ErrInvalidPreset, http.StatusBadRequest},
		{"invalid target", ErrInvalidTarget, http.StatusBadRequest},
		{"invalid subfolder", ErrInvalidSubfolder, http.StatusBadRequest},
		{"jobs unavailable", ErrJobsUnavailable, http.StatusServiceUnavailable},
		{"yt-dlp not installed", ErrYtDlpNotInstalled, http.StatusServiceUnavailable},
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

func fetchErrorMessage(t *testing.T, serviceErr error) string {
	t.Helper()
	router := newTestRouter(&fakeService{fetchErr: serviceErr})
	rec := doRequest(router, http.MethodPost, "/ingest/fetch", `{"url":"https://x.test/v","preset":"audio_mp3","target_root":"/srv"}`)
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return body["error"]
}

func TestFetchHandlerTellsAnInvalidSubfolderApartFromAnInvalidTarget(t *testing.T) {
	subfolderMessage := fetchErrorMessage(t, ErrInvalidSubfolder)
	targetMessage := fetchErrorMessage(t, ErrInvalidTarget)

	if subfolderMessage == "" || subfolderMessage == targetMessage {
		t.Fatalf("subfolder and target need distinct messages, got %q and %q", subfolderMessage, targetMessage)
	}
}

func loadTranslationsForTest(t *testing.T) {
	t.Helper()
	if err := i18n.LoadTranslations(); err != nil {
		t.Fatalf("load translations: %v", err)
	}
}

func assertTranslatedYtDlpMessage(t *testing.T, body string) {
	t.Helper()
	if strings.Contains(body, "ERROR_DOWNLOAD_YTDLP_NOT_INSTALLED") || !strings.Contains(body, "yt-dlp") {
		t.Fatalf("expected translated message, got %s", body)
	}
}

func TestFetchHandlerYtDlpNotInstalledReturnsTranslatedMessage(t *testing.T) {
	loadTranslationsForTest(t)
	router := newTestRouter(&fakeService{fetchErr: ErrYtDlpNotInstalled})
	rec := doRequest(router, http.MethodPost, "/ingest/fetch", `{"url":"https://x.test/v","preset":"audio_mp3","target_root":"/srv"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	assertTranslatedYtDlpMessage(t, rec.Body.String())
}
