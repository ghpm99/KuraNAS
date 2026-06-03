package ollama

import "context"

type ServiceInterface interface {
	GetStatus(ctx context.Context) StatusDto
	ListModels(ctx context.Context) ([]ModelDto, error)
	DeleteModel(ctx context.Context, name string) error
	PullModel(name string) (int, error)
}
