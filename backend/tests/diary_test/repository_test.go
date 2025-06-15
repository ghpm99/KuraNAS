package diary_test

import (
	"database/sql"
	"errors"
	"nas-go/api/internal/api/v1/diary"
	"nas-go/api/pkg/utils"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepository_GetDiary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &diary.Repository{DbContext: db}

	type args struct {
		filter   diary.DiaryFilter
		page     int
		pageSize int
	}
	tests := []struct {
		name    string
		repo    *diary.Repository
		args    args
		want    utils.PaginationResponse[diary.DiaryModel]
		wantErr bool
		mock    func(args)
	}{
		{
			name: "success",
			repo: repo,
			args: args{
				filter: diary.DiaryFilter{
					ID: utils.Optional[int]{
						HasValue: false,
						Value:    0,
					},
					Name: utils.Optional[string]{
						HasValue: false,
						Value:    "",
					},
					Description: utils.Optional[string]{
						HasValue: false,
						Value:    "",
					},
					StartTime: utils.Optional[time.Time]{
						HasValue: false,
						Value:    time.Time{},
					},
					EndTime: utils.Optional[time.Time]{
						HasValue: false,
						Value:    time.Time{},
					},
					DateRange: utils.Optional[diary.DateRange]{
						HasValue: false,
						Value: diary.DateRange{
							Start: time.Time{},
							End:   time.Time{},
						},
					},
				},
				page:     1,
				pageSize: 10,
			},
			want: utils.PaginationResponse[diary.DiaryModel]{
				Items: []diary.DiaryModel{},
				Pagination: utils.Pagination{
					Page:     1,
					PageSize: 10,
					HasNext:  false,
					HasPrev:  false,
				},
			},
			wantErr: false,
			mock: func(args args) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "start_time", "end_time"})
				mock.ExpectQuery(`SELECT id, name, description, start_time, end_time FROM activity_diary.*ORDER BY.*LIMIT.*OFFSET`).
					WithArgs(
						true, args.filter.ID.Value,
						true, args.filter.Name.Value,
						true, args.filter.Description.Value,
						true, args.filter.StartTime.Value,
						true, args.filter.EndTime.Value,
						true, args.filter.DateRange.Value.Start, args.filter.DateRange.Value.End,
						args.pageSize+1, utils.CalculateOffset(args.page, args.pageSize),
					).
					WillReturnRows(rows)
			},
		},
		{
			name: "error",
			repo: repo,
			args: args{
				filter: diary.DiaryFilter{
					ID: utils.Optional[int]{
						HasValue: false,
						Value:    0,
					},
					Name: utils.Optional[string]{
						HasValue: false,
						Value:    "",
					},
					Description: utils.Optional[string]{
						HasValue: false,
						Value:    "",
					},
					StartTime: utils.Optional[time.Time]{
						HasValue: false,
						Value:    time.Time{},
					},
					EndTime: utils.Optional[time.Time]{
						HasValue: false,
						Value:    time.Time{},
					},
					DateRange: utils.Optional[diary.DateRange]{
						HasValue: false,
						Value: diary.DateRange{
							Start: time.Time{},
							End:   time.Time{},
						},
					},
				},
				page:     1,
				pageSize: 10,
			},
			want: utils.PaginationResponse[diary.DiaryModel]{
				Items: []diary.DiaryModel{},
				Pagination: utils.Pagination{
					Page:     1,
					PageSize: 10,
					HasNext:  false,
					HasPrev:  false,
				},
			},
			wantErr: true,
			mock: func(args args) {
				mock.ExpectQuery(`SELECT id, name, description, start_time, end_time FROM activity_diary.*ORDER BY.*LIMIT.*OFFSET`).
					WithArgs(
						true, args.filter.ID.Value,
						true, args.filter.Name.Value,
						true, args.filter.Description.Value,
						true, args.filter.StartTime.Value,
						true, args.filter.EndTime.Value,
						true, args.filter.DateRange.Value.Start, args.filter.DateRange.Value.End,
						args.pageSize+1, utils.CalculateOffset(args.page, args.pageSize),
					).
					WillReturnError(errors.New("test error"))
			},
		},
		{
			name: "scan error",
			repo: repo,
			args: args{
				filter: diary.DiaryFilter{
					ID: utils.Optional[int]{
						HasValue: false,
						Value:    0,
					},
					Name: utils.Optional[string]{
						HasValue: false,
						Value:    "",
					},
					Description: utils.Optional[string]{
						HasValue: false,
						Value:    "",
					},
					StartTime: utils.Optional[time.Time]{
						HasValue: false,
						Value:    time.Time{},
					},
					EndTime: utils.Optional[time.Time]{
						HasValue: false,
						Value:    time.Time{},
					},
					DateRange: utils.Optional[diary.DateRange]{
						HasValue: false,
						Value: diary.DateRange{
							Start: time.Time{},
							End:   time.Time{},
						},
					},
				},
				page:     1,
				pageSize: 10,
			},
			want: utils.PaginationResponse[diary.DiaryModel]{
				Items: []diary.DiaryModel{},
				Pagination: utils.Pagination{
					Page:     1,
					PageSize: 10,
					HasNext:  false,
					HasPrev:  false,
				},
			},
			wantErr: true,
			mock: func(args args) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "start_time", "end_time"}).
					AddRow(1, "test", "test", "test", "test")
				mock.ExpectQuery(`SELECT id, name, description, start_time, end_time FROM activity_diary.*ORDER BY.*LIMIT.*OFFSET`).
					WithArgs(
						true, args.filter.ID.Value,
						true, args.filter.Name.Value,
						true, args.filter.Description.Value,
						true, args.filter.StartTime.Value,
						true, args.filter.EndTime.Value,
						true, args.filter.DateRange.Value.Start, args.filter.DateRange.Value.End,
						args.pageSize+1, utils.CalculateOffset(args.page, args.pageSize),
					).
					WillReturnRows(rows)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.args)
			got, err := tt.repo.GetDiary(tt.args.filter, tt.args.page, tt.args.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetDiary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repository.GetDiary() = %v, want %v", got, tt.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestRepository_GetDbContext(t *testing.T) {
	db := &sql.DB{}
	repo := &diary.Repository{DbContext: db}

	result := repo.GetDbContext()

	if result != db {
		t.Errorf("Expected %v, got %v", db, result)
	}
}
