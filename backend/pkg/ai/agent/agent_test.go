package agent

import (
	"context"
	"encoding/json"
	"errors"
	"nas-go/api/pkg/ai"
	"strings"
	"testing"
)

// scriptedAI returns queued Execute responses and optionally supports streaming.
type scriptedAI struct {
	execResponses []ai.Response
	execErr       error
	execCalls     int

	streaming    bool
	streamDeltas []string
	streamResp   ai.Response
	streamErr    error
}

func (f *scriptedAI) Execute(ctx context.Context, req ai.Request) (ai.Response, error) {
	if f.execErr != nil {
		return ai.Response{}, f.execErr
	}
	resp := f.execResponses[f.execCalls]
	f.execCalls++
	return resp, nil
}

type streamingScriptedAI struct {
	scriptedAI
}

func (f *streamingScriptedAI) ExecuteStream(ctx context.Context, req ai.Request, onChunk ai.StreamFunc) (ai.Response, error) {
	for _, d := range f.streamDeltas {
		if err := onChunk(d); err != nil {
			return ai.Response{}, err
		}
	}
	return f.streamResp, f.streamErr
}

func searchTool(invoked *bool) Tool {
	return Tool{
		Name:        "buscar_arquivos",
		Description: "busca arquivos",
		Parameters:  json.RawMessage(`{"type":"object"}`),
		Keywords:    []string{"arquivo", "busca", "procur"},
		Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
			*invoked = true
			return "encontrei 2 arquivos", nil
		},
	}
}

func TestRegistrySelect(t *testing.T) {
	registry := NewRegistry()
	registry.Register(searchTool(new(bool)))

	if got := registry.Select("preciso procurar um arquivo"); len(got) != 1 {
		t.Fatalf("expected 1 tool selected, got %d", len(got))
	}
	if got := registry.Select("bom dia, tudo bem?"); len(got) != 0 {
		t.Fatalf("expected no tools, got %d", len(got))
	}
}

func TestHasToolsFor(t *testing.T) {
	registry := NewRegistry()
	registry.Register(searchTool(new(bool)))
	agent := NewAgent(&scriptedAI{}, registry)

	if !agent.HasToolsFor("busca foto") {
		t.Fatal("expected tools for a matching message")
	}
	if agent.HasToolsFor("oi") {
		t.Fatal("expected no tools for chitchat")
	}

	if NewAgent(nil, registry).HasToolsFor("busca") {
		t.Fatal("nil ai should report no tools")
	}
	if NewAgent(&scriptedAI{}, nil).HasToolsFor("busca") {
		t.Fatal("nil registry should report no tools")
	}
}

func TestGenerateDirectAnswer(t *testing.T) {
	registry := NewRegistry()
	invoked := false
	registry.Register(searchTool(&invoked))
	fake := &scriptedAI{execResponses: []ai.Response{{Content: "resposta direta"}}}
	agent := NewAgent(fake, registry)

	var collected []string
	resp, err := agent.Generate(context.Background(), "sys", "Assistente:", "busca arquivo", func(c string) error {
		collected = append(collected, c)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "resposta direta" {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
	if invoked {
		t.Fatal("tool should not run when the model answers directly")
	}
	if strings.Join(collected, "") != "resposta direta" {
		t.Fatalf("expected direct answer streamed once, got %v", collected)
	}
}

func TestGenerateRunsToolThenStreamsFinal(t *testing.T) {
	registry := NewRegistry()
	invoked := false
	registry.Register(searchTool(&invoked))

	fake := &streamingScriptedAI{scriptedAI: scriptedAI{
		execResponses: []ai.Response{{ToolCalls: []ai.ToolCall{{Name: "buscar_arquivos", Arguments: json.RawMessage(`{"query":"ipva"}`)}}}},
		streamDeltas:  []string{"achei ", "2 arquivos"},
		streamResp:    ai.Response{Content: "achei 2 arquivos"},
	}}
	agent := NewAgent(fake, registry)

	var collected []string
	resp, err := agent.Generate(context.Background(), "sys", "Usuário: procura ipva\nAssistente:", "procura arquivo ipva", func(c string) error {
		collected = append(collected, c)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !invoked {
		t.Fatal("expected the tool to be executed")
	}
	if strings.Join(collected, "") != "achei 2 arquivos" || resp.Content != "achei 2 arquivos" {
		t.Fatalf("unexpected final answer: %v / %q", collected, resp.Content)
	}
}

func TestGenerateUnknownToolCall(t *testing.T) {
	registry := NewRegistry()
	registry.Register(searchTool(new(bool)))
	fake := &scriptedAI{execResponses: []ai.Response{
		{ToolCalls: []ai.ToolCall{{Name: "inexistente", Arguments: json.RawMessage(`{}`)}}},
		{Content: "respondo mesmo assim"},
	}}
	agent := NewAgent(fake, registry)

	resp, err := agent.Generate(context.Background(), "sys", "Assistente:", "busca arquivo", func(string) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "respondo mesmo assim" {
		t.Fatalf("expected final answer despite unknown tool, got %q", resp.Content)
	}
}

func TestGenerateToolHandlerError(t *testing.T) {
	registry := NewRegistry()
	registry.Register(Tool{
		Name:     "buscar_arquivos",
		Keywords: []string{"busca"},
		Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
			return "", errors.New("disco cheio")
		},
	})
	fake := &scriptedAI{execResponses: []ai.Response{
		{ToolCalls: []ai.ToolCall{{Name: "buscar_arquivos", Arguments: json.RawMessage(`{}`)}}},
		{Content: "deu erro na busca"},
	}}
	agent := NewAgent(fake, registry)

	resp, err := agent.Generate(context.Background(), "sys", "Assistente:", "busca", func(string) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "deu erro na busca" {
		t.Fatalf("expected final answer after tool error, got %q", resp.Content)
	}
}

func TestGenerateExecuteError(t *testing.T) {
	registry := NewRegistry()
	registry.Register(searchTool(new(bool)))
	fake := &scriptedAI{execErr: errors.New("ai down")}
	agent := NewAgent(fake, registry)

	_, err := agent.Generate(context.Background(), "sys", "Assistente:", "busca", func(string) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "ai down") {
		t.Fatalf("expected execute error, got %v", err)
	}
}
