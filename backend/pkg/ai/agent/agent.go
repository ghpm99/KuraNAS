// Package agent adds tool calling on top of the ai.Service: a keyword router
// decides whether a message needs tools, and a bounded loop lets the model call
// them and then answer from the results. Tools are generic; their handlers are
// supplied by the composition root so this package stays free of feature deps.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"nas-go/api/pkg/ai"
	"strings"
)

const (
	maxResponseTokens = 800
	toolTemperature   = 0.3
	finalTemperature  = 0.7
)

// ToolHandler executes a tool with the raw JSON arguments the model supplied and
// returns a short textual result for the model to read.
type ToolHandler func(ctx context.Context, args json.RawMessage) (string, error)

// Tool is a callable function exposed to the model. Keywords drive the keyword
// routing tier (matched case-insensitively as substrings of the user message).
type Tool struct {
	Name        string
	Description string
	Parameters  json.RawMessage
	Keywords    []string
	Handler     ToolHandler
}

// Registry holds the available tools and performs keyword routing.
type Registry struct {
	tools []Tool
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(tool Tool) {
	r.tools = append(r.tools, tool)
}

// Select returns the tools whose keywords match the message. Keeping the model's
// tool list short (only the relevant ones) keeps the local-model prompt small
// and improves tool-choice accuracy.
func (r *Registry) Select(message string) []Tool {
	lower := strings.ToLower(message)
	var selected []Tool
	for _, tool := range r.tools {
		for _, keyword := range tool.Keywords {
			if keyword != "" && strings.Contains(lower, strings.ToLower(keyword)) {
				selected = append(selected, tool)
				break
			}
		}
	}
	return selected
}

// Agent runs the tool-calling loop over an ai service.
type Agent struct {
	ai       ai.ServiceInterface
	registry *Registry
}

func NewAgent(aiService ai.ServiceInterface, registry *Registry) *Agent {
	return &Agent{ai: aiService, registry: registry}
}

// HasToolsFor reports whether the message routes to at least one tool.
func (a *Agent) HasToolsFor(message string) bool {
	if a.ai == nil || a.registry == nil {
		return false
	}
	return len(a.registry.Select(message)) > 0
}

// Generate runs the bounded tool loop: it offers the routed tools to the model;
// if the model answers directly the answer is returned (emitted as one chunk);
// if it requests tools, they are executed and a final, streamed answer is
// produced from the results. onDelta receives the streamed final answer.
func (a *Agent) Generate(ctx context.Context, systemPrompt, prompt, message string, onDelta ai.StreamFunc) (ai.Response, error) {
	tools := a.registry.Select(message)
	defs := toDefinitions(tools)

	resp, err := a.ai.Execute(ctx, ai.Request{
		TaskType:     ai.TaskGeneration,
		SystemPrompt: systemPrompt,
		Prompt:       prompt,
		Tools:        defs,
		MaxTokens:    maxResponseTokens,
		Temperature:  toolTemperature,
	})
	if err != nil {
		return ai.Response{}, err
	}

	if len(resp.ToolCalls) == 0 {
		if resp.Content != "" {
			if cbErr := onDelta(resp.Content); cbErr != nil {
				return ai.Response{}, cbErr
			}
		}
		return resp, nil
	}

	results := a.runTools(ctx, tools, resp.ToolCalls)
	finalPrompt := strings.TrimSuffix(prompt, "Assistente:") +
		"[Resultados das ferramentas]\n" + results +
		"\nResponda ao usuário em português, com base nos resultados acima.\nAssistente:"

	return a.executeFinal(ctx, ai.Request{
		TaskType:     ai.TaskGeneration,
		SystemPrompt: systemPrompt,
		Prompt:       finalPrompt,
		MaxTokens:    maxResponseTokens,
		Temperature:  finalTemperature,
	}, onDelta)
}

func (a *Agent) runTools(ctx context.Context, tools []Tool, calls []ai.ToolCall) string {
	byName := make(map[string]Tool, len(tools))
	for _, tool := range tools {
		byName[tool.Name] = tool
	}

	var b strings.Builder
	for _, call := range calls {
		tool, ok := byName[call.Name]
		if !ok {
			fmt.Fprintf(&b, "Ferramenta %q indisponível.\n", call.Name)
			continue
		}
		result, err := tool.Handler(ctx, call.Arguments)
		if err != nil {
			fmt.Fprintf(&b, "Ferramenta %s falhou: %v\n", call.Name, err)
			continue
		}
		fmt.Fprintf(&b, "Ferramenta %s: %s\n", call.Name, result)
	}
	return b.String()
}

// executeFinal streams the final answer when the service supports streaming, and
// otherwise falls back to a single chunk — mirroring the ai service contract.
func (a *Agent) executeFinal(ctx context.Context, req ai.Request, onDelta ai.StreamFunc) (ai.Response, error) {
	if streamer, ok := a.ai.(ai.StreamingServiceInterface); ok {
		return streamer.ExecuteStream(ctx, req, onDelta)
	}
	resp, err := a.ai.Execute(ctx, req)
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

func toDefinitions(tools []Tool) []ai.ToolDefinition {
	if len(tools) == 0 {
		return nil
	}
	defs := make([]ai.ToolDefinition, 0, len(tools))
	for _, tool := range tools {
		defs = append(defs, ai.ToolDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
		})
	}
	return defs
}
