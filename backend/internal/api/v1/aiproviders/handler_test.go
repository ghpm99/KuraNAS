package aiproviders

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type serviceMock struct {
	getFn    func() ([]ProviderDto, error)
	updateFn func(name ProviderName, dto UpdateProviderDto) (ProviderDto, error)
}

func (m *serviceMock) GetProviders() ([]ProviderDto, error) {
	if m.getFn != nil {
		return m.getFn()
	}
	return []ProviderDto{}, nil
}
func (m *serviceMock) UpdateProvider(name ProviderName, dto UpdateProviderDto) (ProviderDto, error) {
	if m.updateFn != nil {
		return m.updateFn(name, dto)
	}
	return ProviderDto{Name: string(name)}, nil
}
func (m *serviceMock) EnsureDefaults() error                       { return nil }
func (m *serviceMock) GetProviderModels() ([]ProviderModel, error) { return nil, nil }
func (m *serviceMock) SetOnChange(fn func())                       {}

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

func TestGetProvidersHandlerSuccess(t *testing.T) {
	handler := NewHandler(&serviceMock{
		getFn: func() ([]ProviderDto, error) {
			return []ProviderDto{{Name: "ollama", Enabled: true}}, nil
		},
	})
	ctx, rec := newTestContext(http.MethodGet, "/ai/providers", nil)

	handler.GetProvidersHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUpdateProviderHandlerInvalidName(t *testing.T) {
	handler := NewHandler(&serviceMock{})
	ctx, rec := newTestContext(http.MethodPut, "/ai/providers/bogus", bytes.NewBufferString(`{}`))
	ctx.Params = gin.Params{{Key: "name", Value: "bogus"}}

	handler.UpdateProviderHandler(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid provider name, got %d", rec.Code)
	}
}

func TestUpdateProviderHandlerNotFound(t *testing.T) {
	handler := NewHandler(&serviceMock{
		updateFn: func(name ProviderName, dto UpdateProviderDto) (ProviderDto, error) {
			return ProviderDto{}, ErrProviderNotFound
		},
	})
	body := bytes.NewBufferString(`{"enabled":true,"model":"x","base_url":"","priority":0,"params":{}}`)
	ctx, rec := newTestContext(http.MethodPut, "/ai/providers/openai", body)
	ctx.Params = gin.Params{{Key: "name", Value: "openai"}}

	handler.UpdateProviderHandler(ctx)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// TestUpdateProviderHandlerDecodesPayload pins the request seam: it captures the
// whole UpdateProviderDto the handler decodes (service/aiProviders.ts → PUT
// /ai/providers/:name), including the nested params, and asserts every field
// plus the name path param. A json tag drift in params fails here instead of
// silently dropping the tuning in production.
func TestUpdateProviderHandlerDecodesPayload(t *testing.T) {
	var capturedName ProviderName
	var captured UpdateProviderDto
	handler := NewHandler(&serviceMock{
		updateFn: func(name ProviderName, dto UpdateProviderDto) (ProviderDto, error) {
			capturedName = name
			captured = dto
			return ProviderDto{Name: string(name)}, nil
		},
	})
	body := bytes.NewBufferString(`{"enabled":true,"model":"llama3.1","base_url":"http://host:11434","priority":3,"params":{"timeout_seconds":120,"max_retries":4,"retry_backoff_ms":500,"keep_alive":"5m"}}`)
	ctx, rec := newTestContext(http.MethodPut, "/ai/providers/ollama", body)
	ctx.Params = gin.Params{{Key: "name", Value: "ollama"}}

	handler.UpdateProviderHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if capturedName != "ollama" {
		t.Fatalf("name path param did not reach the service: %q", capturedName)
	}
	if !captured.Enabled || captured.Model != "llama3.1" || captured.BaseURL != "http://host:11434" || captured.Priority != 3 {
		t.Fatalf("top-level fields did not decode: %+v", captured)
	}
	if captured.Params.TimeoutSeconds != 120 || captured.Params.MaxRetries != 4 ||
		captured.Params.RetryBackoffMS != 500 || captured.Params.KeepAlive != "5m" {
		t.Fatalf("nested params did not decode: %+v", captured.Params)
	}
}

func TestUpdateProviderHandlerSuccess(t *testing.T) {
	handler := NewHandler(&serviceMock{
		updateFn: func(name ProviderName, dto UpdateProviderDto) (ProviderDto, error) {
			return ProviderDto{Name: string(name), Enabled: dto.Enabled, Model: dto.Model}, nil
		},
	})
	body := bytes.NewBufferString(`{"enabled":true,"model":"llama3.1","base_url":"http://x","priority":0,"params":{"timeout_seconds":120}}`)
	ctx, rec := newTestContext(http.MethodPut, "/ai/providers/ollama", body)
	ctx.Params = gin.Params{{Key: "name", Value: "ollama"}}

	handler.UpdateProviderHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp ProviderDto
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Name != "ollama" || !resp.Enabled || resp.Model != "llama3.1" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
