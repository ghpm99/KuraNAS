package ollama

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type serviceMock struct {
	statusFn func(ctx context.Context) StatusDto
	listFn   func(ctx context.Context) ([]ModelDto, error)
	deleteFn func(ctx context.Context, name string) error
	pullFn   func(name string) (int, error)
}

func (m *serviceMock) GetStatus(ctx context.Context) StatusDto {
	if m.statusFn != nil {
		return m.statusFn(ctx)
	}
	return StatusDto{}
}
func (m *serviceMock) ListModels(ctx context.Context) ([]ModelDto, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}
func (m *serviceMock) DeleteModel(ctx context.Context, name string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, name)
	}
	return nil
}
func (m *serviceMock) PullModel(name string) (int, error) {
	if m.pullFn != nil {
		return m.pullFn(name)
	}
	return 0, nil
}

func newTestContext(method, url string, body *bytes.Buffer) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	if body == nil {
		body = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	return ctx, rec
}

func TestGetStatusHandler(t *testing.T) {
	handler := NewHandler(&serviceMock{
		statusFn: func(ctx context.Context) StatusDto {
			return StatusDto{Reachable: true, Version: "0.5.0", Models: []ModelDto{}}
		},
	})
	ctx, rec := newTestContext(http.MethodGet, "/ai/ollama/status", nil)

	handler.GetStatusHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListModelsHandlerError(t *testing.T) {
	handler := NewHandler(&serviceMock{
		listFn: func(ctx context.Context) ([]ModelDto, error) {
			return nil, ErrModelNotFound
		},
	})
	ctx, rec := newTestContext(http.MethodGet, "/ai/ollama/models", nil)

	handler.ListModelsHandler(ctx)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 when daemon unreachable, got %d", rec.Code)
	}
}

func TestPullModelHandlerSuccess(t *testing.T) {
	handler := NewHandler(&serviceMock{
		pullFn: func(name string) (int, error) {
			if name != "llama3.1" {
				t.Fatalf("expected model llama3.1, got %s", name)
			}
			return 42, nil
		},
	})
	body := bytes.NewBufferString(`{"model":"llama3.1"}`)
	ctx, rec := newTestContext(http.MethodPost, "/ai/ollama/models/pull", body)

	handler.PullModelHandler(ctx)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}

func TestPullModelHandlerBadRequest(t *testing.T) {
	handler := NewHandler(&serviceMock{})
	ctx, rec := newTestContext(http.MethodPost, "/ai/ollama/models/pull", bytes.NewBufferString(`{}`))

	handler.PullModelHandler(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing model, got %d", rec.Code)
	}
}

func TestPullModelHandlerJobsUnavailable(t *testing.T) {
	handler := NewHandler(&serviceMock{
		pullFn: func(name string) (int, error) {
			return 0, ErrJobsUnavailable
		},
	})
	body := bytes.NewBufferString(`{"model":"llama3.1"}`)
	ctx, rec := newTestContext(http.MethodPost, "/ai/ollama/models/pull", body)

	handler.PullModelHandler(ctx)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestDeleteModelHandlerNotFound(t *testing.T) {
	handler := NewHandler(&serviceMock{
		deleteFn: func(ctx context.Context, name string) error {
			return ErrModelNotFound
		},
	})
	ctx, rec := newTestContext(http.MethodDelete, "/ai/ollama/models/missing", nil)
	ctx.Params = gin.Params{{Key: "name", Value: "missing"}}

	handler.DeleteModelHandler(ctx)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteModelHandlerSuccess(t *testing.T) {
	handler := NewHandler(&serviceMock{})
	ctx, rec := newTestContext(http.MethodDelete, "/ai/ollama/models/llama3.1", nil)
	ctx.Params = gin.Params{{Key: "name", Value: "llama3.1"}}

	handler.DeleteModelHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
