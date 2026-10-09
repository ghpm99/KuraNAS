package configuration

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestGetEnvConfigHandler(t *testing.T) {
	l := &loggerMock{}
	h := NewHandler(&serviceMock{
		getEnvConfigFn: func() (EnvConfigDto, error) {
			return EnvConfigDto{
				Fields:          []EnvFieldDto{{Key: "LANGUAGE", Group: "general", Kind: "string", Value: "pt-BR"}},
				RestartRequired: true,
			}, nil
		},
	}, l)
	ctx, rec := newTestContext(http.MethodGet, nil)

	h.GetEnvConfigHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"restart_required":true`) {
		t.Fatalf("expected restart flag, got %s", rec.Body.String())
	}
}

func TestGetEnvConfigHandlerNilService(t *testing.T) {
	h := NewHandler(nil, &loggerMock{})
	ctx, rec := newTestContext(http.MethodGet, nil)

	h.GetEnvConfigHandler(ctx)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on nil service, got %d", rec.Code)
	}
}

func TestGetEnvConfigHandlerError(t *testing.T) {
	h := NewHandler(&serviceMock{
		getEnvConfigFn: func() (EnvConfigDto, error) { return EnvConfigDto{}, errors.New("boom") },
	}, &loggerMock{})
	ctx, rec := newTestContext(http.MethodGet, nil)

	h.GetEnvConfigHandler(ctx)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestUpdateEnvConfigHandlerBadJSON(t *testing.T) {
	h := NewHandler(&serviceMock{}, &loggerMock{})
	ctx, rec := newTestContext(http.MethodPut, bytes.NewBufferString("{bad"))

	h.UpdateEnvConfigHandler(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateEnvConfigHandlerSuccess(t *testing.T) {
	h := NewHandler(&serviceMock{
		updateEnvConfigFn: func(request UpdateEnvConfigRequest) (EnvConfigDto, error) {
			return EnvConfigDto{RestartRequired: true}, nil
		},
	}, &loggerMock{})
	ctx, rec := newTestContext(http.MethodPut, bytes.NewBufferString(`{"changes":{"LANGUAGE":"en-US"}}`))

	h.UpdateEnvConfigHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUpdateEnvConfigHandlerErrorMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"invalid key", ErrInvalidEnvKey, http.StatusBadRequest},
		{"invalid value", ErrInvalidEnvValue, http.StatusBadRequest},
		{"confirmation", ErrEnvConfirmationRequired, http.StatusBadRequest},
		{"generic", errors.New("disk full"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandler(&serviceMock{
				updateEnvConfigFn: func(request UpdateEnvConfigRequest) (EnvConfigDto, error) {
					return EnvConfigDto{}, tc.err
				},
			}, &loggerMock{})
			ctx, rec := newTestContext(http.MethodPut, bytes.NewBufferString(`{"changes":{}}`))

			h.UpdateEnvConfigHandler(ctx)

			if rec.Code != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, rec.Code)
			}
		})
	}
}

func TestTestEnvDatabaseHandler(t *testing.T) {
	h := NewHandler(&serviceMock{}, &loggerMock{})
	ctx, rec := newTestContext(http.MethodPost, bytes.NewBufferString(`{"host":"h","port":"1","user":"u","name":"n"}`))

	h.TestEnvDatabaseHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("expected ok true, got %s", rec.Body.String())
	}
}

func TestTestEnvDatabaseHandlerFailure(t *testing.T) {
	h := NewHandler(&serviceMock{
		testDatabaseFn: func(request TestDatabaseRequest) error { return errors.New("refused") },
	}, &loggerMock{})
	ctx, rec := newTestContext(http.MethodPost, bytes.NewBufferString(`{"host":"h"}`))

	h.TestEnvDatabaseHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with ok=false, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"ok":false`) {
		t.Fatalf("expected ok false, got %s", rec.Body.String())
	}
}

func TestTestEnvDatabaseHandlerBadJSON(t *testing.T) {
	h := NewHandler(&serviceMock{}, &loggerMock{})
	ctx, rec := newTestContext(http.MethodPost, bytes.NewBufferString("{bad"))

	h.TestEnvDatabaseHandler(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTestEnvPathHandler(t *testing.T) {
	h := NewHandler(&serviceMock{}, &loggerMock{})
	ctx, rec := newTestContext(http.MethodPost, bytes.NewBufferString(`{"path":"/data"}`))

	h.TestEnvPathHandler(ctx)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("expected ok true, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestTestEnvPathHandlerFailure(t *testing.T) {
	h := NewHandler(&serviceMock{
		testPathFn: func(request TestPathRequest) error { return errors.New("missing") },
	}, &loggerMock{})
	ctx, rec := newTestContext(http.MethodPost, bytes.NewBufferString(`{"path":"/nope"}`))

	h.TestEnvPathHandler(ctx)

	if !strings.Contains(rec.Body.String(), `"ok":false`) {
		t.Fatalf("expected ok false, got %s", rec.Body.String())
	}
}

func TestTestEnvPathHandlerBadJSON(t *testing.T) {
	h := NewHandler(&serviceMock{}, &loggerMock{})
	ctx, rec := newTestContext(http.MethodPost, bytes.NewBufferString("{bad"))

	h.TestEnvPathHandler(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
