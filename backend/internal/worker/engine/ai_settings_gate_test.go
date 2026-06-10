package engine

import (
	"context"
	"errors"
	"testing"

	"nas-go/api/pkg/ai"
)

type stubAIService struct{}

func (stubAIService) Execute(context.Context, ai.Request) (ai.Response, error) {
	return ai.Response{}, nil
}

type stubAISettings struct {
	enabled bool
	err     error
}

func (s stubAISettings) IsAIImageClassificationEnabled() (bool, error) {
	return s.enabled, s.err
}

func TestAIServiceForImageClassification(t *testing.T) {
	aiService := stubAIService{}

	t.Run("nil context returns nil", func(t *testing.T) {
		if aiServiceForImageClassification(nil) != nil {
			t.Fatalf("expected nil for nil context")
		}
	})

	t.Run("no AI service returns nil", func(t *testing.T) {
		ctx := &WorkerContext{AISettings: stubAISettings{enabled: true}}
		if aiServiceForImageClassification(ctx) != nil {
			t.Fatalf("expected nil when no AI service is wired")
		}
	})

	t.Run("no settings reader keeps AI enabled", func(t *testing.T) {
		ctx := &WorkerContext{AIService: aiService}
		if aiServiceForImageClassification(ctx) == nil {
			t.Fatalf("expected AI service when no settings reader is wired")
		}
	})

	t.Run("enabled returns AI service", func(t *testing.T) {
		ctx := &WorkerContext{AIService: aiService, AISettings: stubAISettings{enabled: true}}
		if aiServiceForImageClassification(ctx) == nil {
			t.Fatalf("expected AI service when toggle enabled")
		}
	})

	t.Run("disabled returns nil", func(t *testing.T) {
		ctx := &WorkerContext{AIService: aiService, AISettings: stubAISettings{enabled: false}}
		if aiServiceForImageClassification(ctx) != nil {
			t.Fatalf("expected nil when toggle disabled")
		}
	})

	t.Run("read error fails open", func(t *testing.T) {
		ctx := &WorkerContext{AIService: aiService, AISettings: stubAISettings{enabled: false, err: errors.New("boom")}}
		if aiServiceForImageClassification(ctx) == nil {
			t.Fatalf("expected AI service (fail open) on read error")
		}
	})
}
