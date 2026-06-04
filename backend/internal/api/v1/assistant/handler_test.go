package assistant

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	resp          ChatResponseDto
	err           error
	deltas        []string
	conversations []ConversationDto
	messages      []MessageDto
	listErr       error
	messagesErr   error
	deleteErr     error
	deletedID     int
}

func (f *fakeService) Chat(input ChatRequestDto) (ChatResponseDto, error) {
	return f.resp, f.err
}

func (f *fakeService) ChatStream(input ChatRequestDto, onDelta DeltaFunc) (ChatResponseDto, error) {
	for _, delta := range f.deltas {
		if err := onDelta(delta); err != nil {
			return ChatResponseDto{}, err
		}
	}
	return f.resp, f.err
}

func (f *fakeService) ListConversations() ([]ConversationDto, error) {
	return f.conversations, f.listErr
}

func (f *fakeService) GetMessages(conversationID int) ([]MessageDto, error) {
	return f.messages, f.messagesErr
}

func (f *fakeService) DeleteConversation(conversationID int) error {
	f.deletedID = conversationID
	return f.deleteErr
}

func newRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/assistant/chat", handler.ChatHandler)
	router.POST("/assistant/chat/stream", handler.ChatStreamHandler)
	router.GET("/assistant/conversations", handler.ListConversationsHandler)
	router.GET("/assistant/conversations/:id/messages", handler.GetMessagesHandler)
	router.DELETE("/assistant/conversations/:id", handler.DeleteConversationHandler)
	return router
}

func do(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestChatHandlerNilService(t *testing.T) {
	rec := do(newRouter(NewHandler(nil)), http.MethodPost, "/assistant/chat", `{"message":"oi"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestChatHandlerBadJSON(t *testing.T) {
	rec := do(newRouter(NewHandler(&fakeService{})), http.MethodPost, "/assistant/chat", `{nope`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestChatHandlerInvalidConversation(t *testing.T) {
	rec := do(newRouter(NewHandler(&fakeService{err: ErrInvalidConversation})), http.MethodPost, "/assistant/chat", `{"message":""}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestChatHandlerNotFound(t *testing.T) {
	rec := do(newRouter(NewHandler(&fakeService{err: ErrConversationNotFound})), http.MethodPost, "/assistant/chat", `{"conversation_id":9,"message":"oi"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestChatHandlerAIUnavailable(t *testing.T) {
	rec := do(newRouter(NewHandler(&fakeService{err: ErrAIUnavailable})), http.MethodPost, "/assistant/chat", `{"message":"oi"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestChatHandlerInternalError(t *testing.T) {
	rec := do(newRouter(NewHandler(&fakeService{err: ErrEmptyResponse})), http.MethodPost, "/assistant/chat", `{"message":"oi"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestChatHandlerSuccess(t *testing.T) {
	expected := ChatResponseDto{
		ConversationID: 4,
		Message:        ChatMessageDto{Role: RoleAssistant, Content: "Olá!"},
		Model:          "llama3.1",
		Provider:       "ollama",
	}
	rec := do(newRouter(NewHandler(&fakeService{resp: expected})), http.MethodPost, "/assistant/chat", `{"message":"oi"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got ChatResponseDto
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.ConversationID != 4 || got.Message.Content != "Olá!" {
		t.Fatalf("unexpected body: %+v", got)
	}
}

func TestChatStreamHandlerNilService(t *testing.T) {
	rec := do(newRouter(NewHandler(nil)), http.MethodPost, "/assistant/chat/stream", `{"message":"oi"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestChatStreamHandlerBadJSON(t *testing.T) {
	rec := do(newRouter(NewHandler(&fakeService{})), http.MethodPost, "/assistant/chat/stream", `{nope`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestChatStreamHandlerStreamsDeltasThenDone(t *testing.T) {
	svc := &fakeService{
		deltas: []string{"Olá", ", mundo"},
		resp:   ChatResponseDto{ConversationID: 1, Message: ChatMessageDto{Role: RoleAssistant, Content: "Olá, mundo"}},
	}
	rec := do(newRouter(NewHandler(svc)), http.MethodPost, "/assistant/chat/stream", `{"message":"oi"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "event: delta") || !strings.Contains(body, `"content":"Olá"`) {
		t.Fatalf("expected delta events, got %q", body)
	}
	if !strings.Contains(body, "event: done") || !strings.Contains(body, "Olá, mundo") {
		t.Fatalf("expected a done event, got %q", body)
	}
}

func TestChatStreamHandlerEmitsErrorEvent(t *testing.T) {
	rec := do(newRouter(NewHandler(&fakeService{err: ErrEmptyResponse})), http.MethodPost, "/assistant/chat/stream", `{"message":"oi"}`)
	if !strings.Contains(rec.Body.String(), "event: error") {
		t.Fatalf("expected error event, got %q", rec.Body.String())
	}
}

func TestListConversationsHandler(t *testing.T) {
	svc := &fakeService{conversations: []ConversationDto{{ID: 1, Title: "Primeira"}}}
	rec := do(newRouter(NewHandler(svc)), http.MethodGet, "/assistant/conversations", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Primeira") {
		t.Fatalf("expected conversation in body, got %q", rec.Body.String())
	}
}

func TestListConversationsHandlerNilService(t *testing.T) {
	rec := do(newRouter(NewHandler(nil)), http.MethodGet, "/assistant/conversations", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestListConversationsHandlerError(t *testing.T) {
	svc := &fakeService{listErr: ErrEmptyResponse}
	rec := do(newRouter(NewHandler(svc)), http.MethodGet, "/assistant/conversations", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetMessagesHandlerSuccess(t *testing.T) {
	svc := &fakeService{messages: []MessageDto{{ID: 1, Role: RoleUser, Content: "oi"}}}
	rec := do(newRouter(NewHandler(svc)), http.MethodGet, "/assistant/conversations/3/messages", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGetMessagesHandlerBadID(t *testing.T) {
	rec := do(newRouter(NewHandler(&fakeService{})), http.MethodGet, "/assistant/conversations/abc/messages", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetMessagesHandlerNotFound(t *testing.T) {
	rec := do(newRouter(NewHandler(&fakeService{messagesErr: ErrConversationNotFound})), http.MethodGet, "/assistant/conversations/3/messages", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteConversationHandlerSuccess(t *testing.T) {
	svc := &fakeService{}
	rec := do(newRouter(NewHandler(svc)), http.MethodDelete, "/assistant/conversations/8", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if svc.deletedID != 8 {
		t.Fatalf("expected delete of 8, got %d", svc.deletedID)
	}
}

func TestDeleteConversationHandlerBadID(t *testing.T) {
	rec := do(newRouter(NewHandler(&fakeService{})), http.MethodDelete, "/assistant/conversations/0", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteConversationHandlerNotFound(t *testing.T) {
	rec := do(newRouter(NewHandler(&fakeService{deleteErr: ErrConversationNotFound})), http.MethodDelete, "/assistant/conversations/8", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
