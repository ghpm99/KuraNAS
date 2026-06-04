package assistant

import (
	"context"
	"errors"
	"nas-go/api/pkg/ai"
	"strings"
	"testing"
	"time"
)

// --- AI fakes ---

type fakeAIService struct {
	resp   ai.Response
	err    error
	gotReq ai.Request
}

func (f *fakeAIService) Execute(ctx context.Context, req ai.Request) (ai.Response, error) {
	f.gotReq = req
	return f.resp, f.err
}

type streamingFakeAIService struct {
	deltas []string
	resp   ai.Response
	err    error
}

func (f *streamingFakeAIService) Execute(ctx context.Context, req ai.Request) (ai.Response, error) {
	return f.resp, f.err
}

func (f *streamingFakeAIService) ExecuteStream(ctx context.Context, req ai.Request, onChunk ai.StreamFunc) (ai.Response, error) {
	for _, delta := range f.deltas {
		if err := onChunk(delta); err != nil {
			return ai.Response{}, err
		}
	}
	return f.resp, f.err
}

type fakeAgent struct {
	hasTools   bool
	resp       ai.Response
	err        error
	deltas     []string
	gotMessage string
}

func (a *fakeAgent) HasToolsFor(message string) bool {
	a.gotMessage = message
	return a.hasTools
}

func (a *fakeAgent) Generate(ctx context.Context, systemPrompt, prompt, message string, onDelta ai.StreamFunc) (ai.Response, error) {
	for _, d := range a.deltas {
		if err := onDelta(d); err != nil {
			return ai.Response{}, err
		}
	}
	return a.resp, a.err
}

// --- in-memory repository ---

type fakeRepository struct {
	conversations map[int]ConversationModel
	messages      map[int][]MessageModel
	nextConvID    int
	nextMsgID     int

	createErr   error
	existsErr   error
	listConvErr error
	listMsgErr  error
	addErr      error
	touchErr    error
	deleteErr   error

	touched []int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		conversations: map[int]ConversationModel{},
		messages:      map[int][]MessageModel{},
	}
}

func (r *fakeRepository) CreateConversation(title string) (ConversationModel, error) {
	if r.createErr != nil {
		return ConversationModel{}, r.createErr
	}
	r.nextConvID++
	model := ConversationModel{ID: r.nextConvID, Title: title, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	r.conversations[model.ID] = model
	return model, nil
}

func (r *fakeRepository) ConversationExists(id int) (bool, error) {
	if r.existsErr != nil {
		return false, r.existsErr
	}
	_, ok := r.conversations[id]
	return ok, nil
}

func (r *fakeRepository) ListConversations() ([]ConversationModel, error) {
	if r.listConvErr != nil {
		return nil, r.listConvErr
	}
	out := []ConversationModel{}
	for _, c := range r.conversations {
		out = append(out, c)
	}
	return out, nil
}

func (r *fakeRepository) TouchConversation(id int) error {
	if r.touchErr != nil {
		return r.touchErr
	}
	r.touched = append(r.touched, id)
	return nil
}

func (r *fakeRepository) DeleteConversation(id int) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.conversations, id)
	delete(r.messages, id)
	return nil
}

func (r *fakeRepository) AddMessage(conversationID int, role, content string) (MessageModel, error) {
	if r.addErr != nil {
		return MessageModel{}, r.addErr
	}
	r.nextMsgID++
	model := MessageModel{ID: r.nextMsgID, ConversationID: conversationID, Role: role, Content: content, CreatedAt: time.Now()}
	r.messages[conversationID] = append(r.messages[conversationID], model)
	return model, nil
}

func (r *fakeRepository) ListMessages(conversationID int) ([]MessageModel, error) {
	if r.listMsgErr != nil {
		return nil, r.listMsgErr
	}
	return r.messages[conversationID], nil
}

func (r *fakeRepository) seedConversation(id int) {
	r.conversations[id] = ConversationModel{ID: id, Title: "seed"}
	if id > r.nextConvID {
		r.nextConvID = id
	}
}

// --- Chat tests ---

func TestChatNilAIService(t *testing.T) {
	service := NewService(nil, newFakeRepository(), nil)

	_, err := service.Chat(ChatRequestDto{Message: "oi"})

	if !errors.Is(err, ErrAIUnavailable) {
		t.Fatalf("expected ErrAIUnavailable, got %v", err)
	}
}

func TestChatBlankMessage(t *testing.T) {
	service := NewService(&fakeAIService{}, newFakeRepository(), nil)

	_, err := service.Chat(ChatRequestDto{Message: "   "})

	if !errors.Is(err, ErrInvalidConversation) {
		t.Fatalf("expected ErrInvalidConversation, got %v", err)
	}
}

func TestChatCreatesConversationAndPersists(t *testing.T) {
	repo := newFakeRepository()
	aiSvc := &fakeAIService{resp: ai.Response{Content: " Olá! ", Model: "llama3.1", Provider: "ollama"}}
	service := NewService(aiSvc, repo, nil)

	resp, err := service.Chat(ChatRequestDto{Message: "oi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ConversationID == 0 {
		t.Fatal("expected a conversation to be created")
	}
	if resp.Message.Content != "Olá!" || resp.Message.Role != RoleAssistant {
		t.Fatalf("unexpected reply: %+v", resp.Message)
	}
	stored := repo.messages[resp.ConversationID]
	if len(stored) != 2 || stored[0].Role != RoleUser || stored[1].Role != RoleAssistant {
		t.Fatalf("expected user+assistant persisted, got %+v", stored)
	}
	if len(repo.touched) != 1 {
		t.Fatalf("expected conversation touched once, got %v", repo.touched)
	}
	if !strings.HasSuffix(aiSvc.gotReq.Prompt, "Assistente:") {
		t.Fatalf("prompt should invite the assistant, got %q", aiSvc.gotReq.Prompt)
	}
}

func TestChatUsesExistingConversationHistory(t *testing.T) {
	repo := newFakeRepository()
	repo.seedConversation(7)
	repo.messages[7] = []MessageModel{
		{Role: RoleUser, Content: "pergunta antiga"},
		{Role: RoleAssistant, Content: "resposta antiga"},
	}
	aiSvc := &fakeAIService{resp: ai.Response{Content: "nova resposta"}}
	service := NewService(aiSvc, repo, nil)

	resp, err := service.Chat(ChatRequestDto{ConversationID: 7, Message: "nova pergunta"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ConversationID != 7 {
		t.Fatalf("expected conversation 7, got %d", resp.ConversationID)
	}
	if !strings.Contains(aiSvc.gotReq.Prompt, "resposta antiga") {
		t.Fatalf("prompt should include prior history, got %q", aiSvc.gotReq.Prompt)
	}
}

func TestChatConversationNotFound(t *testing.T) {
	service := NewService(&fakeAIService{}, newFakeRepository(), nil)

	_, err := service.Chat(ChatRequestDto{ConversationID: 999, Message: "oi"})

	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected ErrConversationNotFound, got %v", err)
	}
}

func TestChatAIError(t *testing.T) {
	service := NewService(&fakeAIService{err: errors.New("boom")}, newFakeRepository(), nil)

	_, err := service.Chat(ChatRequestDto{Message: "oi"})

	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected propagated AI error, got %v", err)
	}
}

func TestChatEmptyAIResponse(t *testing.T) {
	service := NewService(&fakeAIService{resp: ai.Response{Content: "   "}}, newFakeRepository(), nil)

	_, err := service.Chat(ChatRequestDto{Message: "oi"})

	if !errors.Is(err, ErrEmptyResponse) {
		t.Fatalf("expected ErrEmptyResponse, got %v", err)
	}
}

func TestChatCreateConversationError(t *testing.T) {
	repo := newFakeRepository()
	repo.createErr = errors.New("db down")
	service := NewService(&fakeAIService{}, repo, nil)

	_, err := service.Chat(ChatRequestDto{Message: "oi"})
	if err == nil || !strings.Contains(err.Error(), "db down") {
		t.Fatalf("expected create error, got %v", err)
	}
}

func TestChatAddMessageError(t *testing.T) {
	repo := newFakeRepository()
	repo.addErr = errors.New("insert failed")
	service := NewService(&fakeAIService{resp: ai.Response{Content: "ok"}}, repo, nil)

	_, err := service.Chat(ChatRequestDto{Message: "oi"})
	if err == nil || !strings.Contains(err.Error(), "insert failed") {
		t.Fatalf("expected add message error, got %v", err)
	}
}

// --- ChatStream tests ---

func TestChatStreamForwardsDeltasAndPersists(t *testing.T) {
	repo := newFakeRepository()
	fake := &streamingFakeAIService{deltas: []string{"Olá", ", mundo"}, resp: ai.Response{Content: "Olá, mundo", Model: "m", Provider: "p"}}
	service := NewService(fake, repo, nil)

	var collected []string
	resp, err := service.ChatStream(ChatRequestDto{Message: "oi"}, func(d string) error {
		collected = append(collected, d)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Join(collected, "") != "Olá, mundo" {
		t.Fatalf("expected reassembled deltas, got %v", collected)
	}
	if resp.Message.Content != "Olá, mundo" {
		t.Fatalf("unexpected final response: %+v", resp)
	}
	if len(repo.messages[resp.ConversationID]) != 2 {
		t.Fatalf("expected both turns persisted")
	}
}

func TestChatStreamFallsBackToNonStreaming(t *testing.T) {
	service := NewService(&fakeAIService{resp: ai.Response{Content: "resposta inteira"}}, newFakeRepository(), nil)

	var collected []string
	_, err := service.ChatStream(ChatRequestDto{Message: "oi"}, func(d string) error {
		collected = append(collected, d)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(collected) != 1 || collected[0] != "resposta inteira" {
		t.Fatalf("expected a single full-content delta, got %v", collected)
	}
}

func TestChatStreamConversationNotFound(t *testing.T) {
	service := NewService(&streamingFakeAIService{}, newFakeRepository(), nil)

	_, err := service.ChatStream(ChatRequestDto{ConversationID: 5, Message: "oi"}, func(string) error { return nil })
	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected ErrConversationNotFound, got %v", err)
	}
}

// --- Conversation/message queries ---

func TestListConversations(t *testing.T) {
	repo := newFakeRepository()
	repo.seedConversation(1)
	service := NewService(&fakeAIService{}, repo, nil)

	conversations, err := service.ListConversations()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conversations) != 1 || conversations[0].ID != 1 {
		t.Fatalf("unexpected conversations: %+v", conversations)
	}
}

func TestListConversationsError(t *testing.T) {
	repo := newFakeRepository()
	repo.listConvErr = errors.New("boom")
	service := NewService(&fakeAIService{}, repo, nil)

	if _, err := service.ListConversations(); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetMessages(t *testing.T) {
	repo := newFakeRepository()
	repo.seedConversation(3)
	repo.messages[3] = []MessageModel{{ID: 1, Role: RoleUser, Content: "oi", CreatedAt: time.Now()}}
	service := NewService(&fakeAIService{}, repo, nil)

	messages, err := service.GetMessages(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 1 || messages[0].Content != "oi" || messages[0].CreatedAt == "" {
		t.Fatalf("unexpected messages: %+v", messages)
	}
}

func TestGetMessagesNotFound(t *testing.T) {
	service := NewService(&fakeAIService{}, newFakeRepository(), nil)

	if _, err := service.GetMessages(404); !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected ErrConversationNotFound, got %v", err)
	}
}

func TestDeleteConversation(t *testing.T) {
	repo := newFakeRepository()
	repo.seedConversation(9)
	service := NewService(&fakeAIService{}, repo, nil)

	if err := service.DeleteConversation(9); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := repo.conversations[9]; ok {
		t.Fatal("expected conversation removed")
	}
}

func TestDeleteConversationNotFound(t *testing.T) {
	service := NewService(&fakeAIService{}, newFakeRepository(), nil)

	if err := service.DeleteConversation(123); !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("expected ErrConversationNotFound, got %v", err)
	}
}

func TestChatUsesAgentWhenToolsMatch(t *testing.T) {
	repo := newFakeRepository()
	agent := &fakeAgent{hasTools: true, resp: ai.Response{Content: "achei 3 arquivos", Model: "m", Provider: "p"}}
	service := NewService(&fakeAIService{}, repo, agent)

	resp, err := service.Chat(ChatRequestDto{Message: "procura o pdf do ipva"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Message.Content != "achei 3 arquivos" {
		t.Fatalf("expected the agent reply, got %q", resp.Message.Content)
	}
	if agent.gotMessage != "procura o pdf do ipva" {
		t.Fatalf("agent should receive the user message, got %q", agent.gotMessage)
	}
	if len(repo.messages[resp.ConversationID]) != 2 {
		t.Fatalf("expected both turns persisted")
	}
}

func TestChatStreamUsesAgentDeltas(t *testing.T) {
	repo := newFakeRepository()
	agent := &fakeAgent{hasTools: true, deltas: []string{"achei ", "tudo"}, resp: ai.Response{Content: "achei tudo"}}
	service := NewService(&fakeAIService{}, repo, agent)

	var collected []string
	resp, err := service.ChatStream(ChatRequestDto{Message: "busca foto"}, func(d string) error {
		collected = append(collected, d)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Join(collected, "") != "achei tudo" || resp.Message.Content != "achei tudo" {
		t.Fatalf("unexpected agent stream: %v / %+v", collected, resp)
	}
}

func TestChatSkipsAgentWhenNoToolMatch(t *testing.T) {
	repo := newFakeRepository()
	agent := &fakeAgent{hasTools: false}
	service := NewService(&fakeAIService{resp: ai.Response{Content: "olá!"}}, repo, agent)

	resp, err := service.Chat(ChatRequestDto{Message: "bom dia"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Message.Content != "olá!" {
		t.Fatalf("expected the plain AI reply, got %q", resp.Message.Content)
	}
}

func TestMakeTitleTruncates(t *testing.T) {
	long := strings.Repeat("a", 100)
	title := makeTitle(long + "\nsegunda linha")
	if len([]rune(title)) > maxTitleLength+1 {
		t.Fatalf("title not truncated: %q", title)
	}
	if strings.Contains(title, "segunda linha") {
		t.Fatalf("title should use only the first line, got %q", title)
	}
}

func TestCapHistory(t *testing.T) {
	messages := make([]ChatMessageDto, maxHistoryMessages+5)
	for i := range messages {
		messages[i] = ChatMessageDto{Role: RoleUser, Content: "m"}
	}
	if got := capHistory(messages); len(got) != maxHistoryMessages {
		t.Fatalf("expected cap to %d, got %d", maxHistoryMessages, len(got))
	}
}
