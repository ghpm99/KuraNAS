package diary

import (
	"nas-go/api/pkg/utils"
	"time"
)

type DiaryDto struct {
	ID          int
	Name        string
	Description string
	StartTime   time.Time
	EndTime     utils.Optional[time.Time]
}

func (diaryModel *DiaryModel) ToDto() (DiaryDto, error) {
	diaryDto := DiaryDto{
		ID:          diaryModel.ID,
		Name:        diaryModel.Name,
		Description: diaryModel.Description,
		StartTime:   diaryModel.StartTime,
	}
	if err := diaryDto.EndTime.ParseFromNullTime(diaryModel.EndTime); err != nil {
		return diaryDto, err
	}

	return diaryDto, nil
}

type DiaryFilter struct {
	ID          utils.Optional[int]
	Name        utils.Optional[string]
	Description utils.Optional[string]
	StartTime   utils.Optional[time.Time]
	EndTime     utils.Optional[time.Time]
}
