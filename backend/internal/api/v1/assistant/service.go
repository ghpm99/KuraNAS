package assistant

import (
	"context"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/ai/prompts"
	"strings"
	"time"
)

const (
	// maxHistoryMessages caps how many turns are sent to the model, keeping the
	// prompt (and therefore the local-model cost) bounded on long conversations.
	maxHistoryMessages = 20
	// maxResponseTokens bounds the reply length.
	maxResponseTokens = 800
	// chatTemperature favours natural conversation over deterministic output.
	chatTemperature = 0.7
	// chatTimeout is the budget for a single reply from a local model.
	chatTimeout = 60 * time.Second
	// maxTitleLength caps the auto-generated conversation title.
	maxTitleLength = 60
)

type Service struct {
	AIService  ai.ServiceInterface
	Repository RepositoryInterface
	Agent      AgentInterface
}

func NewService(aiService ai.ServiceInterface, repository RepositoryInterface, agent AgentInterface) ServiceInterface {
	return &Service{AIService: aiService, Repository: repository, Agent: agent}
}

// generationInput is the resolved material needed to produce a reply.
type generationInput struct {
	systemPrompt string
	prompt       string
	message      string
}

func (s *Service) Chat(input ChatRequestDto) (ChatResponseDto, error) {
	conversationID, gen, err := s.prepare(input)
	if err != nil {
		return ChatResponseDto{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), chatTimeout)
	defer cancel()

	resp, err := s.generate(ctx, gen, func(string) error { return nil })
	if err != nil {
		return ChatResponseDto{}, err
	}

	return s.finalize(conversationID, resp)
}

func (s *Service) ChatStream(input ChatRequestDto, onDelta DeltaFunc) (ChatResponseDto, error) {
	conversationID, gen, err := s.prepare(input)
	if err != nil {
		return ChatResponseDto{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), chatTimeout)
	defer cancel()

	resp, err := s.generate(ctx, gen, func(chunk string) error { return onDelta(chunk) })
	if err != nil {
		return ChatResponseDto{}, err
	}

	return s.finalize(conversationID, resp)
}

// generate produces the reply, routing through the tool-calling agent when the
// message matches a tool and otherwise streaming directly from the AI service.
// onDelta receives the streamed answer; callers that do not stream pass a no-op.
func (s *Service) generate(ctx context.Context, gen generationInput, onDelta ai.StreamFunc) (ai.Response, error) {
	if s.Agent != nil && s.Agent.HasToolsFor(gen.message) {
		return s.Agent.Generate(ctx, gen.systemPrompt, gen.prompt, gen.message, onDelta)
	}

	req := ai.Request{
		TaskType:     ai.TaskGeneration,
		SystemPrompt: gen.systemPrompt,
		Prompt:       gen.prompt,
		MaxTokens:    maxResponseTokens,
		Temperature:  chatTemperature,
	}

	if streamer, ok := s.AIService.(ai.StreamingServiceInterface); ok {
		return streamer.ExecuteStream(ctx, req, onDelta)
	}

	resp, err := s.AIService.Execute(ctx, req)
	if err != nil {
		return ai.Response{}, err
	}
	if resp.Content != "" {
		if cbErr := onDelta(resp.Content); cbErr != nil {
			return ai.Response{}, cbErr
		}
	}
	return resp, nil
}

func (s *Service) ListConversations() ([]ConversationDto, error) {
	models, err := s.Repository.ListConversations()
	if err != nil {
		return nil, err
	}
	dtos := make([]ConversationDto, 0, len(models))
	for _, model := range models {
		dtos = append(dtos, toConversationDto(model))
	}
	return dtos, nil
}

func (s *Service) GetMessages(conversationID int) ([]MessageDto, error) {
	exists, err := s.Repository.ConversationExists(conversationID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrConversationNotFound
	}

	models, err := s.Repository.ListMessages(conversationID)
	if err != nil {
		return nil, err
	}
	dtos := make([]MessageDto, 0, len(models))
	for _, model := range models {
		dtos = append(dtos, toMessageDto(model))
	}
	return dtos, nil
}

func (s *Service) DeleteConversation(conversationID int) error {
	exists, err := s.Repository.ConversationExists(conversationID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrConversationNotFound
	}
	return s.Repository.DeleteConversation(conversationID)
}

// prepare validates the input, resolves (or creates) the conversation, persists
// the user message, and builds the AI request from the stored history plus the
// new turn. It is shared by the streaming and non-streaming paths.
func (s *Service) prepare(input ChatRequestDto) (int, generationInput, error) {
	if s.AIService == nil {
		return 0, generationInput{}, ErrAIUnavailable
	}

	message := strings.TrimSpace(input.Message)
	if message == "" {
		return 0, generationInput{}, ErrInvalidConversation
	}

	conversationID, err := s.resolveConversation(input.ConversationID, message)
	if err != nil {
		return 0, generationInput{}, err
	}

	prior, err := s.Repository.ListMessages(conversationID)
	if err != nil {
		return 0, generationInput{}, err
	}

	history := capHistory(append(toChatMessages(prior), ChatMessageDto{Role: RoleUser, Content: message}))

	if _, err := s.Repository.AddMessage(conversationID, RoleUser, message); err != nil {
		return 0, generationInput{}, err
	}

	gen := generationInput{
		systemPrompt: prompts.AssistantChatSystemPrompt(),
		prompt:       buildPrompt(history),
		message:      message,
	}
	return conversationID, gen, nil
}

func (s *Service) resolveConversation(conversationID int, firstMessage string) (int, error) {
	if conversationID > 0 {
		exists, err := s.Repository.ConversationExists(conversationID)
		if err != nil {
			return 0, err
		}
		if !exists {
			return 0, ErrConversationNotFound
		}
		return conversationID, nil
	}

	conversation, err := s.Repository.CreateConversation(makeTitle(firstMessage))
	if err != nil {
		return 0, err
	}
	return conversation.ID, nil
}

func (s *Service) finalize(conversationID int, resp ai.Response) (ChatResponseDto, error) {
	content := strings.TrimSpace(resp.Content)
	if content == "" {
		return ChatResponseDto{}, ErrEmptyResponse
	}

	if _, err := s.Repository.AddMessage(conversationID, RoleAssistant, content); err != nil {
		return ChatResponseDto{}, err
	}
	if err := s.Repository.TouchConversation(conversationID); err != nil {
		return ChatResponseDto{}, err
	}

	return ChatResponseDto{
		ConversationID: conversationID,
		Message:        ChatMessageDto{Role: RoleAssistant, Content: content},
		Model:          resp.Model,
		Provider:       resp.Provider,
	}, nil
}

// capHistory keeps only the most recent maxHistoryMessages so the prompt stays
// bounded on long conversations.
func capHistory(messages []ChatMessageDto) []ChatMessageDto {
	if len(messages) > maxHistoryMessages {
		return messages[len(messages)-maxHistoryMessages:]
	}
	return messages
}

// buildPrompt flattens the conversation into a transcript the model continues as
// the assistant. The single-string shape fits the current ai.Request contract;
// multi-turn tool calling comes in a later iteration.
func buildPrompt(messages []ChatMessageDto) string {
	var b strings.Builder
	for _, m := range messages {
		label := "Usuário"
		if m.Role == RoleAssistant {
			label = "Assistente"
		}
		b.WriteString(label)
		b.WriteString(": ")
		b.WriteString(m.Content)
		b.WriteString("\n")
	}
	b.WriteString("Assistente:")
	return b.String()
}

func makeTitle(message string) string {
	title := strings.TrimSpace(strings.SplitN(message, "\n", 2)[0])
	runes := []rune(title)
	if len(runes) > maxTitleLength {
		return strings.TrimSpace(string(runes[:maxTitleLength])) + "…"
	}
	return title
}

func toChatMessages(models []MessageModel) []ChatMessageDto {
	out := make([]ChatMessageDto, 0, len(models))
	for _, m := range models {
		out = append(out, ChatMessageDto{Role: m.Role, Content: m.Content})
	}
	return out
}

func toConversationDto(model ConversationModel) ConversationDto {
	return ConversationDto{
		ID:        model.ID,
		Title:     model.Title,
		CreatedAt: model.CreatedAt.Format(time.RFC3339),
		UpdatedAt: model.UpdatedAt.Format(time.RFC3339),
	}
}

func toMessageDto(model MessageModel) MessageDto {
	return MessageDto{
		ID:        model.ID,
		Role:      model.Role,
		Content:   model.Content,
		CreatedAt: model.CreatedAt.Format(time.RFC3339),
	}
}
