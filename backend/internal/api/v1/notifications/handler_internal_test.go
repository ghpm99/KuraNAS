package notifications

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type handlerServiceMock struct{}

func (m *handlerServiceMock) GetNotificationByID(id int) (NotificationDto, error) {
	return NotificationDto{ID: id, Type: "info", Title: "t", Message: "m"}, nil
}
func (m *handlerServiceMock) ListNotifications(filter NotificationFilter, page int, pageSize int) (utils.PaginationResponse[NotificationDto], error) {
	return utils.PaginationResponse[NotificationDto]{
		Items:      []NotificationDto{{ID: 1, Type: "info", Title: "t"}},
		Pagination: utils.Pagination{Page: page, PageSize: pageSize},
	}, nil
}
func (m *handlerServiceMock) MarkAsRead(id int) error { return nil }
func (m *handlerServiceMock) MarkAllAsRead() error    { return nil }
func (m *handlerServiceMock) GetUnreadCount() (UnreadCountDto, error) {
	return UnreadCountDto{UnreadCount: 3}, nil
}
func (m *handlerServiceMock) GroupOrCreate(dto CreateNotificationDto) (NotificationDto, error) {
	return NotificationDto{ID: 1}, nil
}
func (m *handlerServiceMock) CleanupOldNotifications() error { return nil }

type handlerErrServiceMock struct{ handlerServiceMock }

func (m *handlerErrServiceMock) GetNotificationByID(id int) (NotificationDto, error) {
	return NotificationDto{}, errors.New("get failed")
}
func (m *handlerErrServiceMock) ListNotifications(filter NotificationFilter, page int, pageSize int) (utils.PaginationResponse[NotificationDto], error) {
	return utils.PaginationResponse[NotificationDto]{}, errors.New("list failed")
}
func (m *handlerErrServiceMock) MarkAsRead(id int) error { return errors.New("mark failed") }
func (m *handlerErrServiceMock) MarkAllAsRead() error    { return errors.New("mark all failed") }
func (m *handlerErrServiceMock) GetUnreadCount() (UnreadCountDto, error) {
	return UnreadCountDto{}, errors.New("count failed")
}

type handlerNotFoundServiceMock struct{ handlerServiceMock }

func (m *handlerNotFoundServiceMock) GetNotificationByID(id int) (NotificationDto, error) {
	return NotificationDto{}, ErrNotificationNotFound
}
func (m *handlerNotFoundServiceMock) MarkAsRead(id int) error {
	return ErrNotificationNotFound
}

type handlerInvalidIDServiceMock struct{ handlerServiceMock }

func (m *handlerInvalidIDServiceMock) GetNotificationByID(id int) (NotificationDto, error) {
	return NotificationDto{}, ErrInvalidNotificationID
}
func (m *handlerInvalidIDServiceMock) MarkAsRead(id int) error {
	return ErrInvalidNotificationID
}

func TestHandlerEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&handlerServiceMock{})
	router := gin.New()

	router.GET("/notifications", handler.ListNotificationsHandler)
	router.GET("/notifications/unread-count", handler.GetUnreadCountHandler)
	router.GET("/notifications/:id", handler.GetNotificationByIDHandler)
	router.POST("/notifications/:id/read", handler.MarkAsReadHandler)
	router.POST("/notifications/read-all", handler.MarkAllAsReadHandler)

	tests := []struct {
		method string
		path   string
		code   int
	}{
		{http.MethodGet, "/notifications", http.StatusOK},
		{http.MethodGet, "/notifications?page=1&page_size=10", http.StatusOK},
		{http.MethodGet, "/notifications?type=info&is_read=true", http.StatusOK},
		{http.MethodGet, "/notifications/1", http.StatusOK},
		{http.MethodGet, "/notifications/unread-count", http.StatusOK},
		{http.MethodPost, "/notifications/1/read", http.StatusOK},
		{http.MethodPost, "/notifications/read-all", http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d, body=%s", tc.code, w.Code, w.Body.String())
			}
		})
	}
}

func TestHandlerErrorResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&handlerErrServiceMock{})
	router := gin.New()

	router.GET("/notifications", handler.ListNotificationsHandler)
	router.GET("/notifications/unread-count", handler.GetUnreadCountHandler)
	router.GET("/notifications/:id", handler.GetNotificationByIDHandler)
	router.POST("/notifications/:id/read", handler.MarkAsReadHandler)
	router.POST("/notifications/read-all", handler.MarkAllAsReadHandler)

	tests := []struct {
		method string
		path   string
		code   int
	}{
		{http.MethodGet, "/notifications", http.StatusInternalServerError},
		{http.MethodGet, "/notifications/1", http.StatusInternalServerError},
		{http.MethodGet, "/notifications/unread-count", http.StatusInternalServerError},
		{http.MethodPost, "/notifications/1/read", http.StatusInternalServerError},
		{http.MethodPost, "/notifications/read-all", http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d, body=%s", tc.code, w.Code, w.Body.String())
			}
		})
	}
}

func TestHandlerNotFoundResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&handlerNotFoundServiceMock{})
	router := gin.New()

	router.GET("/notifications/:id", handler.GetNotificationByIDHandler)
	router.POST("/notifications/:id/read", handler.MarkAsReadHandler)

	tests := []struct {
		method string
		path   string
		code   int
	}{
		{http.MethodGet, "/notifications/1", http.StatusNotFound},
		{http.MethodPost, "/notifications/1/read", http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d, body=%s", tc.code, w.Code, w.Body.String())
			}
		})
	}
}

func TestHandlerInvalidIDResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&handlerInvalidIDServiceMock{})
	router := gin.New()

	router.GET("/notifications/:id", handler.GetNotificationByIDHandler)
	router.POST("/notifications/:id/read", handler.MarkAsReadHandler)

	tests := []struct {
		method string
		path   string
		code   int
	}{
		{http.MethodGet, "/notifications/abc", http.StatusBadRequest},
		{http.MethodPost, "/notifications/abc/read", http.StatusBadRequest},
		{http.MethodGet, "/notifications/1", http.StatusBadRequest},
		{http.MethodPost, "/notifications/1/read", http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d, body=%s", tc.code, w.Code, w.Body.String())
			}
		})
	}
}

func TestHandlerInvalidQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&handlerServiceMock{})
	router := gin.New()

	router.GET("/notifications", handler.ListNotificationsHandler)

	tests := []struct {
		path string
		code int
	}{
		{"/notifications?page=abc", http.StatusBadRequest},
		{"/notifications?page=1&page_size=abc", http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d, body=%s", tc.code, w.Code, w.Body.String())
			}
		})
	}
}
