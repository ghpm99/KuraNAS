package notifications

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"

	"github.com/DATA-DOG/go-sqlmock"
)

type repoMock struct {
	dbContext             *database.DbContext
	createNotificationFn  func(tx *sql.Tx, model NotificationModel) (NotificationModel, error)
	getNotificationByIDFn func(id int) (NotificationModel, error)
	listNotificationsFn   func(filter NotificationFilter, page int, pageSize int) (utils.PaginationResponse[NotificationModel], error)
	markAsReadFn          func(tx *sql.Tx, id int) error
	markAllAsReadFn       func(tx *sql.Tx) error
	getUnreadCountFn      func() (int, error)
	findActiveGroupFn     func(tx *sql.Tx, groupKey string, notifType string, windowSeconds int) (NotificationModel, error)
	updateGroupCountFn    func(tx *sql.Tx, id int, count int, message string) error
	deleteOldFn           func(tx *sql.Tx) error
}

func (r *repoMock) GetDbContext() *database.DbContext { return r.dbContext }
func (r *repoMock) CreateNotification(tx *sql.Tx, model NotificationModel) (NotificationModel, error) {
	if r.createNotificationFn != nil {
		return r.createNotificationFn(tx, model)
	}
	model.ID = 1
	model.CreatedAt = time.Now()
	return model, nil
}
func (r *repoMock) GetNotificationByID(id int) (NotificationModel, error) {
	if r.getNotificationByIDFn != nil {
		return r.getNotificationByIDFn(id)
	}
	return NotificationModel{ID: id, Type: "info", Title: "t", Message: "m"}, nil
}
func (r *repoMock) ListNotifications(filter NotificationFilter, page int, pageSize int) (utils.PaginationResponse[NotificationModel], error) {
	if r.listNotificationsFn != nil {
		return r.listNotificationsFn(filter, page, pageSize)
	}
	return utils.PaginationResponse[NotificationModel]{Items: []NotificationModel{}}, nil
}
func (r *repoMock) MarkAsRead(tx *sql.Tx, id int) error {
	if r.markAsReadFn != nil {
		return r.markAsReadFn(tx, id)
	}
	return nil
}
func (r *repoMock) MarkAllAsRead(tx *sql.Tx) error {
	if r.markAllAsReadFn != nil {
		return r.markAllAsReadFn(tx)
	}
	return nil
}
func (r *repoMock) GetUnreadCount() (int, error) {
	if r.getUnreadCountFn != nil {
		return r.getUnreadCountFn()
	}
	return 5, nil
}
func (r *repoMock) FindActiveGroup(tx *sql.Tx, groupKey string, notifType string, windowSeconds int) (NotificationModel, error) {
	if r.findActiveGroupFn != nil {
		return r.findActiveGroupFn(tx, groupKey, notifType, windowSeconds)
	}
	return NotificationModel{}, sql.ErrNoRows
}
func (r *repoMock) UpdateGroupCount(tx *sql.Tx, id int, count int, message string) error {
	if r.updateGroupCountFn != nil {
		return r.updateGroupCountFn(tx, id, count, message)
	}
	return nil
}
func (r *repoMock) DeleteOldNotifications(tx *sql.Tx) error {
	if r.deleteOldFn != nil {
		return r.deleteOldFn(tx)
	}
	return nil
}

func newTestService(t *testing.T, mock *repoMock) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	mock.dbContext = database.NewDbContext(db)
	return &Service{Repository: mock}, sqlMock
}

func TestRepoMockCompileOnly(t *testing.T) {
	var _ RepositoryInterface = (*repoMock)(nil)
}

func TestNewService(t *testing.T) {
	repo := &repoMock{}
	svc := NewService(repo)
	typed, ok := svc.(*Service)
	if !ok || typed.Repository != repo {
		t.Fatalf("expected concrete service with repository")
	}
}

func TestServiceGetNotificationByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &repoMock{
			getNotificationByIDFn: func(id int) (NotificationModel, error) {
				return NotificationModel{ID: id, Type: "info", Title: "test", Message: "msg"}, nil
			},
		}
		service, _ := newTestService(t, mock)

		dto, err := service.GetNotificationByID(1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dto.ID != 1 || dto.Title != "test" {
			t.Fatalf("unexpected dto: %+v", dto)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		mock := &repoMock{}
		service, _ := newTestService(t, mock)

		_, err := service.GetNotificationByID(0)
		if !errors.Is(err, ErrInvalidNotificationID) {
			t.Fatalf("expected ErrInvalidNotificationID, got %v", err)
		}

		_, err = service.GetNotificationByID(-1)
		if !errors.Is(err, ErrInvalidNotificationID) {
			t.Fatalf("expected ErrInvalidNotificationID, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock := &repoMock{
			getNotificationByIDFn: func(id int) (NotificationModel, error) {
				return NotificationModel{}, sql.ErrNoRows
			},
		}
		service, _ := newTestService(t, mock)

		_, err := service.GetNotificationByID(1)
		if !errors.Is(err, ErrNotificationNotFound) {
			t.Fatalf("expected ErrNotificationNotFound, got %v", err)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		mock := &repoMock{
			getNotificationByIDFn: func(id int) (NotificationModel, error) {
				return NotificationModel{}, errors.New("db error")
			},
		}
		service, _ := newTestService(t, mock)

		_, err := service.GetNotificationByID(1)
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestServiceListNotifications(t *testing.T) {
	t.Run("success with defaults", func(t *testing.T) {
		mock := &repoMock{
			listNotificationsFn: func(filter NotificationFilter, page int, pageSize int) (utils.PaginationResponse[NotificationModel], error) {
				if page != 1 || pageSize != 20 {
					t.Fatalf("expected defaults page=1, pageSize=20, got page=%d, pageSize=%d", page, pageSize)
				}
				return utils.PaginationResponse[NotificationModel]{
					Items: []NotificationModel{
						{ID: 1, Type: "info", Title: "a", Message: "b"},
						{ID: 2, Type: "success", Title: "c", Message: "d"},
					},
					Pagination: utils.Pagination{Page: page, PageSize: pageSize},
				}, nil
			},
		}
		service, _ := newTestService(t, mock)

		result, err := service.ListNotifications(NotificationFilter{}, 0, 0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(result.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(result.Items))
		}
	})

	t.Run("repo error", func(t *testing.T) {
		mock := &repoMock{
			listNotificationsFn: func(filter NotificationFilter, page int, pageSize int) (utils.PaginationResponse[NotificationModel], error) {
				return utils.PaginationResponse[NotificationModel]{}, errors.New("list error")
			},
		}
		service, _ := newTestService(t, mock)

		_, err := service.ListNotifications(NotificationFilter{}, 1, 10)
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestServiceGetUnreadCount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &repoMock{
			getUnreadCountFn: func() (int, error) {
				return 7, nil
			},
		}
		service, _ := newTestService(t, mock)

		dto, err := service.GetUnreadCount()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dto.UnreadCount != 7 {
			t.Fatalf("expected 7, got %d", dto.UnreadCount)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &repoMock{
			getUnreadCountFn: func() (int, error) {
				return 0, errors.New("count error")
			},
		}
		service, _ := newTestService(t, mock)

		_, err := service.GetUnreadCount()
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestServiceMarkAsRead(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		mock := &repoMock{}
		service, _ := newTestService(t, mock)

		err := service.MarkAsRead(0)
		if !errors.Is(err, ErrInvalidNotificationID) {
			t.Fatalf("expected ErrInvalidNotificationID, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock := &repoMock{
			getNotificationByIDFn: func(id int) (NotificationModel, error) {
				return NotificationModel{}, sql.ErrNoRows
			},
		}
		service, _ := newTestService(t, mock)

		err := service.MarkAsRead(1)
		if !errors.Is(err, ErrNotificationNotFound) {
			t.Fatalf("expected ErrNotificationNotFound, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		mock := &repoMock{
			getNotificationByIDFn: func(id int) (NotificationModel, error) {
				return NotificationModel{ID: id}, nil
			},
		}
		service, sqlMock := newTestService(t, mock)
		sqlMock.ExpectBegin()
		sqlMock.ExpectCommit()

		err := service.MarkAsRead(1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("repo get error", func(t *testing.T) {
		mock := &repoMock{
			getNotificationByIDFn: func(id int) (NotificationModel, error) {
				return NotificationModel{}, errors.New("db error")
			},
		}
		service, _ := newTestService(t, mock)

		err := service.MarkAsRead(1)
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestServiceMarkAllAsRead(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &repoMock{}
		service, sqlMock := newTestService(t, mock)
		sqlMock.ExpectBegin()
		sqlMock.ExpectCommit()

		err := service.MarkAllAsRead()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestServiceCleanupOldNotifications(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &repoMock{}
		service, sqlMock := newTestService(t, mock)
		sqlMock.ExpectBegin()
		sqlMock.ExpectCommit()

		err := service.CleanupOldNotifications()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &repoMock{
			deleteOldFn: func(tx *sql.Tx) error {
				return errors.New("delete error")
			},
		}
		service, sqlMock := newTestService(t, mock)
		sqlMock.ExpectBegin()
		sqlMock.ExpectRollback()

		err := service.CleanupOldNotifications()
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestServiceGroupOrCreate(t *testing.T) {
	t.Run("error type creates ungrouped", func(t *testing.T) {
		created := false
		mock := &repoMock{
			createNotificationFn: func(tx *sql.Tx, model NotificationModel) (NotificationModel, error) {
				created = true
				if model.IsGrouped {
					t.Fatalf("expected ungrouped notification for error type")
				}
				model.ID = 1
				model.CreatedAt = time.Now()
				return model, nil
			},
		}
		service, sqlMock := newTestService(t, mock)
		sqlMock.ExpectBegin()
		sqlMock.ExpectCommit()

		dto := CreateNotificationDto{
			Type:     string(NotificationTypeError),
			Title:    "error",
			Message:  "something failed",
			GroupKey: "key",
		}
		result, err := service.GroupOrCreate(dto)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !created {
			t.Fatalf("expected create to be called")
		}
		if result.ID != 1 {
			t.Fatalf("expected ID 1, got %d", result.ID)
		}
	})

	t.Run("empty group key creates ungrouped", func(t *testing.T) {
		mock := &repoMock{
			createNotificationFn: func(tx *sql.Tx, model NotificationModel) (NotificationModel, error) {
				if model.IsGrouped {
					t.Fatalf("expected ungrouped notification")
				}
				model.ID = 2
				model.CreatedAt = time.Now()
				return model, nil
			},
		}
		service, sqlMock := newTestService(t, mock)
		sqlMock.ExpectBegin()
		sqlMock.ExpectCommit()

		dto := CreateNotificationDto{
			Type:    string(NotificationTypeInfo),
			Title:   "info",
			Message: "some info",
		}
		_, err := service.GroupOrCreate(dto)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("no active group creates new grouped", func(t *testing.T) {
		mock := &repoMock{
			findActiveGroupFn: func(tx *sql.Tx, groupKey string, notifType string, windowSeconds int) (NotificationModel, error) {
				return NotificationModel{}, sql.ErrNoRows
			},
			createNotificationFn: func(tx *sql.Tx, model NotificationModel) (NotificationModel, error) {
				if !model.IsGrouped {
					t.Fatalf("expected grouped notification")
				}
				model.ID = 3
				model.CreatedAt = time.Now()
				return model, nil
			},
		}
		service, sqlMock := newTestService(t, mock)
		sqlMock.ExpectBegin()
		sqlMock.ExpectCommit()

		dto := CreateNotificationDto{
			Type:     string(NotificationTypeInfo),
			Title:    "grouped",
			Message:  "msg",
			GroupKey: "files",
		}
		result, err := service.GroupOrCreate(dto)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.ID != 3 {
			t.Fatalf("expected ID 3, got %d", result.ID)
		}
	})

	t.Run("active group found increments count", func(t *testing.T) {
		mock := &repoMock{
			findActiveGroupFn: func(tx *sql.Tx, groupKey string, notifType string, windowSeconds int) (NotificationModel, error) {
				return NotificationModel{
					ID:         10,
					Type:       "info",
					Title:      "grouped",
					Message:    "1 grouped",
					GroupCount: 1,
					GroupKey:   sql.NullString{String: "files", Valid: true},
					IsGrouped:  true,
				}, nil
			},
			updateGroupCountFn: func(tx *sql.Tx, id int, count int, message string) error {
				if id != 10 || count != 2 || message != "2 grouped" {
					t.Fatalf("unexpected update: id=%d, count=%d, message=%s", id, count, message)
				}
				return nil
			},
		}
		service, sqlMock := newTestService(t, mock)
		sqlMock.ExpectBegin()
		sqlMock.ExpectCommit()

		dto := CreateNotificationDto{
			Type:     string(NotificationTypeInfo),
			Title:    "grouped",
			Message:  "new msg",
			GroupKey: "files",
		}
		result, err := service.GroupOrCreate(dto)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.GroupCount != 2 {
			t.Fatalf("expected group count 2, got %d", result.GroupCount)
		}
		if result.Message != "2 grouped" {
			t.Fatalf("expected message '2 grouped', got '%s'", result.Message)
		}
	})

	t.Run("find active group error propagates", func(t *testing.T) {
		mock := &repoMock{
			findActiveGroupFn: func(tx *sql.Tx, groupKey string, notifType string, windowSeconds int) (NotificationModel, error) {
				return NotificationModel{}, errors.New("find error")
			},
		}
		service, sqlMock := newTestService(t, mock)
		sqlMock.ExpectBegin()
		sqlMock.ExpectRollback()

		dto := CreateNotificationDto{
			Type:     string(NotificationTypeInfo),
			Title:    "t",
			Message:  "m",
			GroupKey: "key",
		}
		_, err := service.GroupOrCreate(dto)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("create error propagates", func(t *testing.T) {
		mock := &repoMock{
			createNotificationFn: func(tx *sql.Tx, model NotificationModel) (NotificationModel, error) {
				return NotificationModel{}, errors.New("create error")
			},
		}
		service, sqlMock := newTestService(t, mock)
		sqlMock.ExpectBegin()
		sqlMock.ExpectRollback()

		dto := CreateNotificationDto{
			Type:    string(NotificationTypeInfo),
			Title:   "t",
			Message: "m",
		}
		_, err := service.GroupOrCreate(dto)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("update group count error propagates", func(t *testing.T) {
		mock := &repoMock{
			findActiveGroupFn: func(tx *sql.Tx, groupKey string, notifType string, windowSeconds int) (NotificationModel, error) {
				return NotificationModel{ID: 10, GroupCount: 1}, nil
			},
			updateGroupCountFn: func(tx *sql.Tx, id int, count int, message string) error {
				return errors.New("update error")
			},
		}
		service, sqlMock := newTestService(t, mock)
		sqlMock.ExpectBegin()
		sqlMock.ExpectRollback()

		dto := CreateNotificationDto{
			Type:     string(NotificationTypeInfo),
			Title:    "t",
			Message:  "m",
			GroupKey: "key",
		}
		_, err := service.GroupOrCreate(dto)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("with metadata", func(t *testing.T) {
		mock := &repoMock{
			createNotificationFn: func(tx *sql.Tx, model NotificationModel) (NotificationModel, error) {
				if !model.Metadata.Valid {
					t.Fatalf("expected valid metadata")
				}
				model.ID = 5
				model.CreatedAt = time.Now()
				return model, nil
			},
		}
		service, sqlMock := newTestService(t, mock)
		sqlMock.ExpectBegin()
		sqlMock.ExpectCommit()

		dto := CreateNotificationDto{
			Type:     string(NotificationTypeInfo),
			Title:    "meta",
			Message:  "msg",
			Metadata: map[string]string{"key": "value"},
		}
		_, err := service.GroupOrCreate(dto)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
