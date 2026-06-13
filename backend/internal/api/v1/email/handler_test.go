package email

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type mockService struct {
	listFn         func() ([]AccountDto, error)
	deleteFn       func(id int) error
	syncFn         func(id int, enabled bool) error
	authURLFn      func() (GoogleAuthURLDto, error)
	callbackFn     func(state, code string) error
	deviceCodeFn   func() (DeviceCodeDto, error)
	deviceStatusFn func() DeviceCodeStatusDto
	listMessagesFn func(page, pageSize int) (utils.PaginationResponse[MessageDto], error)
	enqueueSyncFn  func(accountID int) (int, error)
	analysisFn     func(messageID int) (AnalysisDto, error)
	getProviderFn  func() (ProviderPreferenceDto, error)
	setProviderFn  func(provider string) (ProviderPreferenceDto, error)
}

func (m *mockService) ListAccounts() ([]AccountDto, error) { return m.listFn() }
func (m *mockService) DeleteAccount(id int) error          { return m.deleteFn(id) }
func (m *mockService) SetSyncEnabled(id int, enabled bool) error {
	return m.syncFn(id, enabled)
}
func (m *mockService) GoogleAuthURL() (GoogleAuthURLDto, error) { return m.authURLFn() }
func (m *mockService) HandleGoogleCallback(state string, code string) error {
	return m.callbackFn(state, code)
}
func (m *mockService) StartMicrosoftDeviceCode() (DeviceCodeDto, error) { return m.deviceCodeFn() }
func (m *mockService) MicrosoftDeviceCodeStatus() DeviceCodeStatusDto   { return m.deviceStatusFn() }
func (m *mockService) ValidAccessToken(accountID int) (string, error)   { return "", nil }
func (m *mockService) ListMessages(page, pageSize int) (utils.PaginationResponse[MessageDto], error) {
	return m.listMessagesFn(page, pageSize)
}
func (m *mockService) EnqueueSync(accountID int) (int, error)  { return m.enqueueSyncFn(accountID) }
func (m *mockService) SyncEnabledAccounts() (SyncStats, error) { return SyncStats{}, nil }
func (m *mockService) PrefilterPending() (int, error)          { return 0, nil }
func (m *mockService) PurgeExpired() (int, error)              { return 0, nil }
func (m *mockService) AnalyzePending() (AnalyzeStats, error)   { return AnalyzeStats{}, nil }
func (m *mockService) GetMessageAnalysis(messageID int) (AnalysisDto, error) {
	return m.analysisFn(messageID)
}
func (m *mockService) GetProviderPreference() (ProviderPreferenceDto, error) {
	return m.getProviderFn()
}
func (m *mockService) SetProviderPreference(provider string) (ProviderPreferenceDto, error) {
	return m.setProviderFn(provider)
}

func newTestRouter(service ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(service)

	router.GET("/email/accounts", handler.GetAccountsHandler)
	router.DELETE("/email/accounts/:id", handler.DeleteAccountHandler)
	router.PUT("/email/accounts/:id/sync-enabled", handler.UpdateSyncEnabledHandler)
	router.POST("/email/accounts/google/auth-url", handler.GoogleAuthURLHandler)
	router.GET("/email/oauth/google/callback", handler.GoogleCallbackHandler)
	router.POST("/email/accounts/microsoft/device-code", handler.MicrosoftDeviceCodeHandler)
	router.GET("/email/accounts/microsoft/device-code/status", handler.MicrosoftDeviceCodeStatusHandler)
	router.POST("/email/accounts/:id/sync", handler.SyncAccountHandler)
	router.GET("/email/messages", handler.GetMessagesHandler)
	router.GET("/email/messages/:id/summary", handler.GetMessageSummaryHandler)
	router.GET("/email/settings/provider", handler.GetProviderHandler)
	router.PUT("/email/settings/provider", handler.SetProviderHandler)
	return router
}

func performRequest(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestGetAccountsHandler(t *testing.T) {
	router := newTestRouter(&mockService{
		listFn: func() ([]AccountDto, error) {
			return []AccountDto{{ID: 1, Provider: "google", Address: "a@gmail.com"}}, nil
		},
	})

	response := performRequest(router, http.MethodGet, "/email/accounts", "")
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var accounts []AccountDto
	if err := json.Unmarshal(response.Body.Bytes(), &accounts); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(accounts) != 1 || accounts[0].Address != "a@gmail.com" {
		t.Fatalf("unexpected accounts: %+v", accounts)
	}
	if strings.Contains(response.Body.String(), "token") {
		t.Fatal("response must not mention tokens")
	}
}

func TestDeleteAccountHandlerNotFound(t *testing.T) {
	router := newTestRouter(&mockService{
		deleteFn: func(id int) error { return ErrAccountNotFound },
	})

	response := performRequest(router, http.MethodDelete, "/email/accounts/9", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestDeleteAccountHandlerInvalidID(t *testing.T) {
	router := newTestRouter(&mockService{
		deleteFn: func(id int) error { t.Fatal("must not be called"); return nil },
	})

	response := performRequest(router, http.MethodDelete, "/email/accounts/abc", "")
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestUpdateSyncEnabledHandler(t *testing.T) {
	var gotID int
	var gotEnabled bool
	router := newTestRouter(&mockService{
		syncFn: func(id int, enabled bool) error {
			gotID, gotEnabled = id, enabled
			return nil
		},
	})

	response := performRequest(router, http.MethodPut, "/email/accounts/3/sync-enabled", `{"sync_enabled":false}`)
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", response.Code)
	}
	if gotID != 3 || gotEnabled != false {
		t.Fatalf("unexpected call: id=%d enabled=%v", gotID, gotEnabled)
	}
}

func TestGoogleAuthURLHandlerNotConfigured(t *testing.T) {
	router := newTestRouter(&mockService{
		authURLFn: func() (GoogleAuthURLDto, error) { return GoogleAuthURLDto{}, ErrProviderNotConfigured },
	})

	response := performRequest(router, http.MethodPost, "/email/accounts/google/auth-url", "")
	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", response.Code)
	}
}

func TestGoogleCallbackHandlerSuccessRendersHTML(t *testing.T) {
	var gotState, gotCode string
	router := newTestRouter(&mockService{
		callbackFn: func(state, code string) error {
			gotState, gotCode = state, code
			return nil
		},
	})

	response := performRequest(router, http.MethodGet, "/email/oauth/google/callback?state=s1&code=c1", "")
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !strings.Contains(response.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("expected html, got %s", response.Header().Get("Content-Type"))
	}
	if gotState != "s1" || gotCode != "c1" {
		t.Fatalf("unexpected callback args: %s %s", gotState, gotCode)
	}
}

func TestGoogleCallbackHandlerProviderError(t *testing.T) {
	router := newTestRouter(&mockService{
		callbackFn: func(state, code string) error { t.Fatal("must not be called"); return nil },
	})

	response := performRequest(router, http.MethodGet, "/email/oauth/google/callback?error=access_denied", "")
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestMicrosoftDeviceCodeHandler(t *testing.T) {
	router := newTestRouter(&mockService{
		deviceCodeFn: func() (DeviceCodeDto, error) {
			return DeviceCodeDto{UserCode: "ABC123", VerificationURI: "https://microsoft.com/devicelogin", ExpiresIn: 900}, nil
		},
	})

	response := performRequest(router, http.MethodPost, "/email/accounts/microsoft/device-code", "")
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var dto DeviceCodeDto
	if err := json.Unmarshal(response.Body.Bytes(), &dto); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if dto.UserCode != "ABC123" || dto.Message == "" {
		t.Fatalf("unexpected dto: %+v", dto)
	}
}

func TestMicrosoftDeviceCodeStatusHandler(t *testing.T) {
	router := newTestRouter(&mockService{
		deviceStatusFn: func() DeviceCodeStatusDto { return DeviceCodeStatusDto{Status: DeviceCodePending} },
	})

	response := performRequest(router, http.MethodGet, "/email/accounts/microsoft/device-code/status", "")
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), DeviceCodePending) {
		t.Fatalf("unexpected body: %s", response.Body.String())
	}
}

func TestGetMessagesHandler(t *testing.T) {
	router := newTestRouter(&mockService{
		listMessagesFn: func(page, pageSize int) (utils.PaginationResponse[MessageDto], error) {
			if page != 2 || pageSize != 25 {
				t.Fatalf("unexpected paging: page=%d size=%d", page, pageSize)
			}
			return utils.PaginationResponse[MessageDto]{
				Items:      []MessageDto{{ID: 7, Subject: "Hi", SenderAddress: "a@example.com"}},
				Pagination: utils.Pagination{Page: page, PageSize: pageSize},
			}, nil
		},
	})

	response := performRequest(router, http.MethodGet, "/email/messages?page=2&page_size=25", "")
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var page utils.PaginationResponse[MessageDto]
	if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != 7 {
		t.Fatalf("unexpected items: %+v", page.Items)
	}
	// The lean DTO must never carry a body field.
	if strings.Contains(response.Body.String(), "sanitized_body") || strings.Contains(response.Body.String(), "\"body\"") {
		t.Fatalf("listing must not include a body: %s", response.Body.String())
	}
}

func TestSyncAccountHandlerAccepted(t *testing.T) {
	router := newTestRouter(&mockService{
		enqueueSyncFn: func(accountID int) (int, error) {
			if accountID != 4 {
				t.Fatalf("unexpected account id: %d", accountID)
			}
			return 42, nil
		},
	})

	response := performRequest(router, http.MethodPost, "/email/accounts/4/sync", "")
	if response.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "42") {
		t.Fatalf("expected job id in body: %s", response.Body.String())
	}
}

func TestSyncAccountHandlerNotFound(t *testing.T) {
	router := newTestRouter(&mockService{
		enqueueSyncFn: func(accountID int) (int, error) { return 0, ErrAccountNotFound },
	})

	response := performRequest(router, http.MethodPost, "/email/accounts/9/sync", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestSyncAccountHandlerUnavailable(t *testing.T) {
	router := newTestRouter(&mockService{
		enqueueSyncFn: func(accountID int) (int, error) { return 0, ErrSyncUnavailable },
	})

	response := performRequest(router, http.MethodPost, "/email/accounts/1/sync", "")
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", response.Code)
	}
}

func TestDisabledHandlerAnswers503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/email/accounts", DisabledHandler)

	response := performRequest(router, http.MethodGet, "/email/accounts", "")
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "EMAIL_FEATURE_DISABLED_NO_KEY") {
		t.Fatalf("expected i18n key fallback in body, got %s", response.Body.String())
	}
}

func TestGetMessageSummaryHandler(t *testing.T) {
	router := newTestRouter(&mockService{
		analysisFn: func(messageID int) (AnalysisDto, error) {
			return AnalysisDto{MessageID: messageID, Verdict: "legitimate", Summary: "ok"}, nil
		},
	})
	w := performRequest(router, http.MethodGet, "/email/messages/7/summary", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "legitimate") {
		t.Fatalf("body missing verdict: %s", w.Body.String())
	}
}

func TestGetMessageSummaryHandlerNotFound(t *testing.T) {
	router := newTestRouter(&mockService{
		analysisFn: func(int) (AnalysisDto, error) { return AnalysisDto{}, ErrAnalysisNotFound },
	})
	w := performRequest(router, http.MethodGet, "/email/messages/7/summary", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestGetProviderHandler(t *testing.T) {
	router := newTestRouter(&mockService{
		getProviderFn: func() (ProviderPreferenceDto, error) {
			return ProviderPreferenceDto{Provider: ProviderPrefOllama}, nil
		},
	})
	w := performRequest(router, http.MethodGet, "/email/settings/provider", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "ollama") {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestSetProviderHandler(t *testing.T) {
	router := newTestRouter(&mockService{
		setProviderFn: func(provider string) (ProviderPreferenceDto, error) {
			return ProviderPreferenceDto{Provider: provider}, nil
		},
	})
	w := performRequest(router, http.MethodPut, "/email/settings/provider", `{"provider":"anthropic"}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "anthropic") {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestSetProviderHandlerInvalid(t *testing.T) {
	router := newTestRouter(&mockService{
		setProviderFn: func(string) (ProviderPreferenceDto, error) { return ProviderPreferenceDto{}, ErrInvalidProvider },
	})
	w := performRequest(router, http.MethodPut, "/email/settings/provider", `{"provider":"gemini"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
