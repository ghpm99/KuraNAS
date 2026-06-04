package assistant

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	resp ChatResponseDto
	err  error
}

func (f *fakeService) Chat(messages []ChatMessageDto) (ChatResponseDto, error) {
	return f.resp, f.err
}

func setupRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/assistant/chat", handler.ChatHandler)
	return router
}

func performRequest(router *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/assistant/chat", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestChatHandlerNilService(t *testing.T) {
	router := setupRouter(NewHandler(nil))

	rec := performRequest(router, `{"messages":[{"role":"user","content":"oi"}]}`)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestChatHandlerBadJSON(t *testing.T) {
	router := setupRouter(NewHandler(&fakeService{}))

	rec := performRequest(router, `{not json`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestChatHandlerInvalidConversation(t *testing.T) {
	router := setupRouter(NewHandler(&fakeService{err: ErrInvalidConversation}))

	rec := performRequest(router, `{"messages":[]}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestChatHandlerAIUnavailable(t *testing.T) {
	router := setupRouter(NewHandler(&fakeService{err: ErrAIUnavailable}))

	rec := performRequest(router, `{"messages":[{"role":"user","content":"oi"}]}`)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestChatHandlerInternalError(t *testing.T) {
	router := setupRouter(NewHandler(&fakeService{err: ErrEmptyResponse}))

	rec := performRequest(router, `{"messages":[{"role":"user","content":"oi"}]}`)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestChatHandlerSuccess(t *testing.T) {
	expected := ChatResponseDto{
		Message:  ChatMessageDto{Role: RoleAssistant, Content: "Olá!"},
		Model:    "llama3.1",
		Provider: "ollama",
	}
	router := setupRouter(NewHandler(&fakeService{resp: expected}))

	rec := performRequest(router, `{"messages":[{"role":"user","content":"oi"}]}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var got ChatResponseDto
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if got.Message.Content != "Olá!" || got.Message.Role != RoleAssistant {
		t.Fatalf("unexpected message: %+v", got.Message)
	}
	if got.Model != "llama3.1" || got.Provider != "ollama" {
		t.Fatalf("missing traceability: %+v", got)
	}
}
