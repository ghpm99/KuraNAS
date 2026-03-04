package diary

import (
	"database/sql"
	"errors"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type repoMock struct {
	getDiaryFn    func(filter DiaryFilter, page int, pageSize int) (utils.PaginationResponse[DiaryModel], error)
	createDiaryFn func(transaction *sql.Tx, diary DiaryModel) (DiaryModel, error)
	updateDiaryFn func(transaction *sql.Tx, diary DiaryModel) (bool, error)
	summaryFn     func(dateReference time.Time) (DiarySummary, error)
	dbContext     *database.DbContext
}

func (r *repoMock) GetDbContext() *database.DbContext { return r.dbContext }
func (r *repoMock) CreateDiary(transaction *sql.Tx, diary DiaryModel) (DiaryModel, error) {
	if r.createDiaryFn != nil {
		return r.createDiaryFn(transaction, diary)
	}
	return diary, nil
}
func (r *repoMock) GetDiary(filter DiaryFilter, page int, pageSize int) (utils.PaginationResponse[DiaryModel], error) {
	if r.getDiaryFn != nil {
		return r.getDiaryFn(filter, page, pageSize)
	}
	return utils.PaginationResponse[DiaryModel]{Items: []DiaryModel{}}, nil
}
func (r *repoMock) UpdateDiary(transaction *sql.Tx, diary DiaryModel) (bool, error) {
	if r.updateDiaryFn != nil {
		return r.updateDiaryFn(transaction, diary)
	}
	return true, nil
}
func (r *repoMock) GetSummary(dateReference time.Time) (DiarySummary, error) {
	if r.summaryFn != nil {
		return r.summaryFn(dateReference)
	}
	return DiarySummary{}, nil
}

func newServiceForTest(t *testing.T, mock *repoMock) *Service {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	mock.dbContext = database.NewDbContext(db)
	return &Service{Repository: mock}
}

func TestServiceCreateDiary(t *testing.T) {
	start := time.Now().Add(-1 * time.Hour)
	mock := &repoMock{
		getDiaryFn: func(filter DiaryFilter, page int, pageSize int) (utils.PaginationResponse[DiaryModel], error) {
			return utils.PaginationResponse[DiaryModel]{
				Items: []DiaryModel{{ID: 1, Name: "previous", StartTime: start}},
			}, nil
		},
		createDiaryFn: func(transaction *sql.Tx, diary DiaryModel) (DiaryModel, error) {
			diary.ID = 2
			return diary, nil
		},
		updateDiaryFn: func(transaction *sql.Tx, diary DiaryModel) (bool, error) {
			return true, nil
		},
	}
	service := newServiceForTest(t, mock)

	result, err := service.CreateDiary(DiaryDto{Name: "new", Description: "desc"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Name != "new" {
		t.Fatalf("expected created diary name new, got %s", result.Name)
	}
}

func TestServiceGetDiaryAndUpdate(t *testing.T) {
	mock := &repoMock{
		getDiaryFn: func(filter DiaryFilter, page int, pageSize int) (utils.PaginationResponse[DiaryModel], error) {
			return utils.PaginationResponse[DiaryModel]{
				Items: []DiaryModel{{ID: 1, Name: "a", Description: "b", StartTime: time.Now()}},
				Pagination: utils.Pagination{
					Page: 1, PageSize: 10,
				},
			}, nil
		},
		updateDiaryFn: func(transaction *sql.Tx, diary DiaryModel) (bool, error) {
			return diary.ID == 1, nil
		},
	}
	service := newServiceForTest(t, mock)

	diaries, err := service.GetDiary(DiaryFilter{}, 1, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(diaries.Items) != 1 {
		t.Fatalf("expected one diary item")
	}

	ok, err := service.UpdateDiary(DiaryDto{
		ID:          1,
		Name:        "a",
		Description: "b",
		StartTime:   time.Now(),
		EndTime:     utils.Optional[time.Time]{HasValue: false},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !ok {
		t.Fatalf("expected update success")
	}
}

func TestServiceGetSummaryAndDuplicate(t *testing.T) {
	now := time.Now()
	mock := &repoMock{
		getDiaryFn: func(filter DiaryFilter, page int, pageSize int) (utils.PaginationResponse[DiaryModel], error) {
			// Duplicate path asks id filter with one item.
			if filter.ID.HasValue {
				return utils.PaginationResponse[DiaryModel]{
					Items: []DiaryModel{
						{ID: filter.ID.Value, Name: "copy-me", Description: "d", StartTime: now.Add(-30 * time.Minute)},
					},
				}, nil
			}

			// Summary path asks date range and expects at least one entry.
			return utils.PaginationResponse[DiaryModel]{
				Items: []DiaryModel{
					{
						ID:          1,
						Name:        "activity",
						Description: "desc",
						StartTime:   now.Add(-1 * time.Hour),
						EndTime: sql.NullTime{
							Valid: true,
							Time:  now,
						},
					},
				},
			}, nil
		},
		createDiaryFn: func(transaction *sql.Tx, diary DiaryModel) (DiaryModel, error) {
			diary.ID = 999
			return diary, nil
		},
	}
	service := newServiceForTest(t, mock)

	summary, err := service.GetSummary()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if summary.TotalActivities < 1 {
		t.Fatalf("expected at least one activity in summary")
	}

	duplicated, err := service.DuplicateDiary(7)
	if err != nil {
		t.Fatalf("expected duplicate to succeed, got %v", err)
	}
	if duplicated.Name != "copy-me" {
		t.Fatalf("expected duplicated name copy-me, got %s", duplicated.Name)
	}
}

func TestServiceErrorPaths(t *testing.T) {
	mock := &repoMock{
		getDiaryFn: func(filter DiaryFilter, page int, pageSize int) (utils.PaginationResponse[DiaryModel], error) {
			return utils.PaginationResponse[DiaryModel]{}, errors.New("repo error")
		},
	}
	service := newServiceForTest(t, mock)

	if _, err := service.GetDiary(DiaryFilter{}, 1, 10); err == nil {
		t.Fatalf("expected get diary error")
	}
	if _, err := service.GetSummary(); err == nil {
		t.Fatalf("expected summary error")
	}
}

func TestCalculateDailyDurationAndLongestActivity(t *testing.T) {
	now := time.Now()
	entries := []DiaryDto{
		{
			Name:      "closed",
			StartTime: now.Add(-2 * time.Hour),
			EndTime:   utils.Optional[time.Time]{HasValue: true, Value: now.Add(-1 * time.Hour)},
		},
		{
			Name:      "open",
			StartTime: now.Add(-30 * time.Minute),
			EndTime:   utils.Optional[time.Time]{HasValue: false},
		},
	}

	duration, err := calculateDailyDuration(entries)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if duration <= 0 {
		t.Fatalf("expected positive total duration")
	}

	longest, err := getLongestActivity(entries)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if longest.Name == "" {
		t.Fatalf("expected longest activity name")
	}
}

func TestWithTransactionUsesDbContext(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer db.Close()

	s := &Service{
		Repository: &repoMock{
			dbContext: database.NewDbContext(db),
		},
	}

	called := false
	err = s.withTransaction(func(tx *sql.Tx) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatalf("expected transaction callback to be called")
	}
}

func TestRepoMockCompileOnly(t *testing.T) {
	// Ensure repoMock satisfies interface at compile time.
	var _ RepositoryInterface = (*repoMock)(nil)
	_ = errors.New("")
}

func TestNewService(t *testing.T) {
	repo := &repoMock{}
	svc := NewService(repo, make(chan utils.Task, 1))
	typed, ok := svc.(*Service)
	if !ok || typed.Repository != repo {
		t.Fatalf("expected concrete service with repository")
	}
}
