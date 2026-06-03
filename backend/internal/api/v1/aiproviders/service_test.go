package aiproviders

import (
	"database/sql"
	"testing"

	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/database"
)

type fakeRepository struct {
	providers map[ProviderName]ProviderModel
	inserted  []ProviderName
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{providers: map[ProviderName]ProviderModel{}}
}

func (r *fakeRepository) GetDbContext() *database.DbContext { return nil }

func (r *fakeRepository) GetAll() ([]ProviderModel, error) {
	out := make([]ProviderModel, 0, len(r.providers))
	for _, m := range r.providers {
		out = append(out, m)
	}
	return out, nil
}

func (r *fakeRepository) GetByName(name ProviderName) (ProviderModel, error) {
	m, ok := r.providers[name]
	if !ok {
		return ProviderModel{}, sql.ErrNoRows
	}
	return m, nil
}

func (r *fakeRepository) InsertIfAbsent(model ProviderModel) error {
	r.inserted = append(r.inserted, model.Name)
	if _, ok := r.providers[model.Name]; !ok {
		r.providers[model.Name] = model
	}
	return nil
}

func (r *fakeRepository) Update(model ProviderModel) (ProviderModel, error) {
	r.providers[model.Name] = model
	return model, nil
}

func TestEnsureDefaultsSeedsAllProviders(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo, ai.Config{OllamaEnabled: true, OllamaModel: "llama3.1", OllamaBaseURL: "http://x:11434"})

	if err := svc.EnsureDefaults(); err != nil {
		t.Fatalf("EnsureDefaults error: %v", err)
	}
	if len(repo.inserted) != 3 {
		t.Fatalf("expected 3 seed attempts, got %d", len(repo.inserted))
	}

	ollama, _ := repo.GetByName(ProviderOllama)
	if !ollama.Enabled || ollama.Model != "llama3.1" {
		t.Fatalf("unexpected ollama default: %+v", ollama)
	}
}

func TestGetProvidersReportsAPIKeyConfigured(t *testing.T) {
	repo := newFakeRepository()
	repo.providers[ProviderOpenAI] = ProviderModel{Name: ProviderOpenAI, Model: "gpt-4o-mini"}
	repo.providers[ProviderOllama] = ProviderModel{Name: ProviderOllama, Model: "llama3.1"}

	svc := NewService(repo, ai.Config{OpenAIAPIKey: ""}) // no key

	providers, err := svc.GetProviders()
	if err != nil {
		t.Fatalf("GetProviders error: %v", err)
	}

	for _, p := range providers {
		switch ProviderName(p.Name) {
		case ProviderOpenAI:
			if !p.RequiresAPIKey || p.APIKeyConfigured {
				t.Fatalf("openai should require key and report none configured: %+v", p)
			}
		case ProviderOllama:
			if p.RequiresAPIKey || !p.APIKeyConfigured {
				t.Fatalf("ollama should not require key and be considered configured: %+v", p)
			}
		}
	}
}

func TestUpdateProviderTriggersOnChange(t *testing.T) {
	repo := newFakeRepository()
	repo.providers[ProviderOllama] = ProviderModel{Name: ProviderOllama, Enabled: false, Model: "llama3.1"}

	svc := NewService(repo, ai.Config{})
	changed := 0
	svc.SetOnChange(func() { changed++ })

	dto, err := svc.UpdateProvider(ProviderOllama, UpdateProviderDto{Enabled: true, Model: "qwen2.5", Priority: 0})
	if err != nil {
		t.Fatalf("UpdateProvider error: %v", err)
	}
	if !dto.Enabled || dto.Model != "qwen2.5" {
		t.Fatalf("unexpected updated dto: %+v", dto)
	}
	if changed != 1 {
		t.Fatalf("expected onChange called once, got %d", changed)
	}
}

func TestUpdateProviderInvalidName(t *testing.T) {
	svc := NewService(newFakeRepository(), ai.Config{})
	if _, err := svc.UpdateProvider(ProviderName("bogus"), UpdateProviderDto{}); err != ErrInvalidProvider {
		t.Fatalf("expected ErrInvalidProvider, got %v", err)
	}
}

func TestUpdateProviderNotFound(t *testing.T) {
	svc := NewService(newFakeRepository(), ai.Config{})
	if _, err := svc.UpdateProvider(ProviderOpenAI, UpdateProviderDto{}); err != ErrProviderNotFound {
		t.Fatalf("expected ErrProviderNotFound, got %v", err)
	}
}
