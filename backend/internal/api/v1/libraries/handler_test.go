package libraries

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type handlerServiceMock struct {
	getLibrariesFn func() ([]LibraryDto, error)
	updateFn       func(category LibraryCategory, dto UpdateLibraryDto) (LibraryDto, error)
}

func (m *handlerServiceMock) GetLibraries() ([]LibraryDto, error) {
	if m.getLibrariesFn != nil {
		return m.getLibrariesFn()
	}
	return []LibraryDto{}, nil
}

func (m *handlerServiceMock) GetLibraryByCategory(category LibraryCategory) (LibraryDto, error) {
	return LibraryDto{}, nil
}

func (m *handlerServiceMock) UpdateLibrary(category LibraryCategory, dto UpdateLibraryDto) (LibraryDto, error) {
	if m.updateFn != nil {
		return m.updateFn(category, dto)
	}
	return LibraryDto{Category: string(category), Path: dto.Path}, nil
}

func (m *handlerServiceMock) ResolveLibraries() error {
	return nil
}

func newLibrariesContext(method string, url string, body *bytes.Buffer) (*gin.Context, *httptest.ResponseRecorder) {
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

func TestGetLibrariesHandlerSuccess(t *testing.T) {
	handler := NewHandler(&handlerServiceMock{
		getLibrariesFn: func() ([]LibraryDto, error) {
			return []LibraryDto{{Category: "images", Path: "/data/Imagens"}}, nil
		},
	}, nil)
	ctx, rec := newLibrariesContext(http.MethodGet, "/libraries", nil)

	handler.GetLibrariesHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGetLibrariesHandlerError(t *testing.T) {
	handler := NewHandler(&handlerServiceMock{
		getLibrariesFn: func() ([]LibraryDto, error) {
			return nil, errors.New("fetch failed")
		},
	}, nil)
	ctx, rec := newLibrariesContext(http.MethodGet, "/libraries", nil)

	handler.GetLibrariesHandler(ctx)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestUpdateLibraryHandlerSuccess(t *testing.T) {
	handler := NewHandler(&handlerServiceMock{
		updateFn: func(category LibraryCategory, dto UpdateLibraryDto) (LibraryDto, error) {
			if category != LibraryCategoryImages {
				t.Fatalf("expected images category, got %s", category)
			}
			if dto.Path != "/data/Imagens" {
				t.Fatalf("expected path to be passed through")
			}
			return LibraryDto{Category: "images", Path: dto.Path}, nil
		},
	}, nil)

	ctx, rec := newLibrariesContext(http.MethodPut, "/libraries/images", bytes.NewBufferString(`{"path":"/data/Imagens"}`))
	ctx.Params = gin.Params{{Key: "category", Value: "images"}}

	handler.UpdateLibraryHandler(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUpdateLibraryHandlerInvalidCategory(t *testing.T) {
	handler := NewHandler(&handlerServiceMock{}, nil)
	ctx, rec := newLibrariesContext(http.MethodPut, "/libraries/unknown", bytes.NewBufferString(`{"path":"/data/unknown"}`))
	ctx.Params = gin.Params{{Key: "category", Value: "unknown"}}

	handler.UpdateLibraryHandler(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateLibraryHandlerInvalidPath(t *testing.T) {
	handler := NewHandler(&handlerServiceMock{
		updateFn: func(category LibraryCategory, dto UpdateLibraryDto) (LibraryDto, error) {
			return LibraryDto{}, ErrPathNotSubfolder
		},
	}, nil)

	ctx, rec := newLibrariesContext(http.MethodPut, "/libraries/images", bytes.NewBufferString(`{"path":"/outside"}`))
	ctx.Params = gin.Params{{Key: "category", Value: "images"}}

	handler.UpdateLibraryHandler(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateLibraryHandlerInvalidJSON(t *testing.T) {
	handler := NewHandler(&handlerServiceMock{}, nil)
	ctx, rec := newLibrariesContext(http.MethodPut, "/libraries/images", bytes.NewBufferString(`{`))
	ctx.Params = gin.Params{{Key: "category", Value: "images"}}

	handler.UpdateLibraryHandler(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateLibraryHandlerUnexpectedError(t *testing.T) {
	handler := NewHandler(&handlerServiceMock{
		updateFn: func(category LibraryCategory, dto UpdateLibraryDto) (LibraryDto, error) {
			return LibraryDto{}, errors.New("boom")
		},
	}, nil)

	ctx, rec := newLibrariesContext(http.MethodPut, "/libraries/images", bytes.NewBufferString(`{"path":"/data/Imagens"}`))
	ctx.Params = gin.Params{{Key: "category", Value: "images"}}

	handler.UpdateLibraryHandler(ctx)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
