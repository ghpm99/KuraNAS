package diary

import (
	"database/sql"
	"time"
)

type DiaryModel struct {
	ID          int
	Name        string
	Description string
	StartTime   time.Time
	EndTime     sql.NullTime
}

func (diaryDto *DiaryDto) ToModel() (DiaryModel, error) {
	diaryModel := DiaryModel{
		ID:          diaryDto.ID,
		Name:        diaryDto.Name,
		Description: diaryDto.Description,
		StartTime:   diaryDto.StartTime,
	}

	endTime, err := diaryDto.EndTime.ParseToNullTime()
	if err != nil {
		return diaryModel, err
	}

	diaryModel.EndTime = endTime

	return diaryModel, nil
}
