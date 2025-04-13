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

func TestGetFiles(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := files.NewRepository(db)

	filter := files.FileFilter{Path: "/test/path"}
	pagination := utils.Pagination{Page: 1, PageSize: 10}

	// Ajuste a consulta esperada para corresponder à consulta real
	expectedQuery := regexp.QuoteMeta(queries.GetFilesQuery)

	rows := sqlmock.NewRows([]string{"id", "name", "path", "format", "size", "updated_at", "created_at", "last_interaction", "last_backup", "type", "checksum", "deleted_at"}).
		AddRow(1, "test_file.txt", "/test/path", "txt", 1024, time.Now(), time.Now(), time.Now(), time.Now(), files.File, "checksum123", time.Time{}).
		AddRow(2, "another_file.txt", "/test/path", "txt", 2048, time.Now(), time.Now(), time.Now(), time.Now(), files.File, "checksum456", time.Time{})

	mock.ExpectQuery(expectedQuery).
		WithArgs(filter.Path, pagination.PageSize+1, pagination.Page).
		WillReturnRows(rows)

	paginationResponse, err := repo.GetFiles(filter, pagination)

	assert.NoError(t, err)
	assert.NotNil(t, paginationResponse)
	assert.Len(t, paginationResponse.Items, 2)
	assert.Equal(t, 1, paginationResponse.Pagination.Page)
	assert.Equal(t, 10, paginationResponse.Pagination.PageSize)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetFiles_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := files.NewRepository(db)

	filter := files.FileFilter{Path: "/test/path"}
	pagination := utils.Pagination{Page: 1, PageSize: 10}

	// Ajuste a consulta esperada para corresponder à consulta real
	expectedQuery := regexp.QuoteMeta(queries.GetFilesQuery)

	mock.ExpectQuery(expectedQuery).
		WithArgs(filter.Path, pagination.PageSize+1, pagination.Page).
		WillReturnError(fmt.Errorf("database error"))

	paginationResponse, err := repo.GetFiles(filter, pagination)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.NotNil(t, paginationResponse)
	assert.Len(t, paginationResponse.Items, 0)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestGetFilesByPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := files.NewRepository(db)

	path := "/test/path"

	expectedQuery := regexp.QuoteMeta(queries.GetFilesByPathQuery)

	rows := sqlmock.NewRows([]string{"id", "name", "path", "format", "size", "updated_at", "created_at", "last_interaction", "last_backup"}).
		AddRow(1, "test_file.txt", "/test/path", "txt", 1024, time.Now(), time.Now(), time.Now(), time.Now()).
		AddRow(2, "another_file.txt", "/test/path", "txt", 2048, time.Now(), time.Now(), time.Now(), time.Now())

	mock.ExpectQuery(expectedQuery).
		WithArgs(path).
		WillReturnRows(rows)

	fileModels, err := repo.GetFilesByPath(path)

	assert.NoError(t, err)
	assert.NotNil(t, fileModels)
	assert.Len(t, fileModels, 2)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetFilesByPath_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := files.NewRepository(db)

	path := "/test/path"

	expectedQuery := regexp.QuoteMeta(queries.GetFilesByPathQuery)

	mock.ExpectQuery(expectedQuery).
		WithArgs(path).
		WillReturnError(fmt.Errorf("database error"))

	fileModels, err := repo.GetFilesByPath(path)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Nil(t, fileModels)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestGetFileByNameAndPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := files.NewRepository(db)

	name := "test_file.txt"
	path := "/test/path"

	expectedQuery := regexp.QuoteMeta(queries.GetFileByNameAndPathQuery)

	rows := sqlmock.NewRows([]string{"id", "name", "path", "format", "size", "updated_at", "created_at", "last_interaction", "last_backup"}).
		AddRow(1, "test_file.txt", "/test/path", "txt", 1024, time.Now(), time.Now(), time.Now(), time.Now())

	mock.ExpectQuery(expectedQuery).
		WithArgs(name, path).
		WillReturnRows(rows)

	fileModel, err := repo.GetFileByNameAndPath(name, path)

	assert.NoError(t, err)
	assert.NotNil(t, fileModel)
	assert.Equal(t, "test_file.txt", fileModel.Name)
	assert.Equal(t, "/test/path", fileModel.Path)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetFileByNameAndPath_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := files.NewRepository(db)

	name := "test_file.txt"
	path := "/test/path"

	expectedQuery := regexp.QuoteMeta(queries.GetFileByNameAndPathQuery)

	mock.ExpectQuery(expectedQuery).
		WithArgs(name, path).
		WillReturnError(fmt.Errorf("database error"))

	fileModel, err := repo.GetFileByNameAndPath(name, path)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Equal(t, files.FileModel{}, fileModel)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestCreateFile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := files.NewRepository(db)

	file := files.FileModel{
		Name:            "test_file.txt",
		Path:            "/test/path",
		Format:          "txt",
		Size:            1024,
		UpdatedAt:       time.Now(),
		CreatedAt:       time.Now(),
		LastInteraction: time.Now(),
		LastBackup:      time.Now(),
		Type:            files.File,
		CheckSum:        "checksum123",
		DeletedAt:       time.Time{},
	}

	expectedQuery := regexp.QuoteMeta(queries.InsertFileQuery)

	mock.ExpectBegin()
	mock.ExpectExec(expectedQuery).
		WithArgs(file.Name, file.Path, file.Format, file.Size, file.UpdatedAt, file.CreatedAt, file.LastInteraction, file.LastBackup, file.DeletedAt, file.Type, file.CheckSum).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	assert.NoError(t, err)

	createdFile, err := repo.CreateFile(tx, file)
	tx.Commit()
	assert.NoError(t, err)
	assert.NotNil(t, createdFile)
	assert.Equal(t, 1, createdFile.ID)
	assert.Equal(t, file.Name, createdFile.Name)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateFile_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := files.NewRepository(db)

	file := files.FileModel{
		Name:            "test_file.txt",
		Path:            "/test/path",
		Format:          "txt",
		Size:            1024,
		UpdatedAt:       time.Now(),
		CreatedAt:       time.Now(),
		LastInteraction: time.Now(),
		LastBackup:      time.Now(),
		Type:            files.File,
		CheckSum:        "checksum123",
		DeletedAt:       time.Time{},
	}

	expectedQuery := regexp.QuoteMeta(queries.InsertFileQuery)

	mock.ExpectBegin()
	mock.ExpectExec(expectedQuery).
		WithArgs(file.Name, file.Path, file.Format, file.Size, file.UpdatedAt, file.CreatedAt, file.LastInteraction, file.LastBackup, file.DeletedAt, file.Type, file.CheckSum).
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	tx, err := db.Begin()
	assert.NoError(t, err)

	createdFile, err := repo.CreateFile(tx, file)
	tx.Rollback()
	assert.Error(t, err)
	assert.Equal(t, "CreateFile: database error", err.Error())
	assert.Equal(t, file, createdFile)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestUpdateFile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := files.NewRepository(db)

	file := files.FileModel{
		ID:              1,
		Name:            "test_file.txt",
		Path:            "/test/path",
		Format:          "txt",
		Size:            1024,
		UpdatedAt:       time.Now(),
		CreatedAt:       time.Now(),
		LastInteraction: time.Now(),
		LastBackup:      time.Now(),
		Type:            files.File,
		CheckSum:        "checksum123",
		DeletedAt:       time.Time{},
	}

	expectedQuery := regexp.QuoteMeta(queries.UpdateFileQuery)

	mock.ExpectBegin()
	mock.ExpectExec(expectedQuery).
		WithArgs(file.ID, file.Name, file.Path, file.Format, file.Size, file.UpdatedAt, file.CreatedAt, file.LastInteraction, file.LastBackup, file.Type, file.CheckSum, file.DeletedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	assert.NoError(t, err)

	updated, err := repo.UpdateFile(tx, file)
	tx.Commit()

	assert.NoError(t, err)
	assert.True(t, updated)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateFile_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := files.NewRepository(db)

	file := files.FileModel{
		ID:              1,
		Name:            "test_file.txt",
		Path:            "/test/path",
		Format:          "txt",
		Size:            1024,
		UpdatedAt:       time.Now(),
		CreatedAt:       time.Now(),
		LastInteraction: time.Now(),
		LastBackup:      time.Now(),
		Type:            files.File,
		CheckSum:        "checksum123",
		DeletedAt:       time.Time{},
	}

	expectedQuery := regexp.QuoteMeta(queries.UpdateFileQuery)

	mock.ExpectBegin()
	mock.ExpectExec(expectedQuery).
		WithArgs(file.ID, file.Name, file.Path, file.Format, file.Size, file.UpdatedAt, file.CreatedAt, file.LastInteraction, file.LastBackup, file.Type, file.CheckSum, file.DeletedAt).
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	tx, err := db.Begin()
	assert.NoError(t, err)

	updated, err := repo.UpdateFile(tx, file)
	tx.Rollback()

	assert.Error(t, err)
	assert.Equal(t, "UpdateFile: database error", err.Error())
	assert.False(t, updated)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateFile_MultipleRowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := files.NewRepository(db)

	file := files.FileModel{
		ID:              1,
		Name:            "test_file.txt",
		Path:            "/test/path",
		Format:          "txt",
		Size:            1024,
		UpdatedAt:       time.Now(),
		CreatedAt:       time.Now(),
		LastInteraction: time.Now(),
		LastBackup:      time.Now(),
		Type:            files.File,
		CheckSum:        "checksum123",
		DeletedAt:       time.Time{},
	}

	expectedQuery := regexp.QuoteMeta(queries.UpdateFileQuery)

	mock.ExpectBegin()
	mock.ExpectExec(expectedQuery).
		WithArgs(file.ID, file.Name, file.Path, file.Format, file.Size, file.UpdatedAt, file.CreatedAt, file.LastInteraction, file.LastBackup, file.Type, file.CheckSum, file.DeletedAt).
		WillReturnResult(sqlmock.NewResult(2, 2))
	mock.ExpectRollback()

	tx, err := db.Begin()
	assert.NoError(t, err)

	updated, err := repo.UpdateFile(tx, file)
	tx.Rollback()

	assert.Error(t, err)
	assert.Equal(t, "UpdateFile: multiple rows affected", err.Error())
	assert.False(t, updated)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestGetPathByFileId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := files.NewRepository(db)

	fileId := 1
	expectedPath := "/test/path"

	expectedQuery := regexp.QuoteMeta(queries.GetPathByFileIdQuery)

	rows := sqlmock.NewRows([]string{"path"}).
		AddRow(expectedPath)

	mock.ExpectQuery(expectedQuery).
		WithArgs(fileId).
		WillReturnRows(rows)

	path, err := repo.GetPathByFileId(fileId)

	assert.NoError(t, err)
	assert.Equal(t, expectedPath, path)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetPathByFileId_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := files.NewRepository(db)

	fileId := 1

	expectedQuery := regexp.QuoteMeta(queries.GetPathByFileIdQuery)

	mock.ExpectQuery(expectedQuery).
		WithArgs(fileId).
		WillReturnError(fmt.Errorf("database error"))

	path, err := repo.GetPathByFileId(fileId)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Empty(t, path)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestRepository_GetFilesByPath(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		path        string
		expected    []files.FileModel
		expectedErr error
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
			path: "/test/path",
			expected: []files.FileModel{
				{ID: 1, Name: "test_file.txt", Path: "/test/path"},
				{ID: 2, Name: "another_file.txt", Path: "/test/path"},
			},
			expectedErr: nil,
		},
		{
			name: "GetFilesByPath database error",
			setupMock: func(mock sqlmock.Sqlmock) {
				expectedQuery := regexp.QuoteMeta(queries.GetFilesByPathQuery)
				mock.ExpectQuery(expectedQuery).
					WithArgs("/test/path").
					WillReturnError(fmt.Errorf("database error"))
			},
			path:        "/test/path",
			expected:    nil,
			expectedErr: fmt.Errorf("database error"),
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
				assert.Equal(t, tt.expected, files)
			}

			if err := mockDB.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
