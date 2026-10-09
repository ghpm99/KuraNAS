package files

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// opCapturingServiceMock embeds the full handler service mock and overrides only
// the file-operation methods to record the arguments the handler decoded from
// the request body, so the tests can assert the json tags landed in the right
// fields.
type opCapturingServiceMock struct {
	*filesHandlerServiceMock

	createParentID *int
	createName     string

	moveSourceID   int
	moveDestFolder *int
	moveDestPath   string

	deleteID        int
	deletePermanent bool

	renameID      int
	renameNewName string

	copySourceID   int
	copyDestFolder *int
	copyDestPath   string
	copyNewName    string
}

func (m *opCapturingServiceMock) CreateFolder(parentID *int, name string) (string, error) {
	m.createParentID, m.createName = parentID, name
	return "/data/nova", nil
}

func (m *opCapturingServiceMock) MoveFile(sourceID int, destinationFolderID *int, destinationPath string) (string, error) {
	m.moveSourceID, m.moveDestFolder, m.moveDestPath = sourceID, destinationFolderID, destinationPath
	return "/data/dst", nil
}

func (m *opCapturingServiceMock) DeleteFileFromDisk(id int, permanent bool) error {
	m.deleteID, m.deletePermanent = id, permanent
	return nil
}

func (m *opCapturingServiceMock) RenameFile(id int, newName string) (string, error) {
	m.renameID, m.renameNewName = id, newName
	return newName, nil
}

func (m *opCapturingServiceMock) CopyFile(sourceID int, destinationFolderID *int, destinationPath string, newName string) (string, error) {
	m.copySourceID, m.copyDestFolder, m.copyDestPath, m.copyNewName = sourceID, destinationFolderID, destinationPath, newName
	return "/data/copy", nil
}

func newOperationsRouter(mock *opCapturingServiceMock) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(mock, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.POST("/files/folder", handler.CreateFolderHandler)
	router.POST("/files/move", handler.MoveFileHandler)
	router.DELETE("/files", handler.DeleteFileHandler)
	router.POST("/files/rename", handler.RenameFileHandler)
	router.POST("/files/copy", handler.CopyFileHandler)
	return router
}

func doOpRequest(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// TestFileOperationHandlersDecodePayload pins the request seam for the whole
// file-manager command surface (frontend service/files.ts → these endpoints).
// Each case captures the args the handler decoded from the JSON body and asserts
// them, so a json tag drift (e.g. new_name → newName, source_id → sourceId)
// fails here instead of silently breaking the file operation in production.
func TestFileOperationHandlersDecodePayload(t *testing.T) {
	t.Run("create folder", func(t *testing.T) {
		mock := &opCapturingServiceMock{filesHandlerServiceMock: &filesHandlerServiceMock{}}
		router := newOperationsRouter(mock)

		rec := doOpRequest(router, http.MethodPost, "/files/folder", `{"parent_id":7,"name":"Documentos"}`)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		if mock.createParentID == nil || *mock.createParentID != 7 || mock.createName != "Documentos" {
			t.Fatalf("create payload did not decode: parentID=%v name=%q", mock.createParentID, mock.createName)
		}
	})

	t.Run("move file", func(t *testing.T) {
		mock := &opCapturingServiceMock{filesHandlerServiceMock: &filesHandlerServiceMock{}}
		router := newOperationsRouter(mock)

		rec := doOpRequest(router, http.MethodPost, "/files/move", `{"source_id":3,"destination_folder_id":9,"destination_path":"/data/dst"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if mock.moveSourceID != 3 || mock.moveDestFolder == nil || *mock.moveDestFolder != 9 || mock.moveDestPath != "/data/dst" {
			t.Fatalf("move payload did not decode: %+v", mock)
		}
	})

	t.Run("delete file with permanent query", func(t *testing.T) {
		mock := &opCapturingServiceMock{filesHandlerServiceMock: &filesHandlerServiceMock{}}
		router := newOperationsRouter(mock)

		rec := doOpRequest(router, http.MethodDelete, "/files?permanent=true", `{"id":42}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if mock.deleteID != 42 || !mock.deletePermanent {
			t.Fatalf("delete payload/query did not decode: id=%d permanent=%v", mock.deleteID, mock.deletePermanent)
		}
	})

	t.Run("rename file", func(t *testing.T) {
		mock := &opCapturingServiceMock{filesHandlerServiceMock: &filesHandlerServiceMock{}}
		router := newOperationsRouter(mock)

		rec := doOpRequest(router, http.MethodPost, "/files/rename", `{"id":5,"new_name":"relatorio.pdf"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if mock.renameID != 5 || mock.renameNewName != "relatorio.pdf" {
			t.Fatalf("rename payload did not decode: id=%d newName=%q", mock.renameID, mock.renameNewName)
		}
	})

	t.Run("copy file", func(t *testing.T) {
		mock := &opCapturingServiceMock{filesHandlerServiceMock: &filesHandlerServiceMock{}}
		router := newOperationsRouter(mock)

		rec := doOpRequest(router, http.MethodPost, "/files/copy", `{"source_id":3,"destination_folder_id":9,"destination_path":"/data/dst","new_name":"copia.pdf"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if mock.copySourceID != 3 || mock.copyDestFolder == nil || *mock.copyDestFolder != 9 ||
			mock.copyDestPath != "/data/dst" || mock.copyNewName != "copia.pdf" {
			t.Fatalf("copy payload did not decode: %+v", mock)
		}
	})
}
