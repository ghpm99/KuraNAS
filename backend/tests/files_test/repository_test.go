package files_test

import (
	"database/sql"
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/database/queries"
	"nas-go/api/pkg/utils"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

type mockDB struct {
	db   *sql.DB
	mock sqlmock.Sqlmock
}

func setupMockDB(t *testing.T) *mockDB {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return &mockDB{db: db, mock: mock}
}

func getTime() time.Time {
	time, err := time.Parse(time.RFC3339, "2023-10-01T00:00:00Z")
	if err != nil {
		panic(fmt.Sprintf("failed to parse time: %v", err))
	}
	return time
}

func TestRepository_GetFiles(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(mock sqlmock.Sqlmock)
		args          files.FileFilter
		pagination    utils.Pagination
		expectedItems int
		expectedErr   error
	}{
		{
			name: "GetFiles success",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.GetFilesQuery)
				rows := sqlmock.NewRows([]string{"id", "name", "path", "format", "size", "updated_at", "created_at", "last_interaction", "last_backup", "type", "checksum", "deleted_at"}).
					AddRow(1, "test_file.txt", "/test/path", "txt", 1024, time.Now(), time.Now(), time.Now(), time.Now(), files.File, "checksum123", time.Time{}).
					AddRow(2, "another_file.txt", "/test/path", "txt", 2048, time.Now(), time.Now(), time.Now(), time.Now(), files.File, "checksum456", time.Time{})
				mock.ExpectQuery(expectedQuery).
					WithArgs("/test/path", 11, 1).
					WillReturnRows(rows)
			},
			args: files.FileFilter{
				Path: "/test/path",
			},
			pagination:    utils.Pagination{Page: 1, PageSize: 10},
			expectedItems: 2,
			expectedErr:   nil,
		},
		{
			name: "GetFiles database error",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.GetFilesQuery)
				mock.ExpectQuery(expectedQuery).
					WithArgs("/test/path", 11, 1).
					WillReturnError(fmt.Errorf("database error"))
			},
			args: files.FileFilter{
				Path: "/test/path",
			},
			pagination:    utils.Pagination{Page: 1, PageSize: 10},
			expectedItems: 0,
			expectedErr:   fmt.Errorf("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := setupMockDB(t)
			defer mockDB.db.Close()

			repo := files.NewRepository(mockDB.db)
			tt.setupMock(mockDB.mock)

			paginationResponse, err := repo.GetFiles(tt.args, tt.pagination)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, paginationResponse.Items, tt.expectedItems)
			}

			if err := mockDB.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestRepository_GetFilesByPath(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(mock sqlmock.Sqlmock)
		path          string
		expectedItems int
		expectedErr   error
	}{
		{
			name: "GetFilesByPath success",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.GetFilesByPathQuery)

				rows := sqlmock.NewRows([]string{"id", "name", "path", "format", "size", "updated_at", "created_at", "last_interaction", "last_backup"}).
					AddRow(1, "test_file.txt", "/test/path", "txt", 1024, time.Now(), time.Now(), time.Now(), time.Now()).
					AddRow(2, "another_file.txt", "/test/path", "txt", 2048, time.Now(), time.Now(), time.Now(), time.Now())

				mock.ExpectQuery(expectedQuery).
					WithArgs("/test/path").
					WillReturnRows(rows)
			},
			path:          "/test/path",
			expectedItems: 2,
			expectedErr:   nil,
		},
		{
			name: "GetFilesByPath database error",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.GetFilesByPathQuery)
				mock.ExpectQuery(expectedQuery).
					WithArgs("/test/path").
					WillReturnError(fmt.Errorf("database error"))
			},
			path:          "/test/path",
			expectedItems: 0,
			expectedErr:   fmt.Errorf("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := setupMockDB(t)
			defer mockDB.db.Close()

			repo := files.NewRepository(mockDB.db)
			tt.setupMock(mockDB.mock)

			files, err := repo.GetFilesByPath(tt.path)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, files, tt.expectedItems)
			}

			if err := mockDB.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestRepository_GetFileByNameAndPath(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		fileName    string
		path        string
		expected    files.FileModel
		expectedErr error
	}{
		{
			name: "GetFileByNameAndPath success",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.GetFileByNameAndPathQuery)
				rows := sqlmock.NewRows(
					[]string{
						"id",
						"name",
						"path",
						"format",
						"size",
						"updated_at",
						"created_at",
						"last_interaction",
						"last_backup",
					}).
					AddRow(
						1,
						"test_file.txt",
						"/test/path",
						".txt",
						1024,
						getTime(),
						getTime(),
						getTime(),
						getTime(),
					)
				mock.ExpectQuery(expectedQuery).
					WithArgs("test_file.txt", "/test/path").
					WillReturnRows(rows)
			},
			fileName: "test_file.txt",
			path:     "/test/path",
			expected: files.FileModel{
				ID:        1,
				Name:      "test_file.txt",
				Path:      "/test/path",
				Format:    ".txt",
				Size:      1024,
				UpdatedAt: getTime(),
				CreatedAt: getTime(),
				LastInteraction: sql.NullTime{
					Time:  getTime(),
					Valid: true,
				},
				LastBackup: sql.NullTime{
					Time:  getTime(),
					Valid: true,
				},
			},
			expectedErr: nil,
		},
		{
			name: "GetFileByNameAndPath database error",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.GetFileByNameAndPathQuery)
				mock.ExpectQuery(expectedQuery).
					WithArgs("test_file.txt", "/test/path").
					WillReturnError(fmt.Errorf("database error"))
			},
			fileName:    "test_file.txt",
			path:        "/test/path",
			expected:    files.FileModel{},
			expectedErr: fmt.Errorf("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := setupMockDB(t)
			defer mockDB.db.Close()

			repo := files.NewRepository(mockDB.db)
			tt.setupMock(mockDB.mock)

			file, err := repo.GetFileByNameAndPath(tt.fileName, tt.path)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, file)
			}

			if err := mockDB.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestRepository_CreateFile(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		file        files.FileModel
		expectedErr error
	}{
		{
			name: "CreateFile success",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.InsertFileQuery)
				mock.ExpectBegin()
				mock.ExpectExec(expectedQuery).
					WithArgs(
						"test_file.txt",
						"/test/path",
						"txt",
						1024,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						files.File,
						"checksum123",
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			file: files.FileModel{
				Name:     "test_file.txt",
				Path:     "/test/path",
				Format:   "txt",
				Size:     1024,
				Type:     files.File,
				CheckSum: "checksum123",
			},
			expectedErr: nil,
		},
		{
			name: "CreateFile database error",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.InsertFileQuery)
				mock.ExpectBegin()
				mock.ExpectExec(expectedQuery).
					WithArgs(
						"test_file.txt",
						"/test/path",
						"txt",
						1024,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						files.File,
						"checksum123",
					).
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			file: files.FileModel{
				Name:     "test_file.txt",
				Path:     "/test/path",
				Format:   "txt",
				Size:     1024,
				Type:     files.File,
				CheckSum: "checksum123",
			},
			expectedErr: fmt.Errorf("CreateFile: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := setupMockDB(t)
			defer mockDB.db.Close()

			repo := files.NewRepository(mockDB.db)
			tt.setupMock(mockDB.mock)

			tx, err := mockDB.db.Begin()
			assert.NoError(t, err)

			_, err = repo.CreateFile(tx, tt.file)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				tx.Rollback()
			} else {
				assert.NoError(t, err)
				tx.Commit()
			}

			if err := mockDB.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestRepository_UpdateFile(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		file        files.FileModel
		expected    bool
		expectedErr error
	}{
		{
			name: "UpdateFile success",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.UpdateFileQuery)
				mock.ExpectBegin()
				mock.ExpectExec(expectedQuery).
					WithArgs(
						1,
						"test_file.txt",
						"/test/path",
						"txt",
						1024,
						getTime(),
						getTime(),
						getTime(),
						getTime(),
						files.File,
						"checksum123",
						getTime(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			file: files.FileModel{
				ID:        1,
				Name:      "test_file.txt",
				Path:      "/test/path",
				Format:    "txt",
				Size:      1024,
				Type:      files.File,
				CheckSum:  "checksum123",
				UpdatedAt: getTime(),
				CreatedAt: getTime(),
				LastInteraction: sql.NullTime{
					Time:  getTime(),
					Valid: true,
				},
				LastBackup: sql.NullTime{
					Time:  getTime(),
					Valid: true,
				},
				DeletedAt: sql.NullTime{
					Time:  getTime(),
					Valid: true,
				},
			},
			expected:    true,
			expectedErr: nil,
		},
		{
			name: "UpdateFile success nil time",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.UpdateFileQuery)
				mock.ExpectBegin()
				mock.ExpectExec(expectedQuery).
					WithArgs(
						1,
						"test_file.txt",
						"/test/path",
						"txt",
						1024,
						getTime(),
						getTime(),
						nil,
						nil,
						files.File,
						"checksum123",
						nil,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			file: files.FileModel{
				ID:        1,
				Name:      "test_file.txt",
				Path:      "/test/path",
				Format:    "txt",
				Size:      1024,
				Type:      files.File,
				CheckSum:  "checksum123",
				UpdatedAt: getTime(),
				CreatedAt: getTime(),
				LastInteraction: sql.NullTime{
					Valid: false,
				},
				LastBackup: sql.NullTime{
					Valid: false,
				},
				DeletedAt: sql.NullTime{
					Valid: false,
				},
			},
			expected:    true,
			expectedErr: nil,
		},
		{
			name: "UpdateFile database error",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.UpdateFileQuery)
				mock.ExpectBegin()
				mock.ExpectExec(expectedQuery).
					WithArgs(
						1,
						"test_file.txt",
						"/test/path",
						"txt",
						1024,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						files.File,
						"checksum123",
						nil,
					).
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
			},
			file: files.FileModel{
				ID:        1,
				Name:      "test_file.txt",
				Path:      "/test/path",
				Format:    "txt",
				Size:      1024,
				Type:      files.File,
				CheckSum:  "checksum123",
				UpdatedAt: time.Now(),
				CreatedAt: time.Now(),
				LastInteraction: sql.NullTime{
					Time:  getTime(),
					Valid: true,
				},
				LastBackup: sql.NullTime{
					Time:  getTime(),
					Valid: true,
				},
				DeletedAt: sql.NullTime{
					Valid: false,
				},
			},
			expected:    false,
			expectedErr: fmt.Errorf("UpdateFile: database error"),
		},
		{
			name: "UpdateFile multiple rows affected error",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.UpdateFileQuery)
				mock.ExpectBegin()
				mock.ExpectExec(expectedQuery).
					WithArgs(
						1,
						"test_file.txt",
						"/test/path",
						"txt",
						1024,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						files.File,
						"checksum123",
						nil,
					).
					WillReturnResult(sqlmock.NewResult(2, 2))
				mock.ExpectRollback()
			},
			file: files.FileModel{
				ID:        1,
				Name:      "test_file.txt",
				Path:      "/test/path",
				Format:    "txt",
				Size:      1024,
				Type:      files.File,
				CheckSum:  "checksum123",
				UpdatedAt: time.Now(),
				CreatedAt: time.Now(),
				LastInteraction: sql.NullTime{
					Time:  getTime(),
					Valid: true,
				},
				LastBackup: sql.NullTime{
					Time:  getTime(),
					Valid: true,
				},
				DeletedAt: sql.NullTime{
					Valid: false,
				},
			},
			expected:    false,
			expectedErr: fmt.Errorf("UpdateFile: multiple rows affected"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := setupMockDB(t)
			defer mockDB.db.Close()

			repo := files.NewRepository(mockDB.db)
			tt.setupMock(mockDB.mock)

			tx, err := mockDB.db.Begin()
			assert.NoError(t, err)

			updated, err := repo.UpdateFile(tx, tt.file)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Equal(t, tt.expected, updated)
				tx.Rollback()
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, updated)
				tx.Commit()
			}

			if err := mockDB.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestRepository_GetPathByFileId(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		fileId      int
		expected    string
		expectedErr error
	}{
		{
			name: "GetPathByFileId success",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.GetPathByFileIdQuery)
				rows := sqlmock.NewRows([]string{"path"}).
					AddRow("/test/path")
				mock.ExpectQuery(expectedQuery).
					WithArgs(1).
					WillReturnRows(rows)
			},
			fileId:      1,
			expected:    "/test/path",
			expectedErr: nil,
		},
		{
			name: "GetPathByFileId database error",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.GetPathByFileIdQuery)
				mock.ExpectQuery(expectedQuery).
					WithArgs(1).
					WillReturnError(fmt.Errorf("database error"))
			},
			fileId:      1,
			expected:    "",
			expectedErr: fmt.Errorf("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := setupMockDB(t)
			defer mockDB.db.Close()

			repo := files.NewRepository(mockDB.db)
			tt.setupMock(mockDB.mock)

			path, err := repo.GetPathByFileId(tt.fileId)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, path)
			}

			if err := mockDB.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
