package diary_test

import (
	"context"
	"database/sql"
	"errors"
	"nas-go/api/internal/api/v1/diary"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db, _ = sql.Open("sqlite3", ":memory:")
var dbContext = database.NewDbContext(db)

type MockRepository struct {
	GetDiaryFunc      func(filter diary.DiaryFilter, page int, pageSize int) (utils.PaginationResponse[diary.DiaryModel], error)
	CreateDiaryFunc   func(tx *sql.Tx, diaryModel diary.DiaryModel) (diary.DiaryModel, error)
	UpdateDiaryFunc   func(tx *sql.Tx, diaryModel diary.DiaryModel) (bool, error)
	GetDbContextFunc  func() *database.DbContext
	BeginTxFunc       func(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	CommitFunc        func() error
	RollbackFunc      func() error
	ExpectedDiaryDto  diary.DiaryDto
	ExpectedDiary     diary.DiaryModel
	ExpectedBool      bool
	ExpectedError     error
	ExpectedDiaries   []diary.DiaryModel
	ExpectedDiaryDtos []diary.DiaryDto
}

func (m *MockRepository) GetDiary(filter diary.DiaryFilter, page int, pageSize int) (utils.PaginationResponse[diary.DiaryModel], error) {
	if m.GetDiaryFunc != nil {
		return m.GetDiaryFunc(filter, page, pageSize)
	}
	return utils.PaginationResponse[diary.DiaryModel]{}, m.ExpectedError
}

func (m *MockRepository) CreateDiary(tx *sql.Tx, diaryModel diary.DiaryModel) (diary.DiaryModel, error) {
	if m.CreateDiaryFunc != nil {
		return m.CreateDiaryFunc(tx, diaryModel)
	}
	return m.ExpectedDiary, m.ExpectedError
}

func (m *MockRepository) UpdateDiary(tx *sql.Tx, diaryModel diary.DiaryModel) (bool, error) {
	if m.UpdateDiaryFunc != nil {
		return m.UpdateDiaryFunc(tx, diaryModel)
	}
	return m.ExpectedBool, m.ExpectedError
}

func (m *MockRepository) GetDbContext() *database.DbContext {
	if m.GetDbContextFunc != nil {
		return m.GetDbContextFunc()
	}
	return dbContext
}

func (m *MockRepository) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx, opts)
	}
	return nil, m.ExpectedError
}

func (m *MockRepository) Commit() error {
	if m.CommitFunc != nil {
		return m.CommitFunc()
	}
	return m.ExpectedError
}

func (m *MockRepository) Rollback() error {
	if m.RollbackFunc != nil {
		return m.RollbackFunc()
	}
	return m.ExpectedError
}

func TestCreateDiary(t *testing.T) {
	testTime := time.Now()
	diaryDto := diary.DiaryDto{
		Name:        "Test Diary",
		Description: "Test Description",
	}

	diaryModel := diary.DiaryModel{
		Name:        diaryDto.Name,
		Description: diaryDto.Description,
		StartTime:   testTime,
	}

	expectedDiaryDto := diary.DiaryDto{
		Name:        "Test Diary",
		Description: "Test Description",
		StartTime:   testTime,
	}

	tests := []struct {
		name                      string
		diaryDto                  diary.DiaryDto
		mockRepository            *MockRepository
		expectedDiaryDto          diary.DiaryDto
		expectedError             error
		currentDiaryExists        bool
		updateDiaryError          error
		createDiaryError          error
		getDiaryError             error
		toModelError              error
		toDtoError                error
		beginTxError              error
		commitError               error
		rollbackError             error
		expectedGetDiaryCallCount int
	}{
		{
			name:     "Successful diary creation with no existing diary",
			diaryDto: diaryDto,
			mockRepository: &MockRepository{
				GetDiaryFunc: func(filter diary.DiaryFilter, page int, pageSize int) (utils.PaginationResponse[diary.DiaryModel], error) {
					return utils.PaginationResponse[diary.DiaryModel]{Items: []diary.DiaryModel{}}, nil
				},
				CreateDiaryFunc: func(tx *sql.Tx, diaryModel diary.DiaryModel) (diary.DiaryModel, error) {
					return diaryModel, nil
				},
				UpdateDiaryFunc: func(tx *sql.Tx, diaryModel diary.DiaryModel) (bool, error) {
					return true, nil
				},
				ExpectedDiary: diaryModel,
			},
			expectedDiaryDto:          expectedDiaryDto,
			expectedError:             nil,
			currentDiaryExists:        false,
			expectedGetDiaryCallCount: 1,
		},
		{
			name:     "Successful diary creation with existing diary",
			diaryDto: diaryDto,
			mockRepository: &MockRepository{
				GetDiaryFunc: func(filter diary.DiaryFilter, page int, pageSize int) (utils.PaginationResponse[diary.DiaryModel], error) {
					return utils.PaginationResponse[diary.DiaryModel]{Items: []diary.DiaryModel{{}}}, nil
				},
				CreateDiaryFunc: func(tx *sql.Tx, diaryModel diary.DiaryModel) (diary.DiaryModel, error) {
					return diaryModel, nil
				},
				UpdateDiaryFunc: func(tx *sql.Tx, diaryModel diary.DiaryModel) (bool, error) {
					return true, nil
				},
				ExpectedDiary: diaryModel,
			},
			expectedDiaryDto:          expectedDiaryDto,
			expectedError:             nil,
			currentDiaryExists:        true,
			expectedGetDiaryCallCount: 1,
		},
		{
			name:     "Error getting current diary",
			diaryDto: diaryDto,
			mockRepository: &MockRepository{
				GetDiaryFunc: func(filter diary.DiaryFilter, page int, pageSize int) (utils.PaginationResponse[diary.DiaryModel], error) {
					return utils.PaginationResponse[diary.DiaryModel]{}, errors.New("get diary error")
				},
				CreateDiaryFunc: func(tx *sql.Tx, diaryModel diary.DiaryModel) (diary.DiaryModel, error) {
					return diaryModel, nil
				},
				UpdateDiaryFunc: func(tx *sql.Tx, diaryModel diary.DiaryModel) (bool, error) {
					return true, nil
				},
				ExpectedDiary: diaryModel,
				ExpectedError: errors.New("get diary error"),
			},
			expectedDiaryDto:          diary.DiaryDto{},
			expectedError:             errors.New("get diary error"),
			currentDiaryExists:        false,
			expectedGetDiaryCallCount: 1,
		},
		{
			name:     "Error updating current diary",
			diaryDto: diaryDto,
			mockRepository: &MockRepository{
				GetDiaryFunc: func(filter diary.DiaryFilter, page int, pageSize int) (utils.PaginationResponse[diary.DiaryModel], error) {
					return utils.PaginationResponse[diary.DiaryModel]{Items: []diary.DiaryModel{{}}}, nil
				},
				CreateDiaryFunc: func(tx *sql.Tx, diaryModel diary.DiaryModel) (diary.DiaryModel, error) {
					return diaryModel, nil
				},
				UpdateDiaryFunc: func(tx *sql.Tx, diaryModel diary.DiaryModel) (bool, error) {
					return false, errors.New("update diary error")
				},
				ExpectedDiary: diaryModel,
				ExpectedError: errors.New("update diary error"),
			},
			expectedDiaryDto:          diary.DiaryDto{},
			expectedError:             errors.New("update diary error"),
			currentDiaryExists:        true,
			expectedGetDiaryCallCount: 1,
		},
		{
			name:     "Error creating diary",
			diaryDto: diaryDto,
			mockRepository: &MockRepository{
				GetDiaryFunc: func(filter diary.DiaryFilter, page int, pageSize int) (utils.PaginationResponse[diary.DiaryModel], error) {
					return utils.PaginationResponse[diary.DiaryModel]{Items: []diary.DiaryModel{}}, nil
				},
				CreateDiaryFunc: func(tx *sql.Tx, diaryModel diary.DiaryModel) (diary.DiaryModel, error) {
					return diary.DiaryModel{}, errors.New("create diary error")
				},
				UpdateDiaryFunc: func(tx *sql.Tx, diaryModel diary.DiaryModel) (bool, error) {
					return true, nil
				},
				ExpectedDiary: diaryModel,
				ExpectedError: errors.New("create diary error"),
			},
			expectedDiaryDto:          diary.DiaryDto{},
			expectedError:             errors.New("create diary error"),
			currentDiaryExists:        false,
			expectedGetDiaryCallCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := make(chan utils.Task, 1)
			service := &diary.Service{Repository: tt.mockRepository, Tasks: tasks}

			diaryDtoResult, err := service.CreateDiary(tt.diaryDto)

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("CreateDiary() error = %v, wantErr %v", err, tt.expectedError)
					return
				}
			} else {
				if err != nil {
					t.Errorf("CreateDiary() error = %v, wantErr %v", err, tt.expectedError)
					return
				}
			}

			if tt.expectedError == nil {
				if diaryDtoResult.Name != tt.expectedDiaryDto.Name || diaryDtoResult.Description != tt.expectedDiaryDto.Description {
					t.Errorf("CreateDiary() diaryDtoResult = %v, want %v", diaryDtoResult, tt.expectedDiaryDto)
				}
			}

		})
	}
}
