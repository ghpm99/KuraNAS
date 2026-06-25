package watchfolders

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type handlerServiceMock struct {
	getFn    func() ([]WatchFolderDto, error)
	createFn func(dto CreateWatchFolderDto) (WatchFolderDto, error)
	updateFn func(id int, dto UpdateWatchFolderDto) (WatchFolderDto, error)
	deleteFn func(id int) error
}

func (m *handlerServiceMock) GetWatchFolders() ([]WatchFolderDto, error) {
	if m.getFn != nil {
		return m.getFn()
	}
	return []WatchFolderDto{}, nil
}

func (m *handlerServiceMock) CreateWatchFolder(dto CreateWatchFolderDto) (WatchFolderDto, error) {
	if m.createFn != nil {
		return m.createFn(dto)
	}
	return WatchFolderDto{ID: 1, Path: dto.Path, Label: dto.Label, Enabled: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (m *handlerServiceMock) UpdateWatchFolder(id int, dto UpdateWatchFolderDto) (WatchFolderDto, error) {
	if m.updateFn != nil {
		return m.updateFn(id, dto)
	}
	return WatchFolderDto{ID: id, Path: "/tmp/watch", Enabled: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (m *handlerServiceMock) DeleteWatchFolder(id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

func (m *handlerServiceMock) GetEnabledWatchFolders() ([]WatchFolderModel, error) { return nil, nil }
func (m *handlerServiceMock) UpdateWatchFolderLastScan(id int, lastScanAt time.Time) error {
	return nil
}

func newWatchFoldersContext(method string, url string, body *bytes.Buffer) (*gin.Context, *httptest.ResponseRecorder) {
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

func TestGetWatchFoldersHandlerSuccess(t *testing.T) {
	handler := NewHandler(&handlerServiceMock{getFn: func() ([]WatchFolderDto, error) {
		return []WatchFolderDto{{ID: 1, Path: "/tmp/watch", Enabled: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}}, nil
	}}, nil)
	ctx, rec := newWatchFoldersContext(http.MethodGet, "/watch-folders", nil)

	handler.GetWatchFoldersHandler(ctx)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCreateWatchFolderHandlerSuccess(t *testing.T) {
	handler := NewHandler(&handlerServiceMock{}, nil)
	ctx, rec := newWatchFoldersContext(http.MethodPost, "/watch-folders", bytes.NewBufferString(`{"path":"/tmp/watch"}`))

	handler.CreateWatchFolderHandler(ctx)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

// TestCreateWatchFolderHandlerDecodesPayload pins the request seam: it proves
// the handler decodes the exact JSON the frontend sends (POST /watch-folders)
// into CreateWatchFolderDto. A json tag drift fails here instead of breaking
// the frontend integration silently.
func TestCreateWatchFolderHandlerDecodesPayload(t *testing.T) {
	var captured CreateWatchFolderDto
	handler := NewHandler(&handlerServiceMock{createFn: func(dto CreateWatchFolderDto) (WatchFolderDto, error) {
		captured = dto
		return WatchFolderDto{ID: 7, Path: dto.Path, Label: dto.Label, Enabled: true}, nil
	}}, nil)
	ctx, rec := newWatchFoldersContext(http.MethodPost, "/watch-folders", bytes.NewBufferString(`{"path":"/tmp/watch","label":"Fotos"}`))

	handler.CreateWatchFolderHandler(ctx)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if captured != (CreateWatchFolderDto{Path: "/tmp/watch", Label: "Fotos"}) {
		t.Fatalf("handler decoded create payload into %#v", captured)
	}
}

// TestUpdateWatchFolderHandlerDecodesPayload proves the PUT body decodes into
// the optional pointer fields of UpdateWatchFolderDto, and the id path param is
// forwarded to the service.
func TestUpdateWatchFolderHandlerDecodesPayload(t *testing.T) {
	var capturedID int
	var captured UpdateWatchFolderDto
	handler := NewHandler(&handlerServiceMock{updateFn: func(id int, dto UpdateWatchFolderDto) (WatchFolderDto, error) {
		capturedID = id
		captured = dto
		return WatchFolderDto{ID: id, Path: "/tmp/watch", Enabled: false}, nil
	}}, nil)
	ctx, rec := newWatchFoldersContext(http.MethodPut, "/watch-folders/4", bytes.NewBufferString(`{"label":"Backup","enabled":false}`))
	ctx.AddParam("id", "4")

	handler.UpdateWatchFolderHandler(ctx)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if capturedID != 4 {
		t.Fatalf("expected id 4 forwarded, got %d", capturedID)
	}
	if captured.Path != nil {
		t.Fatalf("path must stay nil when omitted, got %v", *captured.Path)
	}
	if captured.Label == nil || *captured.Label != "Backup" {
		t.Fatalf("label decoded as %v", captured.Label)
	}
	if captured.Enabled == nil || *captured.Enabled != false {
		t.Fatalf("enabled decoded as %v", captured.Enabled)
	}
}

func TestCreateWatchFolderHandlerBadRequest(t *testing.T) {
	handler := NewHandler(&handlerServiceMock{}, nil)
	ctx, rec := newWatchFoldersContext(http.MethodPost, "/watch-folders", bytes.NewBufferString(`{`))

	handler.CreateWatchFolderHandler(ctx)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateWatchFolderHandlerNotFound(t *testing.T) {
	handler := NewHandler(&handlerServiceMock{updateFn: func(id int, dto UpdateWatchFolderDto) (WatchFolderDto, error) {
		return WatchFolderDto{}, ErrWatchFolderNotFound
	}}, nil)
	ctx, rec := newWatchFoldersContext(http.MethodPut, "/watch-folders/3", bytes.NewBufferString(`{"enabled":false}`))
	ctx.AddParam("id", "3")

	handler.UpdateWatchFolderHandler(ctx)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteWatchFolderHandlerSuccess(t *testing.T) {
	handler := NewHandler(&handlerServiceMock{}, nil)
	ctx, rec := newWatchFoldersContext(http.MethodDelete, "/watch-folders/1", nil)
	ctx.AddParam("id", "1")

	handler.DeleteWatchFolderHandler(ctx)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestDeleteWatchFolderHandlerServerError(t *testing.T) {
	handler := NewHandler(&handlerServiceMock{deleteFn: func(id int) error {
		return errors.New("boom")
	}}, nil)
	ctx, rec := newWatchFoldersContext(http.MethodDelete, "/watch-folders/1", nil)
	ctx.AddParam("id", "1")

	handler.DeleteWatchFolderHandler(ctx)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
