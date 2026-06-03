package notifications

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/notifications"
	"nas-go/api/pkg/utils"

	"github.com/DATA-DOG/go-sqlmock"
)

func newNotificationsRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := NewRepository(database.NewDbContext(db))
	return repo, mock, db
}

func notificationColumns() []string {
	return []string{
		"id",
		"type",
		"title",
		"message",
		"metadata",
		"is_read",
		"created_at",
		"group_key",
		"group_count",
		"is_grouped",
	}
}

func TestNewRepositoryAndGetDbContext(t *testing.T) {
	repo, _, _ := newNotificationsRepoWithMock(t)
	if repo.GetDbContext() == nil {
		t.Fatalf("expected db context to be set")
	}
}

func TestRepositoryCreateNotificationAndErrors(t *testing.T) {
	repo, mock, db := newNotificationsRepoWithMock(t)
	now := time.Now()
	model := NotificationModel{
		Type:       "info",
		Title:      "Title",
		Message:    "Message",
		Metadata:   sql.NullString{String: `{"k":"v"}`, Valid: true},
		IsRead:     false,
		GroupKey:   sql.NullString{String: "group_1", Valid: true},
		GroupCount: 1,
		IsGrouped:  true,
	}

	mock.ExpectBegin()
	tx, _ := db.Begin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertNotificationQuery)).
		WithArgs(model.Type, model.Title, model.Message, model.Metadata, model.IsRead, model.GroupKey, model.GroupCount, model.IsGrouped).
		WillReturnRows(sqlmock.NewRows(notificationColumns()).
			AddRow(1, model.Type, model.Title, model.Message, model.Metadata, model.IsRead, now, model.GroupKey, model.GroupCount, model.IsGrouped))

	created, err := repo.CreateNotification(tx, model)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.ID != 1 || created.Title != model.Title {
		t.Fatalf("unexpected created notification: %+v", created)
	}
	mock.ExpectRollback()
	_ = tx.Rollback()

	mock.ExpectBegin()
	txErr, _ := db.Begin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertNotificationQuery)).
		WillReturnError(errors.New("insert failed"))
	if _, err := repo.CreateNotification(txErr, model); err == nil || !regexp.MustCompile(`CreateNotification:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped create error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txErr.Rollback()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryGetNotificationByIDAndErrors(t *testing.T) {
	repo, mock, _ := newNotificationsRepoWithMock(t)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetNotificationByIDQuery)).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows(notificationColumns()).
			AddRow(2, "success", "A", "B", sql.NullString{}, false, now, sql.NullString{}, 1, false))
	mock.ExpectRollback()

	model, err := repo.GetNotificationByID(2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if model.ID != 2 {
		t.Fatalf("unexpected model: %+v", model)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetNotificationByIDQuery)).
		WithArgs(999).
		WillReturnError(errors.New("select failed"))
	mock.ExpectRollback()

	if _, err := repo.GetNotificationByID(999); err == nil || !regexp.MustCompile(`GetNotificationByID:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped GetNotificationByID error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryListNotificationsAndErrors(t *testing.T) {
	repo, mock, _ := newNotificationsRepoWithMock(t)
	now := time.Now()

	filter := NotificationFilter{
		Type:   utils.Optional[string]{HasValue: true, Value: "info"},
		IsRead: utils.Optional[bool]{HasValue: true, Value: false},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListNotificationsQuery)).
		WithArgs(false, "info", false, false, 3, 0).
		WillReturnRows(sqlmock.NewRows(notificationColumns()).
			AddRow(1, "info", "one", "m1", sql.NullString{}, false, now, sql.NullString{}, 1, false).
			AddRow(2, "info", "two", "m2", sql.NullString{}, false, now, sql.NullString{}, 1, false).
			AddRow(3, "info", "three", "m3", sql.NullString{}, false, now, sql.NullString{}, 1, false))
	mock.ExpectRollback()

	result, err := repo.ListNotifications(filter, 1, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Items) != 2 || !result.Pagination.HasNext || result.Pagination.HasPrev {
		t.Fatalf("unexpected pagination response: %+v", result.Pagination)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListNotificationsQuery)).
		WillReturnError(errors.New("list failed"))
	mock.ExpectRollback()

	if _, err := repo.ListNotifications(NotificationFilter{}, 1, 10); err == nil || !regexp.MustCompile(`ListNotifications:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped ListNotifications error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryMutationMethods(t *testing.T) {
	repo, mock, db := newNotificationsRepoWithMock(t)
	now := time.Now()

	mock.ExpectBegin()
	txMarkOne, _ := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(queries.MarkAsReadQuery)).
		WithArgs(3).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.MarkAsRead(txMarkOne, 3); err != nil {
		t.Fatalf("expected no mark-as-read error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txMarkOne.Rollback()

	mock.ExpectBegin()
	txMarkAll, _ := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(queries.MarkAllAsReadQuery)).
		WillReturnResult(sqlmock.NewResult(0, 4))
	if err := repo.MarkAllAsRead(txMarkAll); err != nil {
		t.Fatalf("expected no mark-all error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txMarkAll.Rollback()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetUnreadCountQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(12))
	mock.ExpectRollback()
	count, err := repo.GetUnreadCount()
	if err != nil || count != 12 {
		t.Fatalf("expected unread count 12, got %d (err=%v)", count, err)
	}

	mock.ExpectBegin()
	txFindGroup, _ := db.Begin()
	groupKey := sql.NullString{String: "group_A", Valid: true}
	mock.ExpectQuery(regexp.QuoteMeta(queries.FindActiveGroupQuery)).
		WithArgs("group_A", "info", "60").
		WillReturnRows(sqlmock.NewRows(notificationColumns()).
			AddRow(10, "info", "Grouped", "msg", sql.NullString{}, false, now, groupKey, 2, true))
	active, err := repo.FindActiveGroup(txFindGroup, "group_A", "info", 60)
	if err != nil || active.ID != 10 {
		t.Fatalf("expected active group id 10, got %+v (err=%v)", active, err)
	}
	mock.ExpectRollback()
	_ = txFindGroup.Rollback()

	mock.ExpectBegin()
	txUpdateGroup, _ := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateGroupCountQuery)).
		WithArgs(3, "new message", 10).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.UpdateGroupCount(txUpdateGroup, 10, 3, "new message"); err != nil {
		t.Fatalf("expected no update-group error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txUpdateGroup.Rollback()

	mock.ExpectBegin()
	txDeleteOld, _ := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteOldNotificationsQuery)).
		WillReturnResult(sqlmock.NewResult(0, 5))
	if err := repo.DeleteOldNotifications(txDeleteOld); err != nil {
		t.Fatalf("expected no delete-old error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txDeleteOld.Rollback()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryMutationMethodErrors(t *testing.T) {
	repo, mock, db := newNotificationsRepoWithMock(t)

	mock.ExpectBegin()
	txMarkOne, _ := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(queries.MarkAsReadQuery)).
		WillReturnError(errors.New("mark one failed"))
	if err := repo.MarkAsRead(txMarkOne, 1); err == nil || !regexp.MustCompile(`MarkAsRead:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped mark-as-read error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txMarkOne.Rollback()

	mock.ExpectBegin()
	txMarkAll, _ := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(queries.MarkAllAsReadQuery)).
		WillReturnError(errors.New("mark all failed"))
	if err := repo.MarkAllAsRead(txMarkAll); err == nil || !regexp.MustCompile(`MarkAllAsRead:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped mark-all error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txMarkAll.Rollback()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetUnreadCountQuery)).
		WillReturnError(errors.New("count failed"))
	mock.ExpectRollback()
	if _, err := repo.GetUnreadCount(); err == nil || !regexp.MustCompile(`GetUnreadCount:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped unread-count error, got %v", err)
	}

	mock.ExpectBegin()
	txFindGroup, _ := db.Begin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.FindActiveGroupQuery)).
		WillReturnError(errors.New("find group failed"))
	if _, err := repo.FindActiveGroup(txFindGroup, "group", "info", 60); err == nil {
		t.Fatalf("expected find active group error")
	}
	mock.ExpectRollback()
	_ = txFindGroup.Rollback()

	mock.ExpectBegin()
	txUpdateGroup, _ := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateGroupCountQuery)).
		WillReturnError(errors.New("update group failed"))
	if err := repo.UpdateGroupCount(txUpdateGroup, 1, 2, "msg"); err == nil || !regexp.MustCompile(`UpdateGroupCount:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped update-group error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txUpdateGroup.Rollback()

	mock.ExpectBegin()
	txDeleteOld, _ := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteOldNotificationsQuery)).
		WillReturnError(errors.New("delete old failed"))
	if err := repo.DeleteOldNotifications(txDeleteOld); err == nil || !regexp.MustCompile(`DeleteOldNotifications:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped delete-old error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txDeleteOld.Rollback()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
