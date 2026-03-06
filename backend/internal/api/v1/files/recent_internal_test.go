package files

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/file"
	"nas-go/api/pkg/utils"

	"github.com/DATA-DOG/go-sqlmock"
)

type recentRepoMock struct {
	upsertFn      func(ip string, fileID int) error
	deleteOldFn   func(ip string, keep int) error
	getRecentFn   func(page int, pageSize int) ([]RecentFileModel, error)
	deleteFn      func(ip string, fileID int) error
	getByFileIDFn func(fileID int) ([]RecentFileModel, error)
}

func (m *recentRepoMock) Upsert(ip string, fileID int) error  { return m.upsertFn(ip, fileID) }
func (m *recentRepoMock) DeleteOld(ip string, keep int) error { return m.deleteOldFn(ip, keep) }
func (m *recentRepoMock) GetRecentFiles(page int, pageSize int) ([]RecentFileModel, error) {
	return m.getRecentFn(page, pageSize)
}
func (m *recentRepoMock) Delete(ip string, fileID int) error { return m.deleteFn(ip, fileID) }
func (m *recentRepoMock) GetByFileID(fileID int) ([]RecentFileModel, error) {
	return m.getByFileIDFn(fileID)
}

func TestRecentFileService(t *testing.T) {
	now := time.Now()
	repo := &recentRepoMock{
		upsertFn:    func(ip string, fileID int) error { return nil },
		deleteOldFn: func(ip string, keep int) error { return nil },
		getRecentFn: func(page int, pageSize int) ([]RecentFileModel, error) {
			return []RecentFileModel{{ID: 1, IPAddress: "127.0.0.1", FileID: 2, AccessedAt: now}}, nil
		},
		deleteFn: func(ip string, fileID int) error { return nil },
		getByFileIDFn: func(fileID int) ([]RecentFileModel, error) {
			return []RecentFileModel{{ID: 2, IPAddress: "127.0.0.1", FileID: fileID, AccessedAt: now}}, nil
		},
	}
	svc := NewRecentFileService(repo)

	if err := svc.RegisterAccess("127.0.0.1", 2, 10); err != nil {
		t.Fatalf("RegisterAccess failed: %v", err)
	}
	list, err := svc.GetRecentFiles(0, 0)
	if err != nil || len(list) != 1 {
		t.Fatalf("GetRecentFiles failed len=%d err=%v", len(list), err)
	}
	if err := svc.DeleteRecentFile("127.0.0.1", 2); err != nil {
		t.Fatalf("DeleteRecentFile failed: %v", err)
	}
	byID, err := svc.GetRecentAccessByFileID(2)
	if err != nil || len(byID) != 1 {
		t.Fatalf("GetRecentAccessByFileID failed len=%d err=%v", len(byID), err)
	}
}

func TestRecentFileRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRecentFileRepository(database.NewDbContext(db))
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpsertRecentFileQuery)).
		WithArgs("127.0.0.1", 10).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := repo.Upsert("127.0.0.1", 10); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteOldRecentFilesQuery)).
		WithArgs("127.0.0.1", "127.0.0.1", 10).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := repo.DeleteOld("127.0.0.1", 10); err != nil {
		t.Fatalf("DeleteOld failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetRecentFilesQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "ip_address", "file_id", "accessed_at"}).
			AddRow(1, "127.0.0.1", 10, now))
	result, err := repo.GetRecentFiles(1, 10)
	if err != nil || len(result) != 1 {
		t.Fatalf("GetRecentFiles failed len=%d err=%v", len(result), err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteRecentFileQuery)).
		WithArgs("127.0.0.1", 10).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := repo.Delete("127.0.0.1", 10); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetRecentByFileIDQuery)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "ip_address", "file_id", "accessed_at"}).
			AddRow(2, "127.0.0.1", 10, now))
	byID, err := repo.GetByFileID(10)
	if err != nil || len(byID) != 1 {
		t.Fatalf("GetByFileID failed len=%d err=%v", len(byID), err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestFileModelChecksumAndServiceConstructors(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "x.txt")
	if err := os.WriteFile(p, []byte("abc"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	m := FileModel{Path: p}
	if err := m.GetCheckSumFromFile(); err != nil {
		t.Fatalf("GetCheckSumFromFile failed: %v", err)
	}

	svc := NewService(&filesRepoMock{}, &metadataRepoMock{}, make(chan utils.Task, 1), nil)
	if svc == nil {
		t.Fatalf("expected non-nil service")
	}
}

func TestRecentFileService_RegisterAccessErrors(t *testing.T) {
	upsertErrRepo := &recentRepoMock{
		upsertFn:      func(ip string, fileID int) error { return errors.New("upsert failed") },
		deleteOldFn:   func(ip string, keep int) error { return nil },
		getRecentFn:   func(page int, pageSize int) ([]RecentFileModel, error) { return nil, nil },
		deleteFn:      func(ip string, fileID int) error { return nil },
		getByFileIDFn: func(fileID int) ([]RecentFileModel, error) { return nil, nil },
	}
	if err := NewRecentFileService(upsertErrRepo).RegisterAccess("127.0.0.1", 1, 10); err == nil {
		t.Fatalf("expected RegisterAccess to return upsert error")
	}

	deleteOldErrRepo := &recentRepoMock{
		upsertFn:      func(ip string, fileID int) error { return nil },
		deleteOldFn:   func(ip string, keep int) error { return errors.New("delete old failed") },
		getRecentFn:   func(page int, pageSize int) ([]RecentFileModel, error) { return nil, nil },
		deleteFn:      func(ip string, fileID int) error { return nil },
		getByFileIDFn: func(fileID int) ([]RecentFileModel, error) { return nil, nil },
	}
	if err := NewRecentFileService(deleteOldErrRepo).RegisterAccess("127.0.0.1", 1, 10); err == nil {
		t.Fatalf("expected RegisterAccess to return delete-old error")
	}
}
