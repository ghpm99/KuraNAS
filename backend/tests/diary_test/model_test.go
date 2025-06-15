package diary_test

import (
	"database/sql"
	"nas-go/api/internal/api/v1/diary"
	"nas-go/api/pkg/utils"
	"reflect"
	"testing"
	"time"
)

func TestDiaryDto_ToModel(t *testing.T) {
	// Use um valor fixo de tempo para todos os testes
	fixedStart := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	fixedEnd := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	type fields struct {
		ID          int
		Name        string
		Description string
		StartTime   time.Time
		EndTime     utils.Optional[time.Time]
	}
	tests := []struct {
		name    string
		fields  fields
		want    diary.DiaryModel
		wantErr bool
	}{
		{
			name: "valid diary dto with end time",
			fields: fields{
				ID:          1,
				Name:        "Test Diary",
				Description: "Test Description",
				StartTime:   fixedStart,
				EndTime: utils.Optional[time.Time]{
					HasValue: true,
					Value:    fixedEnd,
				},
			},
			want: diary.DiaryModel{
				ID:          1,
				Name:        "Test Diary",
				Description: "Test Description",
				StartTime:   fixedStart,
				EndTime: sql.NullTime{
					Time:  fixedEnd,
					Valid: true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid diary dto without end time",
			fields: fields{
				ID:          1,
				Name:        "Test Diary",
				Description: "Test Description",
				StartTime:   fixedStart,
				EndTime: utils.Optional[time.Time]{
					HasValue: false,
				},
			},
			want: diary.DiaryModel{
				ID:          1,
				Name:        "Test Diary",
				Description: "Test Description",
				StartTime:   fixedStart,
				EndTime: sql.NullTime{
					Valid: false,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid diary dto with invalid end time",
			fields: fields{
				ID:          1,
				Name:        "Test Diary",
				Description: "Test Description",
				StartTime:   fixedStart,
				EndTime: utils.Optional[time.Time]{
					Value:    time.Time{},
					HasValue: true,
				},
			},
			want: diary.DiaryModel{
				ID:          1,
				Name:        "Test Diary",
				Description: "Test Description",
				StartTime:   fixedStart,
				EndTime: sql.NullTime{
					Time:  time.Time{},
					Valid: true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diaryDto := &diary.DiaryDto{
				ID:          tt.fields.ID,
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				StartTime:   tt.fields.StartTime,
				EndTime:     tt.fields.EndTime,
			}
			got, err := diaryDto.ToModel()
			if (err != nil) != tt.wantErr {
				t.Errorf("DiaryDto.ToModel() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DiaryDto.ToModel() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
